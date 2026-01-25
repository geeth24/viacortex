package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type AccessLogEntry struct {
	ID             int64  `json:"id"`
	DomainID       int64  `json:"domain_id"`
	DomainName     string `json:"domain_name"`
	Timestamp      string `json:"timestamp"`
	ClientIP       string `json:"client_ip"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	StatusCode     int    `json:"status_code"`
	ResponseTimeMS int    `json:"response_time_ms"`
	BytesSent      int64  `json:"bytes_sent"`
	UserAgent      string `json:"user_agent"`
	Referer        string `json:"referer"`
}

func (h *Handlers) getAccessLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	rows, err := h.db.Query(ctx, `
		SELECT l.id, l.domain_id, d.name, l.timestamp, l.client_ip::text,
			l.method, l.path, l.status_code, l.response_time_ms, l.bytes_sent,
			COALESCE(l.user_agent, ''), COALESCE(l.referer, '')
		FROM access_logs l
		JOIN domains d ON d.id = l.domain_id
		ORDER BY l.timestamp DESC
		LIMIT $1
	`, limit)

	if err != nil {
		log.Printf("Error fetching access logs: %v", err)
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []AccessLogEntry
	for rows.Next() {
		var entry AccessLogEntry
		err := rows.Scan(
			&entry.ID, &entry.DomainID, &entry.DomainName, &entry.Timestamp,
			&entry.ClientIP, &entry.Method, &entry.Path, &entry.StatusCode,
			&entry.ResponseTimeMS, &entry.BytesSent, &entry.UserAgent, &entry.Referer,
		)
		if err != nil {
			log.Printf("Error scanning log: %v", err)
			continue
		}
		logs = append(logs, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (h *Handlers) getDomainLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "id")

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	rows, err := h.db.Query(ctx, `
		SELECT l.id, l.domain_id, d.name, l.timestamp, l.client_ip::text,
			l.method, l.path, l.status_code, l.response_time_ms, l.bytes_sent,
			COALESCE(l.user_agent, ''), COALESCE(l.referer, '')
		FROM access_logs l
		JOIN domains d ON d.id = l.domain_id
		WHERE l.domain_id = $1
		ORDER BY l.timestamp DESC
		LIMIT $2
	`, domainID, limit)

	if err != nil {
		log.Printf("Error fetching domain logs: %v", err)
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []AccessLogEntry
	for rows.Next() {
		var entry AccessLogEntry
		err := rows.Scan(
			&entry.ID, &entry.DomainID, &entry.DomainName, &entry.Timestamp,
			&entry.ClientIP, &entry.Method, &entry.Path, &entry.StatusCode,
			&entry.ResponseTimeMS, &entry.BytesSent, &entry.UserAgent, &entry.Referer,
		)
		if err != nil {
			log.Printf("Error scanning log: %v", err)
			continue
		}
		logs = append(logs, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (h *Handlers) getAuditLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	rows, err := h.db.Query(ctx, `
		SELECT a.id, a.user_id, COALESCE(u.email, 'unknown'), a.action,
			a.entity_type, a.entity_id, a.details, a.ip_address::text, a.timestamp
		FROM audit_logs a
		LEFT JOIN users u ON u.id = a.user_id
		ORDER BY a.timestamp DESC
		LIMIT $1
	`, limit)

	if err != nil {
		log.Printf("Error fetching audit logs: %v", err)
		http.Error(w, "Failed to fetch audit logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type AuditEntry struct {
		ID         int64       `json:"id"`
		UserID     *int64      `json:"user_id"`
		UserEmail  string      `json:"user_email"`
		Action     string      `json:"action"`
		EntityType string      `json:"entity_type"`
		EntityID   *int64      `json:"entity_id"`
		Details    interface{} `json:"details"`
		IPAddress  string      `json:"ip_address"`
		Timestamp  string      `json:"timestamp"`
	}

	var logs []AuditEntry
	for rows.Next() {
		var entry AuditEntry
		var details []byte
		var ipAddr *string
		err := rows.Scan(
			&entry.ID, &entry.UserID, &entry.UserEmail, &entry.Action,
			&entry.EntityType, &entry.EntityID, &details, &ipAddr, &entry.Timestamp,
		)
		if err != nil {
			log.Printf("Error scanning audit log: %v", err)
			continue
		}
		if ipAddr != nil {
			entry.IPAddress = *ipAddr
		}
		if len(details) > 0 {
			json.Unmarshal(details, &entry.Details)
		}
		logs = append(logs, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
