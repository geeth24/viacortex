package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"viacortex/internal/db"

	"github.com/go-chi/chi/v5"
)

// getDomains returns all domains with their associated backend servers
func (h *Handlers) getDomains(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    domains := []db.Domain{}
    rows, err := h.db.Query(ctx, `
        SELECT 
            d.id, d.name, d.target_url, d.ssl_enabled, 
            d.health_check_enabled, d.health_check_interval,
            d.custom_error_pages, d.created_at, d.updated_at
        FROM domains d
        ORDER BY d.name
    `)
    if err != nil {
        log.Printf("Error fetching domains: %v", err)
        http.Error(w, "Failed to fetch domains", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    for rows.Next() {
        var d db.Domain
        err := rows.Scan(
            &d.ID, &d.Name, &d.TargetURL, &d.SSLEnabled,
            &d.HealthCheckEnabled, &d.HealthCheckInterval,
            &d.CustomErrorPages, &d.CreatedAt, &d.UpdatedAt,
        )
        if err != nil {
            log.Printf("Error scanning domain: %v", err)
            http.Error(w, "Failed to scan domain", http.StatusInternalServerError)
            return
        }
        
        // Fetch backend servers for this domain
        backendRows, err := h.db.Query(ctx, `
            SELECT id, scheme, ip, port, weight, is_active, last_health_check, health_status
            FROM backend_servers 
            WHERE domain_id = $1
        `, d.ID)
        if err != nil {
            log.Printf("Error fetching backend servers: %v", err)
            continue
        }
        
        var backends []db.BackendServer
        for backendRows.Next() {
            var b db.BackendServer
            err := backendRows.Scan(
                &b.ID, &b.Scheme, &b.IP, &b.Port,
				&b.Weight, &b.IsActive,
                &b.LastHealthCheck, &b.HealthStatus,
            )
            if err != nil {
                log.Printf("Error scanning backend server: %v", err)
                continue
            }
            backends = append(backends, b)
        }
        backendRows.Close()
        
        d.BackendServers = backends
        domains = append(domains, d)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(domains)
}

// createDomain creates a new domain with optional backend servers
func (h *Handlers) createDomain(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    var req struct {
        Domain        db.Domain         `json:"domain"`
        BackendServers []db.BackendServer `json:"backend_servers"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Start transaction
    tx, err := h.db.Begin(ctx)
    if err != nil {
        log.Printf("Error starting transaction: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback(ctx)

    // Insert domain
    var domainID int64
    err = tx.QueryRow(ctx, `
        INSERT INTO domains (
            name, target_url, ssl_enabled, health_check_enabled,
            health_check_interval, custom_error_pages
        ) VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `, req.Domain.Name, req.Domain.TargetURL, req.Domain.SSLEnabled,
       req.Domain.HealthCheckEnabled, req.Domain.HealthCheckInterval,
       req.Domain.CustomErrorPages).Scan(&domainID)

    if err != nil {
        log.Printf("Error creating domain: %v", err)
        http.Error(w, "Failed to create domain", http.StatusInternalServerError)
        return
    }

    // Insert backend servers if provided
    for _, backend := range req.BackendServers {
        _, err := tx.Exec(ctx, `
            INSERT INTO backend_servers (
                domain_id, scheme, ip, port, weight, is_active, health_status
				) VALUES ($1, $2, $3::inet, $4, $5, $6, $7)
		`, domainID, backend.Scheme, backend.IP.String(), backend.Port, backend.Weight, backend.IsActive, "healthy")

        if err != nil {
            log.Printf("Error creating backend server: %v", err)
            http.Error(w, "Failed to create backend servers", http.StatusInternalServerError)
            return
        }
    }

    if err := tx.Commit(ctx); err != nil {
        log.Printf("Error committing transaction: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "id": domainID,
        "message": "Domain created successfully",
    })
}

// updateDomain updates an existing domain and its backend servers
func (h *Handlers) updateDomain(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    domainID := chi.URLParam(r, "id")
    
    var req struct {
        Domain        db.Domain         `json:"domain"`
        BackendServers []db.BackendServer `json:"backend_servers"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    tx, err := h.db.Begin(ctx)
    if err != nil {
        log.Printf("Error starting transaction: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback(ctx)

    // Update domain
    _, err = tx.Exec(ctx, `
        UPDATE domains SET
            name = $1,
            target_url = $2,
            ssl_enabled = $3,
            health_check_enabled = $4,
            health_check_interval = $5,
            custom_error_pages = $6,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $7
    `, req.Domain.Name, req.Domain.TargetURL, req.Domain.SSLEnabled,
       req.Domain.HealthCheckEnabled, req.Domain.HealthCheckInterval,
       req.Domain.CustomErrorPages, domainID)

    if err != nil {
        log.Printf("Error updating domain: %v", err)
        http.Error(w, "Failed to update domain", http.StatusInternalServerError)
        return
    }

    // Delete existing backend servers
    _, err = tx.Exec(ctx, "DELETE FROM backend_servers WHERE domain_id = $1", domainID)
    if err != nil {
        log.Printf("Error deleting backend servers: %v", err)
        http.Error(w, "Failed to update backend servers", http.StatusInternalServerError)
        return
    }

    // Insert new backend servers
    for _, backend := range req.BackendServers {
        _, err := tx.Exec(ctx, `
            INSERT INTO backend_servers (
				domain_id, scheme, ip, port, weight, is_active, health_status
			) VALUES ($1, $2, $3::inet, $4, $5, $6, $7)
		`, domainID, backend.Scheme, backend.IP.String(), backend.Port, backend.Weight, backend.IsActive, "healthy")
		
        if err != nil {
            log.Printf("Error creating backend server: %v", err)
            http.Error(w, "Failed to create backend servers", http.StatusInternalServerError)
            return
        }
    }

    if err := tx.Commit(ctx); err != nil {
        log.Printf("Error committing transaction: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Domain updated successfully",
    })
}

// deleteDomain deletes a domain and all associated data
func (h *Handlers) deleteDomain(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    domainID := chi.URLParam(r, "id")

    id, err := strconv.ParseInt(domainID, 10, 64)
    if err != nil {
        http.Error(w, "Invalid domain ID", http.StatusBadRequest)
        return
    }

    // Start transaction
    tx, err := h.db.Begin(ctx)
    if err != nil {
        log.Printf("Error starting transaction: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback(ctx)

    // Delete associated data
    tables := []string{"backend_servers", "ip_rules", "rate_limits", "request_metrics", "request_logs"}
    for _, table := range tables {
        _, err := tx.Exec(ctx, "DELETE FROM "+table+" WHERE domain_id = $1", id)
        if err != nil {
            log.Printf("Error deleting from %s: %v", table, err)
            http.Error(w, "Server error", http.StatusInternalServerError)
            return
        }
    }

    // Delete the domain
    result, err := tx.Exec(ctx, "DELETE FROM domains WHERE id = $1", id)
    if err != nil {
        log.Printf("Error deleting domain: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
        http.Error(w, "Domain not found", http.StatusNotFound)
        return
    }

    if err := tx.Commit(ctx); err != nil {
        log.Printf("Error committing transaction: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Domain deleted successfully",
    })
}