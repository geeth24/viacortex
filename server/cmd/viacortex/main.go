package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"viacortex/internal/api"
	"viacortex/internal/db"
	"viacortex/internal/healthcheck"
	"viacortex/internal/middleware"
	"viacortex/internal/proxy"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
    // Create a context that we'll cancel on shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Initialize DB connection
    dbpool, err := db.InitDB()
    if err != nil {
        log.Fatalf("Unable to connect to database: %v\n", err)
    }
    defer dbpool.Close()

    // Initialize proxy server
    proxyServer, err := proxy.NewProxyServer()
    if err != nil {
        log.Fatal(err)
    }
	if err := proxyServer.ConfigureCertmagic("geeth0924@gmail.com"); err != nil {
    log.Fatalf("Failed to configure certmagic: %v", err)
}
    proxyServer.Metrics().SetDB(dbpool)

    // Initialize and do first load of domains
    loader := proxy.NewLoader(dbpool, proxyServer)
	if err := loader.LoadAllDomains(); err != nil {
		log.Printf("Initial domain load error: %v", err)
	}
    // Start background domain loading
    go loader.Start(ctx)

	healthChecker := healthcheck.NewChecker(dbpool)
    healthChecker.Start(ctx)

    // Initialize admin router with middleware
    r := chi.NewRouter()

    // Basic middleware
    r.Use(chimiddleware.RequestID)
    r.Use(chimiddleware.RealIP)
    r.Use(chimiddleware.Logger)
    r.Use(chimiddleware.Recoverer)
    r.Use(chimiddleware.Timeout(60 * time.Second))

    // Security middleware
    r.Use(middleware.SecurityHeaders)
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:*", "https://*.viacortex.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Refresh-Token"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:          300,
    }))
    r.Use(chimiddleware.Throttle(1000))
    r.Use(chimiddleware.Compress(5))

    // Initialize handlers and routes
    handlers := api.NewHandlers(dbpool)
    api.SetupRoutes(r, handlers)

    // TLS configuration
    tlsConfig := &tls.Config{
        MinVersion:               tls.VersionTLS12,
        CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
        PreferServerCipherSuites: true,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
            tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
            tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        },
    }

    // Create admin server
    adminServer := &http.Server{
        Addr:         ":8080",
        Handler:      r,
        TLSConfig:    tlsConfig,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    // Create a WaitGroup to manage our servers
    var wg sync.WaitGroup
    wg.Add(2)

    // Start admin server (8080)
    go func() {
        defer wg.Done()
        log.Println("Admin server starting on port 8080")
        if err := adminServer.ListenAndServe(); err != http.ErrServerClosed {
            log.Printf("Admin server error: %v", err)
        }
    }()

    // Start proxy server (80/443)
    go func() {
        defer wg.Done()
        log.Println("Proxy server starting on ports 80 and 443")
        log.Println("TCP proxy for Minecraft should also be starting on port 25565")
        
        // Debug DNS resolution
        go func() {
            time.Sleep(5 * time.Second) // Wait for everything to start
            testDomains := []string{"mc.maxbrowser.win", "vc.maxbrowser.win"}
            for _, domain := range testDomains {
                ips, err := net.LookupIP(domain)
                if err != nil {
                    log.Printf("DNS lookup for %s failed: %v", domain, err)
                } else {
                    log.Printf("DNS lookup for %s succeeded: %v", domain, ips)
                }
            }
        }()
        
        if err := proxyServer.Run(80, 443); err != nil {
            log.Printf("Proxy server error: %v", err)
        }
    }()

    // Set up graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-stop
        log.Println("Shutting down servers...")
        
        // Cancel context to stop the loader
        cancel()

		// Stop health checker
		 healthChecker.Stop()
		 
        // Create shutdown context with timeout
        shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer shutdownCancel()

        // Shutdown admin server
        if err := adminServer.Shutdown(shutdownCtx); err != nil {
            log.Printf("Admin server shutdown error: %v", err)
        }

        // Signal WaitGroup that we're done
        wg.Done()
        wg.Done()
    }()

    // Wait for clean shutdown
    wg.Wait()
    log.Println("Servers shut down gracefully")
}