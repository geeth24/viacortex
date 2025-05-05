-- Create certificates table
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
);

-- Add index on domain_id for faster queries
CREATE INDEX IF NOT EXISTS idx_certificates_domain_id ON certificates(domain_id);

-- Add index on domain_name for faster lookups
CREATE INDEX IF NOT EXISTS idx_certificates_domain_name ON certificates(domain_name);

-- Add index on status for filtering
CREATE INDEX IF NOT EXISTS idx_certificates_status ON certificates(status);

-- Add index on not_after for expiration queries
CREATE INDEX IF NOT EXISTS idx_certificates_not_after ON certificates(not_after); 