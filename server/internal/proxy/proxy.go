package proxy

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
)

type ProxyServer struct {
	domains     sync.Map // map[string]*DomainConfig
	rateLimits  sync.Map // map[string]*rate.Limiter
	metrics     *MetricsCollector
	certManager *certmagic.Config
	acmeIssuer  *certmagic.ACMEIssuer
}

type DomainConfig struct {
	Domain            string
	TargetURL         string
	SSLEnabled        bool
	ForceHTTPS        bool
	HTTP2Enabled      bool
	HSTSEnabled       bool
	HSTSMaxAge        int
	WebsocketEnabled  bool
	CustomHeaders     []byte
	BasicAuthEnabled  bool
	BasicAuthUsername string
	BasicAuthPassword string
	IPRules           []*IPRule
	RateLimit         *RateLimit
	Redirects         []*Redirect
	mu                sync.Mutex
}

type IPRule struct {
	ID          int64
	IPRange     net.IPNet
	RuleType    string // "whitelist" or "blacklist"
	Description string
}

type RateLimit struct {
	ID                int64
	RequestsPerSecond int
	BurstSize         int
	Enabled           bool
}

type Redirect struct {
	ID            int64
	SourcePath    string
	TargetURL     string
	RedirectType  int // 301, 302, 307, 308
	PreserveQuery bool
	Enabled       bool
}

func NewProxyServer() (*ProxyServer, error) {
	certConfig := certmagic.NewDefault()

	return &ProxyServer{
		certManager: certConfig,
		metrics:     NewMetricsCollector(),
	}, nil
}

func (p *ProxyServer) handleACMEChallenge(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/.well-known/acme-challenge/") {
		return false
	}

	log.Printf("Handling ACME challenge for path: %s, host: %s", r.URL.Path, r.Host)

	if p.acmeIssuer != nil && p.acmeIssuer.HandleHTTPChallenge(w, r) {
		log.Printf("Successfully served ACME challenge via certmagic for %s", r.Host)
		return true
	}

	log.Printf("ACME challenge not found for %s%s", r.Host, r.URL.Path)
	http.Error(w, "Challenge not found", http.StatusNotFound)
	return true
}

func (p *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.handleACMEChallenge(w, r) {
		return
	}

	start := time.Now()
	domain := r.Host

	if host, _, err := net.SplitHostPort(domain); err == nil {
		domain = host
	}

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

	// Check basic auth
	if config.BasicAuthEnabled {
		if !p.checkBasicAuth(r, config) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Check redirects
	if redirect := p.checkRedirects(r, config); redirect != nil {
		targetURL := redirect.TargetURL
		if redirect.PreserveQuery && r.URL.RawQuery != "" {
			if strings.Contains(targetURL, "?") {
				targetURL += "&" + r.URL.RawQuery
			} else {
				targetURL += "?" + r.URL.RawQuery
			}
		}
		http.Redirect(w, r, targetURL, redirect.RedirectType)
		return
	}

	// Parse target URL
	targetURL, err := url.Parse(config.TargetURL)
	if err != nil {
		log.Printf("Invalid target URL for %s: %v", domain, err)
		http.Error(w, "Invalid backend configuration", http.StatusInternalServerError)
		return
	}

	log.Printf("Proxying request for %s to %s", domain, config.TargetURL)

	// Parse custom headers
	var customHeaders map[string]string
	if len(config.CustomHeaders) > 0 {
		json.Unmarshal(config.CustomHeaders, &customHeaders)
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.URL.Path = singleJoiningSlash(targetURL.Path, req.URL.Path)

			if targetURL.RawQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetURL.RawQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetURL.RawQuery + "&" + req.URL.RawQuery
			}

			// Set Host header to target
			req.Host = targetURL.Host

			// Preserve original client IP
			if clientIP := req.Header.Get("X-Forwarded-For"); clientIP != "" {
				req.Header.Set("X-Real-IP", strings.Split(clientIP, ",")[0])
			} else if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
				req.Header.Set("X-Real-IP", host)
				req.Header.Set("X-Forwarded-For", host)
			}

			req.Header.Set("X-Forwarded-Host", domain)
			req.Header.Set("X-Forwarded-Proto", "https")

			// Apply custom headers
			for key, value := range customHeaders {
				req.Header.Set(key, value)
			}
		},
		ModifyResponse: func(resp *http.Response) error {
			duration := time.Since(start)
			p.metrics.RecordRequest(domain, resp.StatusCode, duration)

			// Add HSTS header if enabled
			if config.HSTSEnabled {
				resp.Header.Set("Strict-Transport-Security", fmt.Sprintf("max-age=%d; includeSubDomains", config.HSTSMaxAge))
			}

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
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			ForceAttemptHTTP2:     config.HTTP2Enabled,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	proxy.ServeHTTP(w, r)
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func (p *ProxyServer) checkIPRules(r *http.Request, config *DomainConfig) bool {
	if len(config.IPRules) == 0 {
		return true
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	clientIP := net.ParseIP(host)
	if clientIP == nil {
		return false
	}

	// Check whitelist first
	hasWhitelist := false
	for _, rule := range config.IPRules {
		if rule.RuleType == "whitelist" {
			hasWhitelist = true
			if rule.IPRange.Contains(clientIP) {
				return true
			}
		}
	}

	// If there's a whitelist and IP wasn't in it, deny
	if hasWhitelist {
		return false
	}

	// Check blacklist
	for _, rule := range config.IPRules {
		if rule.RuleType == "blacklist" && rule.IPRange.Contains(clientIP) {
			return false
		}
	}

	return true
}

func (p *ProxyServer) checkRateLimit(r *http.Request, config *DomainConfig) bool {
	if config.RateLimit == nil || !config.RateLimit.Enabled {
		return true
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	key := fmt.Sprintf("%s-%s", config.Domain, host)

	limiter, _ := p.rateLimits.LoadOrStore(key, rate.NewLimiter(
		rate.Limit(config.RateLimit.RequestsPerSecond),
		config.RateLimit.BurstSize,
	))

	return limiter.(*rate.Limiter).Allow()
}

func (p *ProxyServer) checkBasicAuth(r *http.Request, config *DomainConfig) bool {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false
	}

	if !strings.HasPrefix(auth, "Basic ") {
		return false
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
	if err != nil {
		return false
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return false
	}

	return parts[0] == config.BasicAuthUsername && parts[1] == config.BasicAuthPassword
}

func (p *ProxyServer) checkRedirects(r *http.Request, config *DomainConfig) *Redirect {
	for _, redirect := range config.Redirects {
		if !redirect.Enabled {
			continue
		}
		// Simple path matching (could be enhanced with regex)
		if strings.HasPrefix(r.URL.Path, redirect.SourcePath) {
			return redirect
		}
	}
	return nil
}

func (p *ProxyServer) UpdateDomain(domain string, config *DomainConfig) {
	p.domains.Store(domain, config)

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

	cleanDomain := domain
	if strings.HasPrefix(domain, "https://") {
		cleanDomain = strings.TrimPrefix(domain, "https://")
	} else if strings.HasPrefix(domain, "http://") {
		cleanDomain = strings.TrimPrefix(domain, "http://")
	}

	if cleanDomain != domain {
		log.Printf("Requesting certificate for %s (stripped from %s)", cleanDomain, domain)
	}

	dataDir := "/root/.local/share/certmagic"
	httpChallengeDomainDir := filepath.Join(dataDir, "acme", "http-01", cleanDomain)
	if err := os.MkdirAll(httpChallengeDomainDir, 0700); err != nil {
		log.Printf("Warning: could not create challenge directory for %s: %v", cleanDomain, err)
	}

	altChallengeDomainDir := filepath.Join(dataDir, "acme-http-01", cleanDomain)
	if err := os.MkdirAll(altChallengeDomainDir, 0700); err != nil {
		log.Printf("Warning: could not create alt challenge directory for %s: %v", cleanDomain, err)
	}

	issuer := certmagic.NewACMEIssuer(p.certManager, certmagic.ACMEIssuer{
		CA:                      certmagic.DefaultACME.CA,
		Email:                   certmagic.DefaultACME.Email,
		Agreed:                  true,
		DisableHTTPChallenge:    false,
		DisableTLSALPNChallenge: true,
		AltHTTPPort:             80,
		Logger:                  certmagic.DefaultACME.Logger,
	})

	p.certManager.Issuers = []certmagic.Issuer{issuer}

	log.Printf("Requesting certificate management for %s", cleanDomain)
	if err := p.certManager.ManageAsync(ctx, []string{cleanDomain}); err != nil {
		return fmt.Errorf("failed to obtain certificate for %s: %w", cleanDomain, err)
	}

	log.Printf("Certificate request initiated for %s", cleanDomain)
	return nil
}

func (p *ProxyServer) ConfigureCertmagic(email string) error {
	dataDir := "/root/.local/share/certmagic"

	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return fmt.Errorf("failed to create certmagic directory: %w", err)
	}

	httpChallengeDir := filepath.Join(dataDir, "acme", "http-01")
	if err := os.MkdirAll(httpChallengeDir, 0700); err != nil {
		return fmt.Errorf("failed to create HTTP challenge directory: %w", err)
	}

	altChallengeDir := filepath.Join(dataDir, "acme-http-01")
	if err := os.MkdirAll(altChallengeDir, 0700); err != nil {
		return fmt.Errorf("failed to create alternative HTTP challenge directory: %w", err)
	}

	storage := &certmagic.FileStorage{Path: dataDir}
	certmagic.Default.Storage = storage

	certConfig := certmagic.NewDefault()
	certConfig.Storage = storage

	certmagic.DefaultACME.Email = email
	certmagic.DefaultACME.Agreed = true
	certmagic.DefaultACME.DisableHTTPChallenge = false
	certmagic.DefaultACME.DisableTLSALPNChallenge = true

	acmeIssuer := certmagic.NewACMEIssuer(certConfig, certmagic.ACMEIssuer{
		CA:                      certmagic.DefaultACME.CA,
		Email:                   email,
		Agreed:                  true,
		DisableHTTPChallenge:    false,
		DisableTLSALPNChallenge: true,
		AltHTTPPort:             80,
		Logger:                  certmagic.DefaultACME.Logger,
	})

	certConfig.Issuers = []certmagic.Issuer{acmeIssuer}

	p.certManager = certConfig
	p.acmeIssuer = acmeIssuer

	log.Printf("Certmagic configured with email: %s, storage path: %s", email, dataDir)

	return nil
}

func (p *ProxyServer) Run(httpPort, httpsPort int) error {
	log.Printf("Starting proxy server with HTTP port %d, HTTPS port %d", httpPort, httpsPort)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", httpPort),
		Handler:      http.HandlerFunc(p.httpHandler),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

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

	select {}
}

func (p *ProxyServer) Metrics() *MetricsCollector {
	return p.metrics
}

func (p *ProxyServer) Shutdown(ctx context.Context) error {
	log.Println("Proxy server shutting down...")
	return nil
}

func (p *ProxyServer) httpHandler(w http.ResponseWriter, r *http.Request) {
	if p.handleACMEChallenge(w, r) {
		log.Printf("Served ACME challenge for %s", r.Host)
		return
	}

	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	configVal, ok := p.domains.Load(host)
	if !ok {
		log.Printf("Domain not found: %s", host)
		http.Error(w, "Domain not found", http.StatusNotFound)
		return
	}

	config := configVal.(*DomainConfig)
	if config.SSLEnabled && config.ForceHTTPS {
		u := r.URL
		u.Host = r.Host
		u.Scheme = "https"
		http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
		return
	}

	p.ServeHTTP(w, r)
}

// storeACMEChallenge stores ACME challenge for manual handling if needed
func (p *ProxyServer) storeACMEChallenge(domain, token, keyAuth string) error {
	dataDir := "/root/.local/share/certmagic"

	locations := []string{
		filepath.Join(dataDir, "acme", "http-01", domain, token),
		filepath.Join(dataDir, "acme-http-01", domain, token),
	}

	for _, location := range locations {
		if err := os.MkdirAll(filepath.Dir(location), 0700); err != nil {
			log.Printf("Warning: failed to create directory for challenge token at %s: %v", location, err)
			continue
		}

		if err := os.WriteFile(location, []byte(keyAuth), 0600); err != nil {
			log.Printf("Warning: failed to write challenge token to %s: %v", location, err)
			continue
		}

		log.Printf("Successfully stored ACME challenge token at %s", location)
	}

	if err := p.certManager.Storage.Store(context.Background(), path.Join("acme", "http-01", domain, token), []byte(keyAuth)); err != nil {
		log.Printf("Warning: failed to store challenge token via storage interface: %v", err)
	} else {
		log.Printf("Successfully stored ACME challenge token via storage interface")
	}

	return nil
}
