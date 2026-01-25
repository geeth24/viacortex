package api

import (
	"encoding/json"
	"net/http"
	"time"

	custommiddleware "viacortex/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func SetupRoutes(r *chi.Mux, handlers *Handlers) {
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Refresh-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/api", func(api chi.Router) {
		api.Use(middleware.AllowContentType("application/json"))

		// Public routes
		api.Post("/register", handlers.handleRegister)
		api.Post("/login", handlers.handleLogin)
		api.Post("/refresh", handlers.handleRefresh)
		api.Get("/check-users", handlers.checkUsers)
		api.Get("/verify", handlers.verifyToken)
		api.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "ok",
				"version": "2.0.0",
			})
		})
		api.Options("/*", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Protected routes
		api.Group(func(r chi.Router) {
			r.Use(custommiddleware.AuthMiddleware)

			// Domains
			r.Get("/domains", handlers.getDomains)
			r.Post("/domains", handlers.createDomain)
			r.Get("/domains/{id}", handlers.getDomain)
			r.Put("/domains/{id}", handlers.updateDomain)
			r.Delete("/domains/{id}", handlers.deleteDomain)
			r.Patch("/domains/{id}/toggle", handlers.toggleDomain)

			// IP Rules
			r.Get("/domains/{id}/ip-rules", handlers.getIPRules)
			r.Post("/domains/{id}/ip-rules", handlers.addIPRule)
			r.Delete("/domains/{id}/ip-rules/{ruleID}", handlers.deleteIPRule)

			// Rate Limits
			r.Get("/domains/{id}/rate-limit", handlers.getRateLimit)
			r.Put("/domains/{id}/rate-limit", handlers.setRateLimit)
			r.Delete("/domains/{id}/rate-limit", handlers.deleteRateLimit)

			// Redirects
			r.Get("/domains/{id}/redirects", handlers.getRedirects)
			r.Post("/domains/{id}/redirects", handlers.addRedirect)
			r.Put("/domains/{id}/redirects/{redirectID}", handlers.updateRedirect)
			r.Delete("/domains/{id}/redirects/{redirectID}", handlers.deleteRedirect)

			// Access Logs
			r.Get("/logs", handlers.getAccessLogs)
			r.Get("/domains/{id}/logs", handlers.getDomainLogs)

			// Users
			r.Get("/users", handlers.getUsers)
			r.Post("/users", handlers.createUser)
			r.Get("/users/{id}", handlers.getUser)
			r.Put("/users/{id}", handlers.updateUser)
			r.Delete("/users/{id}", handlers.deleteUser)

			// Profile
			r.Get("/profile", handlers.getProfile)
			r.Put("/profile", handlers.updateProfile)

			// Audit Logs
			r.Get("/audit", handlers.getAuditLogs)
		})
	})
}
