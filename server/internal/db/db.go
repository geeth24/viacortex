package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

func InitDB() (*pgxpool.Pool, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/viacortex?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 3600

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	if err := createSchema(pool); err != nil {
		return nil, err
	}

	return pool, nil
}

func createSchema(pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	ctx := context.Background()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ LANGUAGE 'plpgsql';
	`)
	if err != nil {
		log.Printf("Error creating update_updated_at function: %v", err)
		return err
	}

	tableQueries := []string{
		// Domains - simplified, just target URL (k8s service handles the rest)
		`CREATE TABLE IF NOT EXISTS domains (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			target_url VARCHAR(512) NOT NULL,
			ssl_enabled BOOLEAN DEFAULT true,
			force_https BOOLEAN DEFAULT true,
			http2_enabled BOOLEAN DEFAULT true,
			hsts_enabled BOOLEAN DEFAULT false,
			hsts_max_age INTEGER DEFAULT 31536000,
			websocket_enabled BOOLEAN DEFAULT true,
			custom_headers JSONB DEFAULT '{}',
			basic_auth_enabled BOOLEAN DEFAULT false,
			basic_auth_username VARCHAR(255),
			basic_auth_password VARCHAR(255),
			enabled BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		// IP Rules - whitelist/blacklist
		`CREATE TABLE IF NOT EXISTS ip_rules (
			id SERIAL PRIMARY KEY,
			domain_id INTEGER REFERENCES domains(id) ON DELETE CASCADE,
			ip_range CIDR NOT NULL,
			rule_type VARCHAR(20) NOT NULL CHECK (rule_type IN ('whitelist', 'blacklist')),
			description TEXT,
			enabled BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		// Rate Limits
		`CREATE TABLE IF NOT EXISTS rate_limits (
			id SERIAL PRIMARY KEY,
			domain_id INTEGER REFERENCES domains(id) ON DELETE CASCADE,
			requests_per_second INTEGER NOT NULL,
			burst_size INTEGER DEFAULT 10,
			per_ip BOOLEAN DEFAULT true,
			enabled BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(domain_id)
		)`,

		// Redirects
		`CREATE TABLE IF NOT EXISTS redirects (
			id SERIAL PRIMARY KEY,
			domain_id INTEGER REFERENCES domains(id) ON DELETE CASCADE,
			source_path VARCHAR(512) NOT NULL,
			target_url VARCHAR(512) NOT NULL,
			redirect_type INTEGER DEFAULT 301 CHECK (redirect_type IN (301, 302, 307, 308)),
			preserve_query BOOLEAN DEFAULT true,
			enabled BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		// Access Logs
		`CREATE TABLE IF NOT EXISTS access_logs (
			id SERIAL PRIMARY KEY,
			domain_id INTEGER REFERENCES domains(id) ON DELETE CASCADE,
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			client_ip INET NOT NULL,
			method VARCHAR(10) NOT NULL,
			path TEXT NOT NULL,
			status_code INTEGER NOT NULL,
			response_time_ms INTEGER,
			bytes_sent BIGINT DEFAULT 0,
			user_agent TEXT,
			referer TEXT
		)`,

		// Users
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255),
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) DEFAULT 'user' CHECK (role IN ('admin', 'user')),
			active BOOLEAN DEFAULT true,
			last_login TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		// Audit Logs
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
			action VARCHAR(255) NOT NULL,
			entity_type VARCHAR(50),
			entity_id INTEGER,
			details JSONB,
			ip_address INET,
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_access_logs_domain_time ON access_logs(domain_id, timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_access_logs_timestamp ON access_logs(timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_domains_name ON domains(name)`,
	}

	for _, query := range tableQueries {
		if _, err := tx.Exec(ctx, query); err != nil {
			log.Printf("Error executing query: %v\nQuery: %s", err, query)
			return err
		}
	}

	for _, table := range []string{"domains", "ip_rules", "rate_limits", "redirects", "users"} {
		triggerName := fmt.Sprintf("update_%s_updated_at", table)
		query := fmt.Sprintf(`
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = '%s') THEN
					CREATE TRIGGER %s
					BEFORE UPDATE ON %s
					FOR EACH ROW
					EXECUTE FUNCTION update_updated_at_column();
				END IF;
			END;
			$$;`, triggerName, triggerName, table)
		if _, err := tx.Exec(ctx, query); err != nil {
			log.Printf("Error ensuring trigger exists: %v", err)
			return err
		}
	}

	return tx.Commit(ctx)
}
