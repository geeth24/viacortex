package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// getAuditLogs returns all audit logs with filtering options
func (h *Handlers) getAuditLogs(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse query parameters
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit == 0 {
        limit = 100 // Default limit
    }
    
    entityType := r.URL.Query().Get("entity_type")
    action := r.URL.Query().Get("action")
    userID := r.URL.Query().Get("user_id")
    
    // Build query with filters
    query := `
        SELECT 
            al.id, al.user_id, u.email as user_email,
            al.action, al.entity_type, al.entity_id,
            al.changes, al.timestamp
        FROM audit_logs al
        LEFT JOIN users u ON al.user_id = u.id
        WHERE 1=1
    `
    args := []interface{}{}
    argCount := 1

    if entityType != "" {
        query += ` AND al.entity_type = $` + strconv.Itoa(argCount)
        args = append(args, entityType)
        argCount++
    }
    
    if action != "" {
        query += ` AND al.action = $` + strconv.Itoa(argCount)
        args = append(args, action)
        argCount++
    }
    
    if userID != "" {
        query += ` AND al.user_id = $` + strconv.Itoa(argCount)
        args = append(args, userID)
        argCount++
    }
    
    query += ` ORDER BY al.timestamp DESC LIMIT $` + strconv.Itoa(argCount)
    args = append(args, limit)

    rows, err := h.db.Query(ctx, query, args...)
    if err != nil {
        log.Printf("Error fetching audit logs: %v", err)
        http.Error(w, "Failed to fetch audit logs", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    logs := []map[string]interface{}{}
    for rows.Next() {
        var l struct {
            ID          int64           `json:"id"`
            UserID      int64           `json:"user_id"`
            UserEmail   string          `json:"user_email"`
            Action      string          `json:"action"`
            EntityType  string          `json:"entity_type"`
            EntityID    int64           `json:"entity_id"`
            Changes     json.RawMessage `json:"changes"`
            Timestamp   time.Time       `json:"timestamp"`
        }
        
        err := rows.Scan(
            &l.ID, &l.UserID, &l.UserEmail,
            &l.Action, &l.EntityType, &l.EntityID,
            &l.Changes, &l.Timestamp,
        )
        if err != nil {
            log.Printf("Error scanning audit log: %v", err)
            continue
        }
        
        logs = append(logs, map[string]interface{}{
            "id":           l.ID,
            "user_id":      l.UserID,
            "user_email":   l.UserEmail,
            "action":       l.Action,
            "entity_type":  l.EntityType,
            "entity_id":    l.EntityID,
            "changes":      l.Changes,
            "timestamp":    l.Timestamp,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(logs)
}

// getEntityAuditLogs returns audit logs for a specific entity
func (h *Handlers) getEntityAuditLogs(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    entityType := chi.URLParam(r, "entityType")
    entityID := chi.URLParam(r, "entityID")
    
    // Parse query parameters
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit == 0 {
        limit = 100
    }
    
    rows, err := h.db.Query(ctx, `
        SELECT 
            al.id, al.user_id, u.email as user_email,
            al.action, al.changes, al.timestamp
        FROM audit_logs al
        LEFT JOIN users u ON al.user_id = u.id
        WHERE al.entity_type = $1 AND al.entity_id = $2
        ORDER BY al.timestamp DESC
        LIMIT $3
    `, entityType, entityID, limit)
    
    if err != nil {
        log.Printf("Error fetching entity audit logs: %v", err)
        http.Error(w, "Failed to fetch audit logs", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    logs := []map[string]interface{}{}
    for rows.Next() {
        var l struct {
            ID          int64           `json:"id"`
            UserID      int64           `json:"user_id"`
            UserEmail   string          `json:"user_email"`
            Action      string          `json:"action"`
            Changes     json.RawMessage `json:"changes"`
            Timestamp   time.Time       `json:"timestamp"`
        }
        
        err := rows.Scan(
            &l.ID, &l.UserID, &l.UserEmail,
            &l.Action, &l.Changes, &l.Timestamp,
        )
        if err != nil {
            log.Printf("Error scanning entity audit log: %v", err)
            continue
        }
        
        logs = append(logs, map[string]interface{}{
            "id":           l.ID,
            "user_id":      l.UserID,
            "user_email":   l.UserEmail,
            "action":       l.Action,
            "changes":      l.Changes,
            "timestamp":    l.Timestamp,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(logs)
}

// Helper function to record an audit log entry
func (h *Handlers) recordAudit(ctx context.Context, userID int64, action, entityType string, entityID int64, changes interface{}) error {
    changesJSON, err := json.Marshal(changes)
    if err != nil {
        return err
    }

    _, err = h.db.Exec(ctx, `
        INSERT INTO audit_logs (user_id, action, entity_type, entity_id, changes)
        VALUES ($1, $2, $3, $4, $5)
    `, userID, action, entityType, entityID, changesJSON)

    return err
}