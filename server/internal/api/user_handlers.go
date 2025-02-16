package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"viacortex/internal/db"
	"viacortex/internal/middleware"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

// getUsers returns all users (admin only)
func (h *Handlers) getUsers(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    rows, err := h.db.Query(ctx, `
        SELECT id, email, role, active, last_login, created_at, updated_at
        FROM users
        ORDER BY email
    `)
    if err != nil {
        log.Printf("Error fetching users: %v", err)
        http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    users := []db.User{}
    for rows.Next() {
        var u db.User
        err := rows.Scan(
            &u.ID, &u.Email, &u.Role, &u.Active,
            &u.LastLogin, &u.CreatedAt, &u.UpdatedAt,
        )
        if err != nil {
            log.Printf("Error scanning user: %v", err)
            continue
        }
        users = append(users, u)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

// createUser creates a new user (admin only)
func (h *Handlers) createUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
        Role     string `json:"role"`
        Name     string `json:"name"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate role
    if !isValidRole(req.Role) {
        http.Error(w, "Invalid role", http.StatusBadRequest)
        return
    }

    // Check if email already exists
    var exists bool
    err := h.db.QueryRow(ctx, 
        "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)",
        req.Email,
    ).Scan(&exists)
    if err != nil {
        log.Printf("Error checking email existence: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    if exists {
        http.Error(w, "Email already exists", http.StatusConflict)
        return
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        log.Printf("Error hashing password: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    // Create user
    var userID int64
    err = h.db.QueryRow(ctx, `
        INSERT INTO users (email, password_hash, role, active, name)
        VALUES ($1, $2, $3, true, NULLIF($4, ''))
        RETURNING id
    `, req.Email, string(hashedPassword), req.Role, req.Name).Scan(&userID)

    if err != nil {
        log.Printf("Error creating user: %v", err)
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }

    // Add audit log
    _, err = h.db.Exec(ctx, `
        INSERT INTO audit_logs (user_id, action, entity_type, entity_id, changes)
        VALUES ($1, 'create', 'user', $2, $3)
    `, getUserIDFromContext(ctx), userID, json.RawMessage(`{"email": "`+req.Email+`", "role": "`+req.Role+`"}`))

    if err != nil {
        log.Printf("Error creating audit log: %v", err)
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "id": userID,
        "message": "User created successfully",
    })
}

// updateUser updates a user's details (admin only)
func (h *Handlers) updateUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := chi.URLParam(r, "id")
    
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password,omitempty"`
        Active   bool   `json:"active"`
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

    // Update basic info
    if req.Password != "" {
        // Update with new password
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
        if err != nil {
            log.Printf("Error hashing password: %v", err)
            http.Error(w, "Server error", http.StatusInternalServerError)
            return
        }

        if _, err = tx.Exec(ctx, `
            UPDATE users 
            SET email = $1, password_hash = $2, active = $3, updated_at = CURRENT_TIMESTAMP
            WHERE id = $4
        `, req.Email, string(hashedPassword), req.Active, userID); err != nil {
            log.Printf("Error updating user: %v", err)
            http.Error(w, "Failed to update user", http.StatusInternalServerError)
            return
        }
    } else {
        // Update without changing password
        _, err = tx.Exec(ctx, `
            UPDATE users 
            SET email = $1, active = $2, updated_at = CURRENT_TIMESTAMP
            WHERE id = $3
        `, req.Email, req.Active, userID)
    }

    if err != nil {
        log.Printf("Error updating user: %v", err)
        http.Error(w, "Failed to update user", http.StatusInternalServerError)
        return
    }

    // Add audit log
    changes := map[string]interface{}{
        "email":  req.Email,
        "active": req.Active,
    }
    if req.Password != "" {
        changes["password_changed"] = true
    }
    
    changesJSON, _ := json.Marshal(changes)
    _, err = tx.Exec(ctx, `
        INSERT INTO audit_logs (user_id, action, entity_type, entity_id, changes)
        VALUES ($1, 'update', 'user', $2, $3)
    `, getUserIDFromContext(ctx), userID, changesJSON)

    if err != nil {
        log.Printf("Error creating audit log: %v", err)
    }

    if err := tx.Commit(ctx); err != nil {
        log.Printf("Error committing transaction: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "User updated successfully",
    })
}

// updateUserRole updates a user's role (admin only)
func (h *Handlers) updateUserRole(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := chi.URLParam(r, "id")
    
    var req struct {
        Role string `json:"role"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if !isValidRole(req.Role) {
        http.Error(w, "Invalid role", http.StatusBadRequest)
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

    // Update role
    _, err = tx.Exec(ctx, `
        UPDATE users 
        SET role = $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
    `, req.Role, userID)

    if err != nil {
        log.Printf("Error updating user role: %v", err)
        http.Error(w, "Failed to update user role", http.StatusInternalServerError)
        return
    }

    // Add audit log
    changes, _ := json.Marshal(map[string]string{"role": req.Role})
    _, err = tx.Exec(ctx, `
        INSERT INTO audit_logs (user_id, action, entity_type, entity_id, changes)
        VALUES ($1, 'update_role', 'user', $2, $3)
    `, getUserIDFromContext(ctx), userID, changes)

    if err != nil {
        log.Printf("Error creating audit log: %v", err)
    }

    if err := tx.Commit(ctx); err != nil {
        log.Printf("Error committing transaction: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "User role updated successfully",
    })
}

// deleteUser deletes a user (admin only)
func (h *Handlers) deleteUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := chi.URLParam(r, "id")

    // Start transaction
    tx, err := h.db.Begin(ctx)
    if err != nil {
        log.Printf("Error starting transaction: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback(ctx)

    // Get user details for audit log
    var email string
    err = tx.QueryRow(ctx, "SELECT email FROM users WHERE id = $1", userID).Scan(&email)
    if err != nil {
        log.Printf("Error fetching user details: %v", err)
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    // Delete user
    result, err := tx.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
    if err != nil {
        log.Printf("Error deleting user: %v", err)
        http.Error(w, "Failed to delete user", http.StatusInternalServerError)
        return
    }

    if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    // Add audit log
    changes, _ := json.Marshal(map[string]string{"email": email})
    _, err = tx.Exec(ctx, `
        INSERT INTO audit_logs (user_id, action, entity_type, entity_id, changes)
        VALUES ($1, 'delete', 'user', $2, $3)
    `, getUserIDFromContext(ctx), userID, changes)

    if err != nil {
        log.Printf("Error creating audit log: %v", err)
    }

    if err := tx.Commit(ctx); err != nil {
        log.Printf("Error committing transaction: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "User deleted successfully",
    })
}

// updateUserProfile updates a user's profile
func (h *Handlers) updateUserProfile(w http.ResponseWriter, r *http.Request) {
    log.Println("updateUserProfile")
    ctx := r.Context()
    
    // Get userID from context
    userID := getUserIDFromContext(ctx)
    if userID == 0 {
        http.Error(w, "Not authenticated", http.StatusUnauthorized)
        return
    }
    
    var req struct {
        Name string `json:"name"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("Error decoding request: %v", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate name
    req.Name = strings.TrimSpace(req.Name)
    if req.Name == "" {
        http.Error(w, "Name cannot be empty", http.StatusBadRequest)
        return
    }

    // Update user profile
    result, err := h.db.Exec(ctx, `
        UPDATE users 
        SET name = $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
    `, req.Name, userID)

    if err != nil {
        log.Printf("Error updating user profile: %v", err)
        http.Error(w, "Failed to update profile", http.StatusInternalServerError)
        return
    }

    // Check if user was found and updated
    rowsAffected := result.RowsAffected()
    if rowsAffected == 0 {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    var user db.User
    err = h.db.QueryRow(ctx, `
        SELECT id, email, name, role, active, last_login, created_at, updated_at
        FROM users WHERE id = $1
    `, userID).Scan(
        &user.ID, &user.Email, &user.Name, &user.Role,
        &user.Active, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
    )

    if err != nil {
        log.Printf("Error fetching updated user: %v", err)
        http.Error(w, "Failed to fetch updated profile", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Profile updated successfully",
        "user":    user,
    })
}

// Helper functions

func isValidRole(role string) bool {
    validRoles := map[string]bool{
        "admin":    true,
        "user":     true,
        "readonly": true,
    }
    return validRoles[role]
}

func getUserIDFromContext(ctx context.Context) int64 {
    return middleware.GetUserIDFromContext(ctx)
}