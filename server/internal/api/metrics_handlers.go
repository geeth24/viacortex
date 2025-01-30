package api

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "time"

    "github.com/go-chi/chi/v5"
)

// getGlobalMetrics returns metrics across all domains
func (h *Handlers) getGlobalMetrics(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse time range from query parameters
    timeRange := r.URL.Query().Get("range")
    if timeRange == "" {
        timeRange = "24h" // Default to last 24 hours
    }
    
    duration, err := time.ParseDuration(timeRange)
    if err != nil {
        http.Error(w, "Invalid time range", http.StatusBadRequest)
        return
    }

    startTime := time.Now().Add(-duration)

    rows, err := h.db.Query(ctx, `
        SELECT 
            domain_id,
            SUM(request_count) as total_requests,
            SUM(error_count) as total_errors,
            AVG(avg_latency_ms) as avg_latency,
            MAX(p95_latency_ms) as max_p95_latency,
            MAX(p99_latency_ms) as max_p99_latency
        FROM request_metrics
        WHERE timestamp > $1
        GROUP BY domain_id
    `, startTime)
    
    if err != nil {
        log.Printf("Error fetching metrics: %v", err)
        http.Error(w, "Failed to fetch metrics", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    metrics := []map[string]interface{}{}
    for rows.Next() {
        var m struct {
            DomainID      int64   `json:"domain_id"`
            TotalRequests int     `json:"total_requests"`
            TotalErrors   int     `json:"total_errors"`
            AvgLatency    float64 `json:"avg_latency_ms"`
            MaxP95Latency float64 `json:"max_p95_latency_ms"`
            MaxP99Latency float64 `json:"max_p99_latency_ms"`
        }
        
        err := rows.Scan(
            &m.DomainID, &m.TotalRequests, &m.TotalErrors,
            &m.AvgLatency, &m.MaxP95Latency, &m.MaxP99Latency,
        )
        if err != nil {
            log.Printf("Error scanning metrics: %v", err)
            continue
        }
        
        metrics = append(metrics, map[string]interface{}{
            "domain_id":          m.DomainID,
            "total_requests":     m.TotalRequests,
            "total_errors":       m.TotalErrors,
            "error_rate":         float64(m.TotalErrors) / float64(m.TotalRequests),
            "avg_latency_ms":     m.AvgLatency,
            "max_p95_latency_ms": m.MaxP95Latency,
            "max_p99_latency_ms": m.MaxP99Latency,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(metrics)
}

// getDomainMetrics returns metrics for a specific domain
func (h *Handlers) getDomainMetrics(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    domainID := chi.URLParam(r, "domainID")
    
    timeRange := r.URL.Query().Get("range")
    if timeRange == "" {
        timeRange = "24h"
    }
    
    duration, err := time.ParseDuration(timeRange)
    if err != nil {
        http.Error(w, "Invalid time range", http.StatusBadRequest)
        return
    }

    startTime := time.Now().Add(-duration)
    
    // Get metrics in time series format
    rows, err := h.db.Query(ctx, `
        SELECT 
            timestamp,
            request_count,
            error_count,
            avg_latency_ms,
            p95_latency_ms,
            p99_latency_ms
        FROM request_metrics
        WHERE domain_id = $1 AND timestamp > $2
        ORDER BY timestamp DESC
    `, domainID, startTime)
    
    if err != nil {
        log.Printf("Error fetching domain metrics: %v", err)
        http.Error(w, "Failed to fetch metrics", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    metrics := []map[string]interface{}{}
    for rows.Next() {
        var m struct {
            Timestamp    time.Time `json:"timestamp"`
            Requests     int       `json:"requests"`
            Errors       int       `json:"errors"`
            AvgLatency   float64   `json:"avg_latency_ms"`
            P95Latency   float64   `json:"p95_latency_ms"`
            P99Latency   float64   `json:"p99_latency_ms"`
        }
        
        err := rows.Scan(
            &m.Timestamp, &m.Requests, &m.Errors,
            &m.AvgLatency, &m.P95Latency, &m.P99Latency,
        )
        if err != nil {
            log.Printf("Error scanning domain metrics: %v", err)
            continue
        }
        
        metrics = append(metrics, map[string]interface{}{
            "timestamp":      m.Timestamp,
            "requests":       m.Requests,
            "errors":        m.Errors,
            "error_rate":    float64(m.Errors) / float64(m.Requests),
            "avg_latency":   m.AvgLatency,
            "p95_latency":   m.P95Latency,
            "p99_latency":   m.P99Latency,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(metrics)
}

// getGlobalLogs returns logs across all domains with filtering
func (h *Handlers) getGlobalLogs(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse query parameters for filtering
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit == 0 {
        limit = 100 // Default limit
    }
    
    statusCode, _ := strconv.Atoi(r.URL.Query().Get("status"))
    clientIP := r.URL.Query().Get("client_ip")
    method := r.URL.Query().Get("method")
    
    // Build query with filters
    query := `
        SELECT 
            id, domain_id, timestamp, client_ip, method,
            path, status_code, response_time_ms,
            user_agent, referer
        FROM request_logs
        WHERE 1=1
    `
    args := []interface{}{}
    argCount := 1

    if statusCode != 0 {
        query += ` AND status_code = $` + strconv.Itoa(argCount)
        args = append(args, statusCode)
        argCount++
    }
    
    if clientIP != "" {
        query += ` AND client_ip = $` + strconv.Itoa(argCount)
        args = append(args, clientIP)
        argCount++
    }
    
    if method != "" {
        query += ` AND method = $` + strconv.Itoa(argCount)
        args = append(args, method)
        argCount++
    }
    
    query += ` ORDER BY timestamp DESC LIMIT $` + strconv.Itoa(argCount)
    args = append(args, limit)

    rows, err := h.db.Query(ctx, query, args...)
    if err != nil {
        log.Printf("Error fetching logs: %v", err)
        http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    logs := []map[string]interface{}{}
    for rows.Next() {
        var l struct {
            ID            int64     `json:"id"`
            DomainID      int64     `json:"domain_id"`
            Timestamp     time.Time `json:"timestamp"`
            ClientIP      string    `json:"client_ip"`
            Method        string    `json:"method"`
            Path          string    `json:"path"`
            StatusCode    int       `json:"status_code"`
            ResponseTime  int       `json:"response_time_ms"`
            UserAgent     string    `json:"user_agent"`
            Referer      string    `json:"referer"`
        }
        
        err := rows.Scan(
            &l.ID, &l.DomainID, &l.Timestamp, &l.ClientIP,
            &l.Method, &l.Path, &l.StatusCode, &l.ResponseTime,
            &l.UserAgent, &l.Referer,
        )
        if err != nil {
            log.Printf("Error scanning log: %v", err)
            continue
        }
        
        logs = append(logs, map[string]interface{}{
            "id":              l.ID,
            "domain_id":       l.DomainID,
            "timestamp":       l.Timestamp,
            "client_ip":       l.ClientIP,
            "method":         l.Method,
            "path":           l.Path,
            "status_code":     l.StatusCode,
            "response_time":   l.ResponseTime,
            "user_agent":      l.UserAgent,
            "referer":        l.Referer,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(logs)
}

// getDomainLogs returns logs for a specific domain with filtering
func (h *Handlers) getDomainLogs(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    domainID := chi.URLParam(r, "domainID")
    
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit == 0 {
        limit = 100
    }
    
    statusCode, _ := strconv.Atoi(r.URL.Query().Get("status"))
    clientIP := r.URL.Query().Get("client_ip")
    method := r.URL.Query().Get("method")
    
    query := `
        SELECT 
            id, timestamp, client_ip, method,
            path, status_code, response_time_ms,
            user_agent, referer
        FROM request_logs
        WHERE domain_id = $1
    `
    args := []interface{}{domainID}
    argCount := 2

    if statusCode != 0 {
        query += ` AND status_code = $` + strconv.Itoa(argCount)
        args = append(args, statusCode)
        argCount++
    }
    
    if clientIP != "" {
        query += ` AND client_ip = $` + strconv.Itoa(argCount)
        args = append(args, clientIP)
        argCount++
    }
    
    if method != "" {
        query += ` AND method = $` + strconv.Itoa(argCount)
        args = append(args, method)
        argCount++
    }
    
    query += ` ORDER BY timestamp DESC LIMIT $` + strconv.Itoa(argCount)
    args = append(args, limit)

    rows, err := h.db.Query(ctx, query, args...)
    if err != nil {
        log.Printf("Error fetching domain logs: %v", err)
        http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    logs := []map[string]interface{}{}
    for rows.Next() {
        var l struct {
            ID            int64     `json:"id"`
            Timestamp     time.Time `json:"timestamp"`
            ClientIP      string    `json:"client_ip"`
            Method        string    `json:"method"`
            Path          string    `json:"path"`
            StatusCode    int       `json:"status_code"`
            ResponseTime  int       `json:"response_time_ms"`
            UserAgent     string    `json:"user_agent"`
            Referer      string    `json:"referer"`
        }
        
        err := rows.Scan(
            &l.ID, &l.Timestamp, &l.ClientIP, &l.Method,
            &l.Path, &l.StatusCode, &l.ResponseTime,
            &l.UserAgent, &l.Referer,
        )
        if err != nil {
            log.Printf("Error scanning domain log: %v", err)
            continue
        }
        
        logs = append(logs, map[string]interface{}{
            "id":             l.ID,
            "timestamp":      l.Timestamp,
            "client_ip":      l.ClientIP,
            "method":         l.Method,
            "path":          l.Path,
            "status_code":    l.StatusCode,
            "response_time":  l.ResponseTime,
            "user_agent":     l.UserAgent,
            "referer":       l.Referer,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(logs)
}