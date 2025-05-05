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
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            CONSTRAINT valid_scheme CHECK (scheme IN ('http', 'https', 'tcp'))
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
        CREATE TABLE IF NOT EXISTS tcp_metrics (
            id SERIAL PRIMARY KEY,
            domain_id INTEGER NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
            timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
            connection_count INTEGER DEFAULT 0,
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
        CREATE INDEX IF NOT EXISTS idx_request_metrics_domain_time ON request_metrics(domain_id, timestamp);
        `,
        `
        CREATE INDEX IF NOT EXISTS idx_tcp_metrics_domain_time ON tcp_metrics(domain_id, timestamp);
        `,
        `
        CREATE TABLE IF NOT EXISTS certificates (
            id SERIAL PRIMARY KEY,
            domain_id INTEGER NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
            domain_name VARCHAR(255) NOT NULL,
            issuer VARCHAR(255) NOT NULL,
            serial_number VARCHAR(255) NOT NULL,
            not_before TIMESTAMP WITH TIME ZONE NOT NULL,
            not_after TIMESTAMP WITH TIME ZONE NOT NULL,
            status VARCHAR(50) NOT NULL,
            last_renewal TIMESTAMP WITH TIME ZONE NOT NULL,
            next_renewal TIMESTAMP WITH TIME ZONE NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )`,
        `
        CREATE INDEX IF NOT EXISTS idx_certificates_domain_id ON certificates(domain_id);
        `,
        `
        CREATE INDEX IF NOT EXISTS idx_certificates_domain_name ON certificates(domain_name);
        `,
        `
        CREATE INDEX IF NOT EXISTS idx_certificates_status ON certificates(status);
        `,
        `
        CREATE INDEX IF NOT EXISTS idx_certificates_not_after ON certificates(not_after);
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
        "request_metrics", "request_logs", "users", "audit_logs", "certificates",
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

// GetCertificateByID retrieves a certificate by its ID
func GetCertificateByID(ctx context.Context, pool *pgxpool.Pool, id int64) (*Certificate, error) {
    query := `
        SELECT id, domain_id, domain_name, issuer, serial_number, not_before, not_after, 
               status, last_renewal, next_renewal, created_at, updated_at
        FROM certificates
        WHERE id = $1
    `
    
    var cert Certificate
    err := pool.QueryRow(ctx, query, id).Scan(
        &cert.ID, &cert.DomainID, &cert.DomainName, &cert.Issuer, &cert.SerialNumber,
        &cert.NotBefore, &cert.NotAfter, &cert.Status, &cert.LastRenewal, &cert.NextRenewal,
        &cert.CreatedAt, &cert.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }
    
    return &cert, nil
}

// GetCertificatesByDomainID retrieves all certificates for a specific domain
func GetCertificatesByDomainID(ctx context.Context, pool *pgxpool.Pool, domainID int64) ([]Certificate, error) {
    query := `
        SELECT id, domain_id, domain_name, issuer, serial_number, not_before, not_after, 
               status, last_renewal, next_renewal, created_at, updated_at
        FROM certificates
        WHERE domain_id = $1
        ORDER BY created_at DESC
    `
    
    rows, err := pool.Query(ctx, query, domainID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var certificates []Certificate
    for rows.Next() {
        var cert Certificate
        err := rows.Scan(
            &cert.ID, &cert.DomainID, &cert.DomainName, &cert.Issuer, &cert.SerialNumber,
            &cert.NotBefore, &cert.NotAfter, &cert.Status, &cert.LastRenewal, &cert.NextRenewal,
            &cert.CreatedAt, &cert.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        certificates = append(certificates, cert)
    }
    
    if err := rows.Err(); err != nil {
        return nil, err
    }
    
    return certificates, nil
}

// CreateCertificate inserts a new certificate record
func CreateCertificate(ctx context.Context, pool *pgxpool.Pool, cert *Certificate) (int64, error) {
    query := `
        INSERT INTO certificates (
            domain_id, domain_name, issuer, serial_number, not_before, not_after, 
            status, last_renewal, next_renewal
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id
    `
    
    var id int64
    err := pool.QueryRow(ctx, query, 
        cert.DomainID, cert.DomainName, cert.Issuer, cert.SerialNumber,
        cert.NotBefore, cert.NotAfter, cert.Status, cert.LastRenewal, cert.NextRenewal,
    ).Scan(&id)
    if err != nil {
        return 0, err
    }
    
    return id, nil
}

// UpdateCertificate updates an existing certificate record
func UpdateCertificate(ctx context.Context, pool *pgxpool.Pool, cert *Certificate) error {
    query := `
        UPDATE certificates
        SET domain_id = $1, domain_name = $2, issuer = $3, serial_number = $4,
            not_before = $5, not_after = $6, status = $7, last_renewal = $8, next_renewal = $9
        WHERE id = $10
    `
    
    _, err := pool.Exec(ctx, query,
        cert.DomainID, cert.DomainName, cert.Issuer, cert.SerialNumber,
        cert.NotBefore, cert.NotAfter, cert.Status, cert.LastRenewal, cert.NextRenewal,
        cert.ID,
    )
    
    return err
}

// DeleteCertificate removes a certificate by ID
func DeleteCertificate(ctx context.Context, pool *pgxpool.Pool, id int64) error {
    query := `DELETE FROM certificates WHERE id = $1`
    _, err := pool.Exec(ctx, query, id)
    return err
}

// GetAllCertificates retrieves all certificates in the system
func GetAllCertificates(ctx context.Context, pool *pgxpool.Pool) ([]Certificate, error) {
    query := `
        SELECT id, domain_id, domain_name, issuer, serial_number, not_before, not_after, 
               status, last_renewal, next_renewal, created_at, updated_at
        FROM certificates
        ORDER BY domain_id, not_after
    `
    
    rows, err := pool.Query(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var certificates []Certificate
    for rows.Next() {
        var cert Certificate
        err := rows.Scan(
            &cert.ID, &cert.DomainID, &cert.DomainName, &cert.Issuer, &cert.SerialNumber,
            &cert.NotBefore, &cert.NotAfter, &cert.Status, &cert.LastRenewal, &cert.NextRenewal,
            &cert.CreatedAt, &cert.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        certificates = append(certificates, cert)
    }
    
    if err := rows.Err(); err != nil {
        return nil, err
    }
    
    return certificates, nil
}

// GetExpiringCertificates retrieves certificates that will expire within the specified days
func GetExpiringCertificates(ctx context.Context, pool *pgxpool.Pool, days int) ([]Certificate, error) {
    query := `
        SELECT id, domain_id, domain_name, issuer, serial_number, not_before, not_after, 
               status, last_renewal, next_renewal, created_at, updated_at
        FROM certificates
        WHERE not_after < (CURRENT_TIMESTAMP + INTERVAL '1 day' * $1)
        ORDER BY not_after
    `
    
    rows, err := pool.Query(ctx, query, days)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var certificates []Certificate
    for rows.Next() {
        var cert Certificate
        err := rows.Scan(
            &cert.ID, &cert.DomainID, &cert.DomainName, &cert.Issuer, &cert.SerialNumber,
            &cert.NotBefore, &cert.NotAfter, &cert.Status, &cert.LastRenewal, &cert.NextRenewal,
            &cert.CreatedAt, &cert.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        certificates = append(certificates, cert)
    }
    
    if err := rows.Err(); err != nil {
        return nil, err
    }
    
    return certificates, nil
}
