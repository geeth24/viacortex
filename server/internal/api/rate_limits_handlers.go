package api

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/go-chi/chi/v5"
    "viacortex/internal/db"
)

// getRateLimits returns all rate limits for a domain
func (h *Handlers) getRateLimits(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    domainID := chi.URLParam(r, "id")

    rows, err := h.db.Query(ctx, `
        SELECT id, requests_per_second, burst_size, per_ip, created_at, updated_at
        FROM rate_limits 
        WHERE domain_id = $1
        ORDER BY created_at DESC
    `, domainID)
    
    if err != nil {
        log.Printf("Error fetching rate limits: %v", err)
        http.Error(w, "Failed to fetch rate limits", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    limits := []db.RateLimit{}
    for rows.Next() {
        var limit db.RateLimit
        err := rows.Scan(
            &limit.ID, &limit.RequestsPerSecond, &limit.BurstSize,
            &limit.PerIP, &limit.CreatedAt, &limit.UpdatedAt,
        )
        if err != nil {
            log.Printf("Error scanning rate limit: %v", err)
            continue
        }
        limits = append(limits, limit)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(limits)
}

// addRateLimit adds a new rate limit to a domain
func (h *Handlers) addRateLimit(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    domainID := chi.URLParam(r, "id")

    var limit db.RateLimit
    if err := json.NewDecoder(r.Body).Decode(&limit); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate rate limit values
    if limit.RequestsPerSecond <= 0 || limit.BurstSize <= 0 {
        http.Error(w, "Invalid rate limit values", http.StatusBadRequest)
        return
    }

    var limitID int64
    err := h.db.QueryRow(ctx, `
        INSERT INTO rate_limits (domain_id, requests_per_second, burst_size, per_ip)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `, domainID, limit.RequestsPerSecond, limit.BurstSize, limit.PerIP).Scan(&limitID)

    if err != nil {
        log.Printf("Error creating rate limit: %v", err)
        http.Error(w, "Failed to create rate limit", http.StatusInternalServerError)
        return
    }

    // Record audit log
    userID := getUserIDFromContext(ctx)
    if err := h.recordAudit(ctx, userID, "create", "rate_limit", limitID, limit); err != nil {
        log.Printf("Error recording audit: %v", err)
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "id": limitID,
        "message": "Rate limit created successfully",
    })
}

// updateRateLimit updates an existing rate limit
func (h *Handlers) updateRateLimit(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    limitID := chi.URLParam(r, "limitID")

    var limit db.RateLimit
    if err := json.NewDecoder(r.Body).Decode(&limit); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate rate limit values
    if limit.RequestsPerSecond <= 0 || limit.BurstSize <= 0 {
        http.Error(w, "Invalid rate limit values", http.StatusBadRequest)
        return
    }

    // Get old values for audit log
    var oldLimit db.RateLimit
    err := h.db.QueryRow(ctx, `
        SELECT requests_per_second, burst_size, per_ip 
        FROM rate_limits WHERE id = $1
    `, limitID).Scan(&oldLimit.RequestsPerSecond, &oldLimit.BurstSize, &oldLimit.PerIP)
    
    if err != nil {
        log.Printf("Error fetching rate limit: %v", err)
        http.Error(w, "Rate limit not found", http.StatusNotFound)
        return
    }

    result, err := h.db.Exec(ctx, `
        UPDATE rate_limits 
        SET requests_per_second = $1, burst_size = $2, per_ip = $3
        WHERE id = $4
    `, limit.RequestsPerSecond, limit.BurstSize, limit.PerIP, limitID)

    if err != nil {
        log.Printf("Error updating rate limit: %v", err)
        http.Error(w, "Failed to update rate limit", http.StatusInternalServerError)
        return
    }

    if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
        http.Error(w, "Rate limit not found", http.StatusNotFound)
        return
    }

    // Record audit log
    userID := getUserIDFromContext(ctx)
    changes := map[string]interface{}{
        "old": oldLimit,
        "new": limit,
    }
    if err := h.recordAudit(ctx, userID, "update", "rate_limit", 
        mustParseInt64(limitID), changes); err != nil {
        log.Printf("Error recording audit: %v", err)
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Rate limit updated successfully",
    })
}

// deleteRateLimit deletes a rate limit
func (h *Handlers) deleteRateLimit(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    limitID := chi.URLParam(r, "limitID")

    // Get rate limit details for audit log before deletion
    var oldLimit db.RateLimit
    err := h.db.QueryRow(ctx, `
        SELECT requests_per_second, burst_size, per_ip 
        FROM rate_limits WHERE id = $1
    `, limitID).Scan(&oldLimit.RequestsPerSecond, &oldLimit.BurstSize, &oldLimit.PerIP)
    
    if err != nil {
        log.Printf("Error fetching rate limit: %v", err)
        http.Error(w, "Rate limit not found", http.StatusNotFound)
        return
    }

    result, err := h.db.Exec(ctx, "DELETE FROM rate_limits WHERE id = $1", limitID)
    if err != nil {
        log.Printf("Error deleting rate limit: %v", err)
        http.Error(w, "Failed to delete rate limit", http.StatusInternalServerError)
        return
    }

    if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
        http.Error(w, "Rate limit not found", http.StatusNotFound)
        return
    }

    // Record audit log
    userID := getUserIDFromContext(ctx)
    if err := h.recordAudit(ctx, userID, "delete", "rate_limit", 
        mustParseInt64(limitID), oldLimit); err != nil {
        log.Printf("Error recording audit: %v", err)
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Rate limit deleted successfully",
    })
}