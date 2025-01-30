package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"viacortex/internal/db"

	"github.com/go-chi/chi/v5"
)

// getBackendServers returns all backend servers for a domain
func (h *Handlers) getBackendServers(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    domainID := chi.URLParam(r, "id")
	domainIDInt, err := strconv.Atoi(domainID)
	if err != nil {
		log.Printf("Invalid domain ID: %v", err)
		http.Error(w, "Invalid domain ID", http.StatusBadRequest)
		return
	}
    rows, err := h.db.Query(ctx, `
        SELECT id, scheme, ip, port, weight, is_active, last_health_check, health_status,
               created_at, updated_at
        FROM backend_servers 
        WHERE domain_id = $1
        ORDER BY created_at DESC
    `, domainIDInt)
	

    
    if err != nil {
        log.Printf("Error fetching backend servers: %v", err)
        http.Error(w, "Failed to fetch backend servers", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    servers := []db.BackendServer{}
    for rows.Next() {
        var server db.BackendServer
        err := rows.Scan(
            &server.ID, &server.Scheme, &server.IP, &server.Port,
			&server.Weight, &server.IsActive,
            &server.LastHealthCheck, &server.HealthStatus,
            &server.CreatedAt, &server.UpdatedAt,
        )
        if err != nil {
            log.Printf("Error scanning backend server: %v", err)
            continue
        }
        servers = append(servers, server)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(servers)
}

// addBackendServer adds a new backend server to a domain
func (h *Handlers) addBackendServer(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    domainID := chi.URLParam(r, "id")

    var server db.BackendServer
    if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate server Scheme, IP, Port and weight
    if server.Scheme == "" || server.IP.String() == "" || server.Port == 0 {
		http.Error(w, "Invalid server details", http.StatusBadRequest)
		return
	}
    if server.Weight < 1 {
        server.Weight = 1 // Set default weight if invalid
    }

    var serverID int64
    err := h.db.QueryRow(ctx, `
		INSERT INTO backend_servers (domain_id, scheme, ip, port, weight, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, domainID, server.Scheme, server.IP, server.Port, server.Weight, server.IsActive).Scan(&serverID)


    if err != nil {
        log.Printf("Error creating backend server: %v", err)
        http.Error(w, "Failed to create backend server", http.StatusInternalServerError)
        return
    }

    // Record audit log
    userID := getUserIDFromContext(ctx)
    if err := h.recordAudit(ctx, userID, "create", "backend_server", serverID, server); err != nil {
        log.Printf("Error recording audit: %v", err)
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "id": serverID,
        "message": "Backend server created successfully",
    })
}

// updateBackendServer updates an existing backend server
func (h *Handlers) updateBackendServer(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    serverID := chi.URLParam(r, "serverID")

    var server db.BackendServer
    if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate server scheme, IP, port and weight
	if server.Scheme == "" || server.IP.String() == "" || server.Port == 0 {
		http.Error(w, "Invalid server details", http.StatusBadRequest)
		return
	}
    if server.Weight < 1 {
        server.Weight = 1 // Set default weight if invalid
    }

    // Get old values for audit log
    var oldServer db.BackendServer
    err := h.db.QueryRow(ctx, `
        SELECT scheme, ip, port, weight, is_active, health_status
		FROM backend_servers WHERE id = $1
	`, serverID).Scan(&oldServer.Scheme, &oldServer.IP, &oldServer.Port, &oldServer.Weight, &oldServer.IsActive, &oldServer.HealthStatus)

    if err != nil {
        log.Printf("Error fetching backend server: %v", err)
        http.Error(w, "Backend server not found", http.StatusNotFound)
        return
    }

    result, err := h.db.Exec(ctx, `
        UPDATE backend_servers 
        SET scheme = $1, ip = $2, port = $3, weight = $4, is_active = $5
		WHERE id = $6
	`, server.Scheme, server.IP, server.Port, server.Weight, server.IsActive, serverID)
    if err != nil {
        log.Printf("Error updating backend server: %v", err)
        http.Error(w, "Failed to update backend server", http.StatusInternalServerError)
        return
    }

    if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
        http.Error(w, "Backend server not found", http.StatusNotFound)
        return
    }

    // Record audit log
    userID := getUserIDFromContext(ctx)
    changes := map[string]interface{}{
        "old": oldServer,
        "new": server,
    }
    if err := h.recordAudit(ctx, userID, "update", "backend_server", 
        mustParseInt64(serverID), changes); err != nil {
        log.Printf("Error recording audit: %v", err)
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Backend server updated successfully",
    })
}

// deleteBackendServer deletes a backend server
func (h *Handlers) deleteBackendServer(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    serverID := chi.URLParam(r, "serverID")

    // Get server details for audit log before deletion
    var oldServer db.BackendServer
    err := h.db.QueryRow(ctx, `
        SELECT scheme, ip, port, weight, is_active, health_status
		FROM backend_servers WHERE id = $1
	`, serverID).Scan(&oldServer.Scheme, &oldServer.IP, &oldServer.Port, &oldServer.Weight, &oldServer.IsActive, &oldServer.HealthStatus)
    if err != nil {
        log.Printf("Error fetching backend server: %v", err)
        http.Error(w, "Backend server not found", http.StatusNotFound)
        return
    }

    result, err := h.db.Exec(ctx, "DELETE FROM backend_servers WHERE id = $1", serverID)
    if err != nil {
        log.Printf("Error deleting backend server: %v", err)
        http.Error(w, "Failed to delete backend server", http.StatusInternalServerError)
        return
    }

    if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
        http.Error(w, "Backend server not found", http.StatusNotFound)
        return
    }

    // Record audit log
    userID := getUserIDFromContext(ctx)
    if err := h.recordAudit(ctx, userID, "delete", "backend_server", 
        mustParseInt64(serverID), oldServer); err != nil {
        log.Printf("Error recording audit: %v", err)
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Backend server deleted successfully",
    })
}

// Helper function to parse int64 ID values
func mustParseInt64(s string) int64 {
    id, err := strconv.ParseInt(s, 10, 64)
    if err != nil {
        return 0
    }
    return id
}