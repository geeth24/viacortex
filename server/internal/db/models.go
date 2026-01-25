package db

import (
	"database/sql"
	"encoding/json"
	"net"
	"time"
)

type Domain struct {
	ID                 int64           `json:"id"`
	Name               string          `json:"name"`
	TargetURL          string          `json:"target_url"`
	SSLEnabled         bool            `json:"ssl_enabled"`
	ForceHTTPS         bool            `json:"force_https"`
	HTTP2Enabled       bool            `json:"http2_enabled"`
	HSTSEnabled        bool            `json:"hsts_enabled"`
	HSTSMaxAge         int             `json:"hsts_max_age"`
	WebsocketEnabled   bool            `json:"websocket_enabled"`
	CustomHeaders      json.RawMessage `json:"custom_headers"`
	BasicAuthEnabled   bool            `json:"basic_auth_enabled"`
	BasicAuthUsername  string          `json:"basic_auth_username,omitempty"`
	BasicAuthPassword  string          `json:"basic_auth_password,omitempty"`
	Enabled            bool            `json:"enabled"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type IPRule struct {
	ID          int64     `json:"id"`
	DomainID    *int64    `json:"domain_id,omitempty"`
	IPRange     string    `json:"ip_range"`
	RuleType    string    `json:"rule_type"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RateLimit struct {
	ID                int64     `json:"id"`
	DomainID          int64     `json:"domain_id"`
	RequestsPerSecond int       `json:"requests_per_second"`
	BurstSize         int       `json:"burst_size"`
	PerIP             bool      `json:"per_ip"`
	Enabled           bool      `json:"enabled"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Redirect struct {
	ID            int64     `json:"id"`
	DomainID      int64     `json:"domain_id"`
	SourcePath    string    `json:"source_path"`
	TargetURL     string    `json:"target_url"`
	RedirectType  int       `json:"redirect_type"`
	PreserveQuery bool      `json:"preserve_query"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AccessLog struct {
	ID             int64     `json:"id"`
	DomainID       int64     `json:"domain_id"`
	Timestamp      time.Time `json:"timestamp"`
	ClientIP       net.IP    `json:"client_ip"`
	Method         string    `json:"method"`
	Path           string    `json:"path"`
	StatusCode     int       `json:"status_code"`
	ResponseTimeMS int       `json:"response_time_ms"`
	BytesSent      int64     `json:"bytes_sent"`
	UserAgent      string    `json:"user_agent"`
	Referer        string    `json:"referer"`
}

type User struct {
	ID        int64        `json:"id"`
	Email     string       `json:"email"`
	Name      string       `json:"name,omitempty"`
	Password  string       `json:"-"`
	Role      string       `json:"role"`
	Active    bool         `json:"active"`
	LastLogin sql.NullTime `json:"last_login,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type AuditLog struct {
	ID         int64           `json:"id"`
	UserID     *int64          `json:"user_id,omitempty"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   *int64          `json:"entity_id,omitempty"`
	Details    json.RawMessage `json:"details"`
	IPAddress  string          `json:"ip_address"`
	Timestamp  time.Time       `json:"timestamp"`
}
