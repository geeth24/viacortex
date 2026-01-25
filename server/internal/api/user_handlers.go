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

func (h *Handlers) getUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows, err := h.db.Query(ctx, `
		SELECT id, email, name, role, active, last_login, created_at, updated_at
		FROM users
		ORDER BY email
	`)
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []db.User
	for rows.Next() {
		var u db.User
		var name *string
		err := rows.Scan(
			&u.ID, &u.Email, &name, &u.Role, &u.Active,
			&u.LastLogin, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			continue
		}
		if name != nil {
			u.Name = *name
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *Handlers) getUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	var u db.User
	var name *string
	err := h.db.QueryRow(ctx, `
		SELECT id, email, name, role, active, last_login, created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&u.ID, &u.Email, &name, &u.Role, &u.Active,
		&u.LastLogin, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if name != nil {
		u.Name = *name
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func (h *Handlers) createUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = "user"
	}

	if req.Role != "admin" && req.Role != "user" {
		http.Error(w, "Invalid role. Must be 'admin' or 'user'", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	var user db.User
	err = h.db.QueryRow(ctx, `
		INSERT INTO users (email, name, password_hash, role, active)
		VALUES ($1, $2, $3, $4, true)
		RETURNING id, created_at, updated_at
	`, req.Email, req.Name, string(hashedPassword), req.Role).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Failed to create user. Email may already exist.", http.StatusBadRequest)
		return
	}

	user.Email = req.Email
	user.Name = req.Name
	user.Role = req.Role
	user.Active = true

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *Handlers) updateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password,omitempty"`
		Role     string `json:"role"`
		Active   bool   `json:"active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var err error
	if req.Password != "" {
		hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			log.Printf("Error hashing password: %v", hashErr)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		_, err = h.db.Exec(ctx, `
			UPDATE users SET
				email = $1, name = $2, password_hash = $3, role = $4, active = $5,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $6
		`, req.Email, req.Name, string(hashedPassword), req.Role, req.Active, userID)
	} else {
		_, err = h.db.Exec(ctx, `
			UPDATE users SET
				email = $1, name = $2, role = $3, active = $4,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $5
		`, req.Email, req.Name, req.Role, req.Active, userID)
	}

	if err != nil {
		log.Printf("Error updating user: %v", err)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User updated successfully"})
}

func (h *Handlers) deleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	result, err := h.db.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}

func (h *Handlers) getProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := getUserIDFromContext(ctx)

	var u db.User
	var name *string
	err := h.db.QueryRow(ctx, `
		SELECT id, email, name, role, active, last_login, created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&u.ID, &u.Email, &name, &u.Role, &u.Active,
		&u.LastLogin, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if name != nil {
		u.Name = *name
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func (h *Handlers) updateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := getUserIDFromContext(ctx)

	var req struct {
		Name     string `json:"name"`
		Password string `json:"password,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)

	var err error
	if req.Password != "" {
		hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			log.Printf("Error hashing password: %v", hashErr)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		_, err = h.db.Exec(ctx, `
			UPDATE users SET name = $1, password_hash = $2, updated_at = CURRENT_TIMESTAMP
			WHERE id = $3
		`, req.Name, string(hashedPassword), userID)
	} else {
		_, err = h.db.Exec(ctx, `
			UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
		`, req.Name, userID)
	}

	if err != nil {
		log.Printf("Error updating profile: %v", err)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

func getUserIDFromContext(ctx context.Context) int64 {
	return middleware.GetUserIDFromContext(ctx)
}
