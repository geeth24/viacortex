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

    // Configure connection pool
    config.MaxConns = 10
    config.MinConns = 2
    config.MaxConnLifetime = 3600 // 1 hour

    pool, err := pgxpool.ConnectConfig(context.Background(), config)
    if err != nil {
        return nil, err
    }

    // Initialize schema
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

    // Create tables in a transaction
    tx, err := conn.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // Create updated_at function for triggers
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

    // Create tables
    tableQueries := []string{
        `
        CREATE TABLE IF NOT EXISTS domains (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL UNIQUE,
            target_url VARCHAR(255) NOT NULL,
            ssl_enabled BOOLEAN DEFAULT true,
            health_check_enabled BOOLEAN DEFAULT false,
            health_check_interval INTEGER DEFAULT 60,
            custom_error_pages JSONB,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )`,
        `
        CREATE TABLE IF NOT EXISTS backend_servers (
            id SERIAL PRIMARY KEY,
            domain_id INTEGER NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
            scheme VARCHAR(10) DEFAULT 'http',
			ip INET NOT NULL,
            port INTEGER NOT NULL,
            weight INTEGER DEFAULT 1,
            is_active BOOLEAN DEFAULT true,
            last_health_check TIMESTAMP,
            health_status VARCHAR(50),
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )`,
        `
        CREATE TABLE IF NOT EXISTS ip_rules (
            id SERIAL PRIMARY KEY,
            domain_id INTEGER NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
            ip_range CIDR NOT NULL,
            rule_type VARCHAR(50) NOT NULL,
            description TEXT,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )`,
        `
        CREATE TABLE IF NOT EXISTS rate_limits (
            id SERIAL PRIMARY KEY,
            domain_id INTEGER NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
            requests_per_second INTEGER NOT NULL,
            burst_size INTEGER DEFAULT 0,
            per_ip BOOLEAN DEFAULT true,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )`,
        `
        CREATE TABLE IF NOT EXISTS request_metrics (
            id SERIAL PRIMARY KEY,
            domain_id INTEGER NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
            timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
            request_count INTEGER DEFAULT 0,
            error_count INTEGER DEFAULT 0,
            avg_latency_ms FLOAT DEFAULT 0,
            p95_latency_ms FLOAT DEFAULT 0,
            p99_latency_ms FLOAT DEFAULT 0
        )`,
        `
        CREATE TABLE IF NOT EXISTS request_logs (
            id SERIAL PRIMARY KEY,
            domain_id INTEGER NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
            timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
            client_ip INET NOT NULL,
            method VARCHAR(10) NOT NULL,
            path TEXT NOT NULL,
            status_code INTEGER NOT NULL,
            response_time_ms INTEGER,
            user_agent TEXT,
            referer TEXT
        )`,
        `
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            email VARCHAR(255) NOT NULL UNIQUE,
            name VARCHAR(255),
            password_hash VARCHAR(255) NOT NULL,
            role VARCHAR(50) DEFAULT 'user',
            active BOOLEAN DEFAULT true,
            last_login TIMESTAMP WITH TIME ZONE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )`,
        `
        CREATE TABLE IF NOT EXISTS audit_logs (
            id SERIAL PRIMARY KEY,
            user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE SET NULL,
            action VARCHAR(255) NOT NULL,
            entity_type VARCHAR(50),
            entity_id INTEGER,
            changes JSONB,
            timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )`,
        `
        `,
    }

    for _, query := range tableQueries {
        if _, err := tx.Exec(ctx, query); err != nil {
            log.Printf("Error executing query: %v\nQuery: %s", err, query)
            return err
        }
    }

    // Create triggers for updated_at
    for _, table := range []string{
        "domains", "backend_servers", "ip_rules", "rate_limits",
        "request_metrics", "request_logs", "users", "audit_logs",
    } {
        triggerName := fmt.Sprintf("update_%s_updated_at", table)
        query := fmt.Sprintf(`
            DO $$
            BEGIN
                IF NOT EXISTS (
                    SELECT 1
                    FROM pg_trigger
                    WHERE tgname = '%s'
                ) THEN
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

    // Commit transaction
    return tx.Commit(ctx)
}
