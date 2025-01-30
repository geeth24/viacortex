package proxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/caddyserver/certmagic"
	"golang.org/x/time/rate"
)

type ProxyServer struct {
    domains     sync.Map // map[string]*DomainConfig
    rateLimits  sync.Map // map[string]*rate.Limiter
    metrics     *MetricsCollector
    certManager *certmagic.Config
}

type DomainConfig struct {
    Domain             string
    Backends          []*BackendServer
    IPRules           []*IPRule
    RateLimit         *RateLimit
    SSLEnabled        bool
    HealthCheckEnabled bool
    currentBackend    int
    mu               sync.Mutex
}

type BackendServer struct {
    ID              int64
    Scheme          string
    IP              net.IP
    Port            int
    Weight          int
    IsActive        bool
    LastHealthCheck *time.Time
    HealthStatus    *string
}

type IPRule struct {
    ID          int64
    IPRange     net.IPNet
    RuleType    string    // "whitelist" or "blacklist"
    Description string
}

type RateLimit struct {
    ID                int64
    RequestsPerSecond int
    BurstSize        int
    PerIP            bool
}

func NewProxyServer() (*ProxyServer, error) {
    // Initialize certmagic with default config
    certConfig := certmagic.NewDefault()
    
    return &ProxyServer{
        certManager: certConfig,
        metrics:     NewMetricsCollector(),
    }, nil
}

func (p *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    domain := r.Host
    
    // Get domain config
    configVal, ok := p.domains.Load(domain)
    if !ok {
        http.Error(w, "Domain not found", http.StatusNotFound)
        return
    }
    config := configVal.(*DomainConfig)
    
    // Check IP rules
    if !p.checkIPRules(r, config) {
        http.Error(w, "Access denied", http.StatusForbidden)
        return
    }
    
    // Check rate limit
    if !p.checkRateLimit(r, config) {
        http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
        return
    }
    
    // Select backend using round-robin
    backend := p.selectBackend(config)
    if backend == nil {
        http.Error(w, "No healthy backends available", http.StatusServiceUnavailable)
        return
    }
    
    // Create the reverse proxy
    targetURL := &url.URL{
        Scheme: backend.Scheme,
        Host:   fmt.Sprintf("%s:%d", backend.IP.String(), backend.Port),
    }
    
    proxy := &httputil.ReverseProxy{
        Director: func(req *http.Request) {
            req.URL.Scheme = targetURL.Scheme
            req.URL.Host = targetURL.Host
            req.Host = domain
        },
        ModifyResponse: func(resp *http.Response) error {
            duration := time.Since(start)
            p.metrics.RecordRequest(domain, resp.StatusCode, duration)
            return nil
        },
        ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
            p.metrics.RecordError(domain)
            http.Error(w, "Backend error", http.StatusBadGateway)
        },
    }
    
    proxy.ServeHTTP(w, r)
}

func (p *ProxyServer) checkIPRules(r *http.Request, config *DomainConfig) bool {
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        // If there's no port, use RemoteAddr as is
        host = r.RemoteAddr
    }
    clientIP := net.ParseIP(host)
    if clientIP == nil {
        return false
    }
    
    for _, rule := range config.IPRules {
        if rule.IPRange.Contains(clientIP) {
            return rule.RuleType == "whitelist"
        }
    }
    
    // If no rules match, default to allow
    return true
}

func (p *ProxyServer) checkRateLimit(r *http.Request, config *DomainConfig) bool {
    if config.RateLimit == nil {
        return true
    }
    
    var key string
    if config.RateLimit.PerIP {
        key = fmt.Sprintf("%s-%s", config.Domain, r.RemoteAddr)
    } else {
        key = config.Domain
    }
    
    limiter, _ := p.rateLimits.LoadOrStore(key, rate.NewLimiter(
        rate.Limit(config.RateLimit.RequestsPerSecond),
        config.RateLimit.BurstSize,
    ))
    
    return limiter.(*rate.Limiter).Allow()
}

func (p *ProxyServer) selectBackend(config *DomainConfig) *BackendServer {
    config.mu.Lock()
    defer config.mu.Unlock()
    
    if len(config.Backends) == 0 {
        return nil
    }
    
    // Skip unhealthy backends
    for i := 0; i < len(config.Backends); i++ {
        config.currentBackend = (config.currentBackend + 1) % len(config.Backends)
        backend := config.Backends[config.currentBackend]
        
        if backend.IsActive && backend.HealthStatus != nil && *backend.HealthStatus == "healthy" {
            return backend
        }
    }
    
    return nil
}

func (p *ProxyServer) UpdateDomain(domain string, config *DomainConfig) {
    p.domains.Store(domain, config)
    
    // If SSL is enabled, ensure we have a certificate
    if config.SSLEnabled {
        if err := p.ObtainCertificate(domain); err != nil {
            log.Printf("Error obtaining certificate for %s: %v", domain, err)
        }
    }
}

func (p *ProxyServer) DeleteDomain(domain string) {
    p.domains.Delete(domain)
}

func (p *ProxyServer) ObtainCertificate(domain string) error {
    ctx := context.Background()
    if err := p.certManager.ManageAsync(ctx, []string{domain}); err != nil {
        return fmt.Errorf("failed to obtain certificate for %s: %w", domain, err)
    }
    return nil
}

func (p *ProxyServer) Run(httpPort, httpsPort int) error {
    // Start HTTP server
    go func() {
        server := &http.Server{
            Addr:    fmt.Sprintf(":%d", httpPort),
            Handler: p,
        }
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Printf("HTTP server error: %v\n", err)
        }
    }()

    // Get initial SSL domains and obtain certificates
    var domains []string
    p.domains.Range(func(key, value interface{}) bool {
        domain := key.(string)
        config := value.(*DomainConfig)
        if config.SSLEnabled {
            domains = append(domains, domain)
            // Obtain certificate for each domain
            if err := p.ObtainCertificate(domain); err != nil {
                log.Printf("Error obtaining certificate for %s: %v", domain, err)
            }
        }
        return true
    })

    log.Printf("Managing SSL certificates for domains: %v", domains)

    // Start HTTPS server with TLS config from certmagic
    server := &http.Server{
        Addr:      fmt.Sprintf(":%d", httpsPort),
        Handler:   p,
        TLSConfig: p.certManager.TLSConfig(),
    }
    
    return server.ListenAndServeTLS("", "") // Empty strings because certmagic handles the certs
}

func (p *ProxyServer) Metrics() *MetricsCollector {
    return p.metrics
}

func (p *ProxyServer) ConfigureCertmagic(email string) error {
    // Set default config for certmagic
    certmagic.DefaultACME.Email = email
    certmagic.DefaultACME.Agreed = true  // Agree to Let's Encrypt terms
    
    // Optional: Configure storage location
    dataDir := "/root/.local/share/certmagic"  // Or another permanent location
    if err := os.MkdirAll(dataDir, 0700); err != nil {
        return fmt.Errorf("failed to create certmagic directory: %w", err)
    }
    
    // Configure storage for the default config
    certmagic.Default.Storage = &certmagic.FileStorage{Path: dataDir}
    
    return nil
}