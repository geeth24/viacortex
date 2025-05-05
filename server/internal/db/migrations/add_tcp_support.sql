-- Add TCP support for Minecraft and other TCP services

-- Add tcp_metrics table to track TCP connections
CREATE TABLE IF NOT EXISTS tcp_metrics (
    id SERIAL PRIMARY KEY,
    domain_id INTEGER NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    connection_count INTEGER NOT NULL,
    avg_latency_ms FLOAT NOT NULL,
    p95_latency_ms FLOAT NOT NULL,
    p99_latency_ms FLOAT NOT NULL
);

-- Add index for efficient querying of TCP metrics by domain and time
CREATE INDEX IF NOT EXISTS idx_tcp_metrics_domain_time ON tcp_metrics(domain_id, timestamp);

-- Allow 'tcp' as a valid scheme in backend_servers
ALTER TABLE backend_servers DROP CONSTRAINT IF EXISTS valid_scheme;
ALTER TABLE backend_servers ADD CONSTRAINT valid_scheme CHECK (scheme IN ('http', 'https', 'tcp'));

-- Make sure health_status can be properly updated
COMMENT ON COLUMN backend_servers.health_status IS 'Health status of the backend server (healthy/unhealthy). Checked differently for HTTP/HTTPS vs TCP.'; 