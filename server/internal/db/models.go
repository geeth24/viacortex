package db

import (
	"database/sql"
	"encoding/json"
	"net"
	"time"
)

type Domain struct {
    ID                  int64           `json:"id" db:"id"`
    Name               string          `json:"name" db:"name"`
    TargetURL          string          `json:"target_url" db:"target_url"`
    SSLEnabled         bool            `json:"ssl_enabled" db:"ssl_enabled"`
    HealthCheckEnabled bool            `json:"health_check_enabled" db:"health_check_enabled"`
    HealthCheckInterval int            `json:"health_check_interval" db:"health_check_interval"`
    CustomErrorPages   json.RawMessage `json:"custom_error_pages" db:"custom_error_pages"`
    CreatedAt          time.Time       `json:"created_at" db:"created_at"`
    UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`
	BackendServers     []BackendServer `json:"backend_servers,omitempty"`
}

type BackendServer struct {
    ID              int64     `json:"id" db:"id"`
    DomainID        int64     `json:"domain_id" db:"domain_id"`
	Scheme			string    `json:"scheme" db:"scheme"`
    IP			  net.IP    `json:"ip" db:"ip"`
    Port			int       `json:"port" db:"port"`
    Weight          int       `json:"weight" db:"weight"`
    IsActive        bool      `json:"is_active" db:"is_active"`
    LastHealthCheck *time.Time `json:"last_health_check,omitempty"`
    HealthStatus    *string    `json:"health_status,omitempty"`
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type IPRule struct {
    ID          int64     `json:"id" db:"id"`
    DomainID    int64     `json:"domain_id" db:"domain_id"`
    IPRange     net.IPNet `json:"ip_range" db:"ip_range"`
    RuleType    string    `json:"rule_type" db:"rule_type"`
    Description string    `json:"description" db:"description"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type RateLimit struct {
    ID                int64     `json:"id" db:"id"`
    DomainID         int64     `json:"domain_id" db:"domain_id"`
    RequestsPerSecond int       `json:"requests_per_second" db:"requests_per_second"`
    BurstSize        int       `json:"burst_size" db:"burst_size"`
    PerIP            bool      `json:"per_ip" db:"per_ip"`
    CreatedAt        time.Time `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type RequestMetrics struct {
    ID            int64     `json:"id" db:"id"`
    DomainID      int64     `json:"domain_id" db:"domain_id"`
    Timestamp     time.Time `json:"timestamp" db:"timestamp"`
    RequestCount  int       `json:"request_count" db:"request_count"`
    ErrorCount    int       `json:"error_count" db:"error_count"`
    AvgLatencyMS  float64   `json:"avg_latency_ms" db:"avg_latency_ms"`
    P95LatencyMS float64   `json:"p95_latency_ms" db:"p95_latency_ms"`
    P99LatencyMS float64   `json:"p99_latency_ms" db:"p99_latency_ms"`
}

type RequestLog struct {
    ID             int64     `json:"id" db:"id"`
    DomainID       int64     `json:"domain_id" db:"domain_id"`
    Timestamp      time.Time `json:"timestamp" db:"timestamp"`
    ClientIP       net.IP    `json:"client_ip" db:"client_ip"`
    Method         string    `json:"method" db:"method"`
    Path           string    `json:"path" db:"path"`
    StatusCode     int       `json:"status_code" db:"status_code"`
    ResponseTimeMS int       `json:"response_time_ms" db:"response_time_ms"`
    UserAgent      string    `json:"user_agent" db:"user_agent"`
    Referer        string    `json:"referer" db:"referer"`
}

type User struct {
    ID         int64          `json:"id" db:"id"`
    Email      string         `json:"email" db:"email"`
    Name       string         `json:"name,omitempty" db:"name"`
    Password   string         `json:"-" db:"password_hash"`
    Role       string         `json:"role" db:"role"`
    Active     bool          `json:"active" db:"active"`
    LastLogin  sql.NullTime  `json:"last_login,omitempty" db:"last_login"`
    CreatedAt  time.Time     `json:"created_at" db:"created_at"`
    UpdatedAt  time.Time     `json:"updated_at" db:"updated_at"`
}

type AuditLog struct {
    ID         int64           `json:"id" db:"id"`
    UserID     int64           `json:"user_id" db:"user_id"`
    Action     string          `json:"action" db:"action"`
    EntityType string          `json:"entity_type" db:"entity_type"`
    EntityID   int64           `json:"entity_id" db:"entity_id"`
    Changes    json.RawMessage `json:"changes" db:"changes"`
    Timestamp  time.Time       `json:"timestamp" db:"timestamp"`
}