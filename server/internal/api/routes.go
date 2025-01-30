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
    // Global middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(60 * time.Second))
    
    // Setup CORS
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"*"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Refresh-Token"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    }))

    r.Route("/api", func(apiRouter chi.Router) {
        // Middleware for all API routes
        apiRouter.Use(middleware.AllowContentType("application/json"))
        apiRouter.Use(middleware.SetHeader("Content-Type", "application/json"))

        // Public routes
        apiRouter.Group(func(r chi.Router) {
            r.Post("/register", handlers.handleRegister)
            r.Post("/login", handlers.handleLogin)
            r.Post("/refresh", handlers.handleRefresh)
        })

        // Status endpoint (public)
        apiRouter.Get("/status", func(w http.ResponseWriter, r *http.Request) {
            json.NewEncoder(w).Encode(map[string]string{
                "status":  "ok",
                "version": "1.0.0",
            })
        })

        // Protected routes
        apiRouter.Group(func(r chi.Router) {
            r.Use(custommiddleware.AuthMiddleware)
            
            // Domains
            r.Route("/domains", func(r chi.Router) {
                r.Get("/", handlers.getDomains)
                r.Post("/", handlers.createDomain)
                r.Route("/{id}", func(r chi.Router) {
                    r.Put("/", handlers.updateDomain)
                    r.Delete("/", handlers.deleteDomain)
                    
                    // Backend servers for a domain
                    r.Route("/backends", func(r chi.Router) {
                        r.Get("/", handlers.getBackendServers)
                        r.Post("/", handlers.addBackendServer)
                        r.Put("/{serverID}", handlers.updateBackendServer)
                        r.Delete("/{serverID}", handlers.deleteBackendServer)
                    })
                    
                    // IP rules for a domain
                    r.Route("/ip-rules", func(r chi.Router) {
                        r.Get("/", handlers.getIPRules)
                        r.Post("/", handlers.addIPRule)
                        r.Delete("/{ruleID}", handlers.deleteIPRule)
                    })
                    
                    // Rate limits for a domain
                    r.Route("/rate-limits", func(r chi.Router) {
                        r.Get("/", handlers.getRateLimits)
                        r.Post("/", handlers.addRateLimit)
                        r.Put("/{limitID}", handlers.updateRateLimit)
                        r.Delete("/{limitID}", handlers.deleteRateLimit)
                    })
                })
            })
            
            // Metrics and logs
            r.Route("/metrics", func(r chi.Router) {
                r.Get("/", handlers.getGlobalMetrics)
                r.Get("/{domainID}", handlers.getDomainMetrics)
            })
            
            r.Route("/logs", func(r chi.Router) {
                r.Get("/", handlers.getGlobalLogs)
                r.Get("/{domainID}", handlers.getDomainLogs)
            })
            
            // User management
            r.Route("/users", func(r chi.Router) {
                r.Get("/", handlers.getUsers)
                r.Post("/", handlers.createUser)
                r.Route("/{id}", func(r chi.Router) {
                    r.Put("/", handlers.updateUser)
                    r.Delete("/", handlers.deleteUser)
                    r.Put("/role", handlers.updateUserRole)
                })
            })

            // Audit logs
            r.Route("/audit", func(r chi.Router) {
                r.Get("/", handlers.getAuditLogs)
                r.Get("/{entityType}/{entityID}", handlers.getEntityAuditLogs)
            })
        })
    })
}