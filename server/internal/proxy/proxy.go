package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/caddyserver/certmagic"
	"golang.org/x/time/rate"
	"crypto/tls"
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

// storeACMEChallenge is a helper to manually create an ACME challenge token file if needed
func (p *ProxyServer) storeACMEChallenge(domain, token, keyAuth string) error {
	// Ensure base directories exist
	dataDir := "/root/.local/share/certmagic"
	
	// Store in multiple possible locations for compatibility
	locations := []string{
		filepath.Join(dataDir, "acme", "http-01", domain, token),
		filepath.Join(dataDir, "acme-http-01", domain, token),
	}
	
	for _, location := range locations {
		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(location), 0700); err != nil {
			log.Printf("Warning: failed to create directory for challenge token at %s: %v", location, err)
			continue
		}
		
		// Write the token
		if err := os.WriteFile(location, []byte(keyAuth), 0600); err != nil {
			log.Printf("Warning: failed to write challenge token to %s: %v", location, err)
			continue
		}
		
		log.Printf("Successfully stored ACME challenge token at %s", location)
	}
	
	// Also try to store via the storage interface
	if err := p.certManager.Storage.Store(context.Background(), path.Join("acme", "http-01", domain, token), []byte(keyAuth)); err != nil {
		log.Printf("Warning: failed to store challenge token via storage interface: %v", err)
	} else {
		log.Printf("Successfully stored ACME challenge token via storage interface")
	}
	
	return nil
}

// handleACMEChallenge handles HTTP-01 ACME challenges
func (p *ProxyServer) handleACMEChallenge(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/.well-known/acme-challenge/") {
		return false
	}

	// Get the token from the path
	token := path.Base(r.URL.Path)
	
	log.Printf("Handling ACME challenge for token: %s, host: %s", token, r.Host)
	
	// Get the key authorization from certmagic's storage
	challengePath := path.Join("acme", "http-01", r.Host, token)
	keyAuth, err := p.certManager.Storage.Load(context.Background(), challengePath)
	if err != nil {
		// Try alternate path format used by some certmagic versions
		challengePath = path.Join("acme-http-01", r.Host, token)
		keyAuth, err = p.certManager.Storage.Load(context.Background(), challengePath)
		if err != nil {
			log.Printf("ACME challenge error for token %s: %v", token, err)
			
			// As a fallback, check if token exists directly in the storage directory
			dataDir := "/root/.local/share/certmagic"
			tokenPath := filepath.Join(dataDir, "acme", "http-01", r.Host, token)
			log.Printf("Trying to read token directly from: %s", tokenPath)
			
			if content, err := os.ReadFile(tokenPath); err == nil {
				log.Printf("Successfully read token from direct file: %s", tokenPath)
				w.Header().Set("Content-Type", "text/plain")
				w.Write(content)
				return true
			}
			
			http.Error(w, "Challenge not found", http.StatusNotFound)
			return true
		}
	}

	log.Printf("Successfully serving ACME challenge for %s: %s", r.Host, string(keyAuth))
	
	// Serve the challenge response
	w.Header().Set("Content-Type", "text/plain")
	w.Write(keyAuth)
	return true
}

func (p *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check for ACME challenge first
	if p.handleACMEChallenge(w, r) {
		return
	}

	start := time.Now()
	domain := r.Host

	// Strip port from domain if present
	if host, _, err := net.SplitHostPort(domain); err == nil {
		domain = host
	}
	
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

			// Preserve original client IP if behind another proxy
			if clientIP := req.Header.Get("X-Forwarded-For"); clientIP != "" {
				req.Header.Set("X-Real-IP", clientIP)
			} else {
				req.Header.Set("X-Real-IP", req.RemoteAddr)
			}
		},
		ModifyResponse: func(resp *http.Response) error {
			duration := time.Since(start)
			p.metrics.RecordRequest(domain, resp.StatusCode, duration)
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error for %s: %v", domain, err)
			p.metrics.RecordError(domain)
			http.Error(w, "Backend error", http.StatusBadGateway)
		},
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
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
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		key = fmt.Sprintf("%s-%s", config.Domain, host)
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
		
		if backend.IsActive && (backend.HealthStatus == nil || *backend.HealthStatus == "healthy") {
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
	
	// Strip any protocol prefixes to get a clean domain name
	cleanDomain := domain
	if strings.HasPrefix(domain, "https://") {
		cleanDomain = strings.TrimPrefix(domain, "https://")
	} else if strings.HasPrefix(domain, "http://") {
		cleanDomain = strings.TrimPrefix(domain, "http://")
	} else if strings.HasPrefix(domain, "tcp://") {
		cleanDomain = strings.TrimPrefix(domain, "tcp://")
	}
	
	// Log the domain transformation for debugging
	if cleanDomain != domain {
		log.Printf("Requesting certificate for %s (stripped from %s)", cleanDomain, domain)
	}
	
	// Ensure challenge directories exist for this specific domain
	dataDir := "/root/.local/share/certmagic"
	httpChallengeDomainDir := filepath.Join(dataDir, "acme", "http-01", cleanDomain)
	if err := os.MkdirAll(httpChallengeDomainDir, 0700); err != nil {
		log.Printf("Warning: could not create challenge directory for %s: %v", cleanDomain, err)
	}
	
	// Also create the alternative path used by some certmagic versions
	altChallengeDomainDir := filepath.Join(dataDir, "acme-http-01", cleanDomain)
	if err := os.MkdirAll(altChallengeDomainDir, 0700); err != nil {
		log.Printf("Warning: could not create alt challenge directory for %s: %v", cleanDomain, err)
	}
	
	// Configure with HTTP-01 only for this request
	issuer := certmagic.NewACMEIssuer(p.certManager, certmagic.ACMEIssuer{
		CA:                      certmagic.DefaultACME.CA,
		Email:                   certmagic.DefaultACME.Email,
		Agreed:                  true,
		DisableHTTPChallenge:    false,
		DisableTLSALPNChallenge: true,
		AltHTTPPort:             80, // Ensure we're using standard HTTP port
		Logger:                  certmagic.DefaultACME.Logger,
	})
	
	// Create a temporary issuer just for this certificate
	p.certManager.Issuers = []certmagic.Issuer{issuer}
	
	// Request certificate management
	log.Printf("Requesting certificate management for %s", cleanDomain)
	if err := p.certManager.ManageAsync(ctx, []string{cleanDomain}); err != nil {
		return fmt.Errorf("failed to obtain certificate for %s: %w", cleanDomain, err)
	}
	
	log.Printf("Certificate request initiated for %s", cleanDomain)
	return nil
}

func (p *ProxyServer) ConfigureCertmagic(email string) error {
	// Configure storage location
	dataDir := "/root/.local/share/certmagic"
	
	// Ensure directories exist
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return fmt.Errorf("failed to create certmagic directory: %w", err)
	}
	
	// Create additional directories needed for HTTP-01 challenges
	httpChallengeDir := filepath.Join(dataDir, "acme", "http-01")
	if err := os.MkdirAll(httpChallengeDir, 0700); err != nil {
		return fmt.Errorf("failed to create HTTP challenge directory: %w", err)
	}
	
	// Also create the alternative path used by some certmagic versions
	altChallengeDir := filepath.Join(dataDir, "acme-http-01")
	if err := os.MkdirAll(altChallengeDir, 0700); err != nil {
		return fmt.Errorf("failed to create alternative HTTP challenge directory: %w", err)
	}
	
	// Configure storage for certmagic
	storage := &certmagic.FileStorage{Path: dataDir}
	certmagic.Default.Storage = storage
	
	// Set up the certmagic instance
	certConfig := certmagic.NewDefault()
	certConfig.Storage = storage
	
	// Set default config for ACME
	certmagic.DefaultACME.Email = email
	certmagic.DefaultACME.Agreed = true
	certmagic.DefaultACME.DisableHTTPChallenge = false
	certmagic.DefaultACME.DisableTLSALPNChallenge = true
	
	// Create ACME issuer
	acmeIssuer := certmagic.NewACMEIssuer(certConfig, certmagic.ACMEIssuer{
		CA:                      certmagic.DefaultACME.CA,
		Email:                   email,
		Agreed:                  true,
		DisableHTTPChallenge:    false,
		DisableTLSALPNChallenge: true,
		AltHTTPPort:             80, // Ensure we're using standard HTTP port
		Logger:                  certmagic.DefaultACME.Logger,
	})
	
	// Set issuer for the config
	certConfig.Issuers = []certmagic.Issuer{acmeIssuer}
	
	// Store the configured certmagic instance
	p.certManager = certConfig
	
	log.Printf("Certmagic configured with email: %s, storage path: %s", email, dataDir)
	
	// For testing/debugging purposes, uncomment to use staging environment
	// certmagic.DefaultACME.CA = certmagic.LetsEncryptStagingCA
	
	return nil
}

func (p *ProxyServer) Run(httpPort, httpsPort int) error {
	log.Printf("Starting proxy server with HTTP port %d, HTTPS port %d, and TCP proxies", httpPort, httpsPort)

	// Start TCP proxy listeners for different protocols
	// Important: Start this first, before HTTP/HTTPS
	go p.startTCPProxies()

	// HTTP server (for redirects & ACME challenges)
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", httpPort),
		Handler:      http.HandlerFunc(p.httpHandler),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// HTTPS server
	httpsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpsPort),
		Handler: p,
		TLSConfig: &tls.Config{
			GetCertificate: p.certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start the servers in goroutines
	go func() {
		log.Printf("Starting HTTP server on port %d", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	go func() {
		log.Printf("Starting HTTPS server on port %d", httpsPort)
		if err := httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTPS server error: %v", err)
		}
	}()

	// Block indefinitely
	select {}
}

// startTCPProxies starts TCP proxy listeners for configured protocols
func (p *ProxyServer) startTCPProxies() {
	// Default TCP ports for various protocols
	protocolPorts := map[string]int{
		"minecraft": 25565,
		// Add other protocol-specific ports as needed
	}
	
	log.Printf("Starting TCP proxies for protocols: %v", protocolPorts)
	
	// Start a listener for each protocol
	for protocol, port := range protocolPorts {
		go func(proto string, portNum int) {
			log.Printf("Starting TCP proxy for %s on port %d in goroutine", proto, portNum)
			p.startTCPProxy(proto, portNum)
		}(protocol, port)
	}
}

// startTCPProxy starts a TCP proxy listener on the specified port for a specific protocol
func (p *ProxyServer) startTCPProxy(protocol string, port int) {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Printf("Setting up TCP proxy listener for %s on %s", protocol, addr)
	
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("TCP proxy listen error for %s on port %d: %v", protocol, port, err)
		return
	}
	
	log.Printf("Successfully started TCP proxy for %s on port %d", protocol, port)
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("TCP accept error on port %d: %v", port, err)
			continue
		}
		
		log.Printf("Accepted new TCP connection on port %d from %s", port, conn.RemoteAddr().String())
		go p.handleTCPConnection(conn, protocol)
	}
}

// handleTCPConnection handles a TCP connection by determining the target and proxying data
func (p *ProxyServer) handleTCPConnection(clientConn net.Conn, protocol string) {
	defer clientConn.Close()
	
	// Get client address
	clientAddr := clientConn.RemoteAddr().String()
	log.Printf("New %s TCP connection from %s", protocol, clientAddr)
	
	// Log all available domains for debugging
	var availableDomains []string
	p.domains.Range(func(key, value interface{}) bool {
		domain := key.(string)
		availableDomains = append(availableDomains, domain)
		return true
	})
	log.Printf("Available domains: %v", availableDomains)
	
	// Find the first domain with TCP backends for this protocol
	var domain string
	var tcpConfig *DomainConfig
	
	p.domains.Range(func(key, value interface{}) bool {
		domainName := key.(string)
		config := value.(*DomainConfig)
		
		log.Printf("Checking domain %s for TCP backends", domainName)
		
		// Check if this domain has any TCP backends
		hasTcpBackend := false
		for _, backend := range config.Backends {
			if backend.Scheme == "tcp" {
				hasTcpBackend = true
				log.Printf("Domain %s has TCP backend: %s:%d (active: %v, health: %v)", 
					domainName, backend.IP, backend.Port, backend.IsActive, 
					backend.HealthStatus)
				
				if backend.IsActive && (backend.HealthStatus == nil || *backend.HealthStatus == "healthy") {
					domain = domainName
					tcpConfig = config
					return false // Stop iterating
				}
			}
		}
		
		if !hasTcpBackend {
			log.Printf("Domain %s has no TCP backends", domainName)
		}
		
		return true // Continue iterating
	})
	
	if domain == "" || tcpConfig == nil {
		log.Printf("No domain with active TCP backends found for %s", protocol)
		return
	}
	
	log.Printf("Using domain %s for %s TCP connection", domain, protocol)
	
	// Select backend using round-robin
	backend := p.selectBackend(tcpConfig)
	if backend == nil {
		log.Printf("No healthy TCP backends available for %s on %s", domain, protocol)
		return
	}
	
	// Only proxy to TCP backends
	if backend.Scheme != "tcp" {
		log.Printf("Backend for %s is not TCP", domain)
		return
	}
	
	// Connect to backend
	backendAddr := fmt.Sprintf("%s:%d", backend.IP.String(), backend.Port)
	log.Printf("Connecting to backend %s", backendAddr)
	backendConn, err := net.Dial("tcp", backendAddr)
	if err != nil {
		log.Printf("TCP backend connection error: %v", err)
		return
	}
	defer backendConn.Close()
	
	log.Printf("Established %s connection to backend at %s", protocol, backendAddr)
	
	// Start proxying data in both directions
	start := time.Now()
	
	// Create a context for this connection
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Create a WaitGroup to wait for both goroutines to finish
	var wg sync.WaitGroup
	wg.Add(2)
	
	// Client to backend
	go func() {
		defer wg.Done()
		defer cancel() // Cancel context if this direction fails
		
		buf := make([]byte, 32*1024) // 32 KB buffer
		for {
			select {
			case <-ctx.Done():
				return
			default:
				clientConn.SetReadDeadline(time.Now().Add(30 * time.Second))
				n, err := clientConn.Read(buf)
				if err != nil {
					if err != io.EOF {
						log.Printf("TCP client read error: %v", err)
					}
					return
				}
				
				backendConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				_, err = backendConn.Write(buf[:n])
				if err != nil {
					log.Printf("TCP backend write error: %v", err)
					return
				}
			}
		}
	}()
	
	// Backend to client
	go func() {
		defer wg.Done()
		defer cancel() // Cancel context if this direction fails
		
		buf := make([]byte, 32*1024) // 32 KB buffer
		for {
			select {
			case <-ctx.Done():
				return
			default:
				backendConn.SetReadDeadline(time.Now().Add(30 * time.Second))
				n, err := backendConn.Read(buf)
				if err != nil {
					if err != io.EOF {
						log.Printf("TCP backend read error: %v", err)
					}
					return
				}
				
				clientConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				_, err = clientConn.Write(buf[:n])
				if err != nil {
					log.Printf("TCP client write error: %v", err)
					return
				}
			}
		}
	}()
	
	// Wait for both goroutines to finish
	wg.Wait()
	
	// Record metrics
	duration := time.Since(start)
	p.metrics.RecordTCPRequest(domain, duration)
	
	log.Printf("TCP connection closed: %s -> %s, duration: %v", clientAddr, backendAddr, duration)
}

func (p *ProxyServer) Metrics() *MetricsCollector {
	return p.metrics
}

// httpHandler handles HTTP requests, primarily for redirecting to HTTPS
func (p *ProxyServer) httpHandler(w http.ResponseWriter, r *http.Request) {
	// First and most important, check for ACME challenges
	if p.handleACMEChallenge(w, r) {
		log.Printf("Served ACME challenge for %s", r.Host)
		return
	}

	// Get the host from the request
	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	
	// Check if this domain is configured
	configVal, ok := p.domains.Load(host)
	if !ok {
		log.Printf("Domain not found: %s", host)
		http.Error(w, "Domain not found", http.StatusNotFound)
		return
	}
	
	config := configVal.(*DomainConfig)
	if config.SSLEnabled {
		// Redirect to HTTPS
		u := r.URL
		u.Host = r.Host
		u.Scheme = "https"
		http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
		return
	}
	
	// If SSL is not enabled, serve the HTTP request
	p.ServeHTTP(w, r)
}