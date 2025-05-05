package proxy

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type MetricsCollector struct {
    db        *pgxpool.Pool
    metrics   sync.Map // map[string]*DomainMetrics
    flushChan chan struct{}
}

type DomainMetrics struct {
    RequestCount  int
    ErrorCount    int
    TCPCount     int
    Latencies    []float64
    TCPLatencies []float64
    mu           sync.Mutex
}

func NewMetricsCollector() *MetricsCollector {
    m := &MetricsCollector{
        flushChan: make(chan struct{}),
    }
    go m.periodicFlush()
    return m
}

func (m *MetricsCollector) SetDB(db *pgxpool.Pool) {
    m.db = db
}

func (m *MetricsCollector) RecordRequest(domain string, statusCode int, duration time.Duration) {
    metricsVal, _ := m.metrics.LoadOrStore(domain, &DomainMetrics{})
    metrics := metricsVal.(*DomainMetrics)

    metrics.mu.Lock()
    defer metrics.mu.Unlock()

    metrics.RequestCount++
    metrics.Latencies = append(metrics.Latencies, float64(duration.Milliseconds()))

    if statusCode >= 400 {
        metrics.ErrorCount++
    }
}

func (m *MetricsCollector) RecordTCPRequest(domain string, duration time.Duration) {
    metricsVal, _ := m.metrics.LoadOrStore(domain, &DomainMetrics{})
    metrics := metricsVal.(*DomainMetrics)

    metrics.mu.Lock()
    defer metrics.mu.Unlock()

    metrics.TCPCount++
    metrics.TCPLatencies = append(metrics.TCPLatencies, float64(duration.Milliseconds()))
}

func (m *MetricsCollector) RecordError(domain string) {
    metricsVal, _ := m.metrics.LoadOrStore(domain, &DomainMetrics{})
    metrics := metricsVal.(*DomainMetrics)

    metrics.mu.Lock()
    defer metrics.mu.Unlock()

    metrics.ErrorCount++
}

func (m *MetricsCollector) periodicFlush() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            m.flush()
        case <-m.flushChan:
            m.flush()
        }
    }
}

func (m *MetricsCollector) flush() {
    if m.db == nil {
        return
    }

    m.metrics.Range(func(key, value interface{}) bool {
        domain := key.(string)
        metrics := value.(*DomainMetrics)

        metrics.mu.Lock()
        defer metrics.mu.Unlock()

        if metrics.RequestCount == 0 && metrics.TCPCount == 0 {
            return true
        }

        // Calculate HTTP percentiles
        var p95, p99 float64
        if len(metrics.Latencies) > 0 {
            sorted := make([]float64, len(metrics.Latencies))
            copy(sorted, metrics.Latencies)
            sort.Float64s(sorted)

            p95 = sorted[int(float64(len(sorted))*0.95)]
            p99 = sorted[int(float64(len(sorted))*0.99)]
        }

        // Calculate TCP percentiles
        var tcpP95, tcpP99 float64
        if len(metrics.TCPLatencies) > 0 {
            sorted := make([]float64, len(metrics.TCPLatencies))
            copy(sorted, metrics.TCPLatencies)
            sort.Float64s(sorted)

            tcpP95 = sorted[int(float64(len(sorted))*0.95)]
            tcpP99 = sorted[int(float64(len(sorted))*0.99)]
        }

        // Calculate average HTTP latency
        var avgLatency float64
        if len(metrics.Latencies) > 0 {
            sum := 0.0
            for _, lat := range metrics.Latencies {
                sum += lat
            }
            avgLatency = sum / float64(len(metrics.Latencies))
        }

        // Calculate average TCP latency
        var avgTCPLatency float64
        if len(metrics.TCPLatencies) > 0 {
            sum := 0.0
            for _, lat := range metrics.TCPLatencies {
                sum += lat
            }
            avgTCPLatency = sum / float64(len(metrics.TCPLatencies))
        }

        // First, check if the domain exists and get its ID
        ctx := context.Background()
        var domainID int
        err := m.db.QueryRow(ctx, 
            "SELECT id FROM domains WHERE target_url = $1",
            domain,
        ).Scan(&domainID)

        if err != nil {
            if err == pgx.ErrNoRows {
                fmt.Printf("Warning: Skipping metrics for unknown domain: %s\n", domain)
                return true
            }
            fmt.Printf("Error querying domain: %v\n", err)
            return true
        }

        // Insert HTTP metrics into database
        if metrics.RequestCount > 0 {
            _, err = m.db.Exec(ctx,
                `INSERT INTO request_metrics 
                (domain_id, timestamp, request_count, error_count, avg_latency_ms, p95_latency_ms, p99_latency_ms)
                VALUES ($1, $2, $3, $4, $5, $6, $7)`,
                domainID,
                time.Now(),
                metrics.RequestCount,
                metrics.ErrorCount,
                avgLatency,
                p95,
                p99,
            )

            if err != nil {
                fmt.Printf("Error flushing HTTP metrics: %v\n", err)
            }
        }

        // Insert TCP metrics into database
        if metrics.TCPCount > 0 {
            _, err = m.db.Exec(ctx,
                `INSERT INTO tcp_metrics 
                (domain_id, timestamp, connection_count, avg_latency_ms, p95_latency_ms, p99_latency_ms)
                VALUES ($1, $2, $3, $4, $5, $6)`,
                domainID,
                time.Now(),
                metrics.TCPCount,
                avgTCPLatency,
                tcpP95,
                tcpP99,
            )

            if err != nil {
                fmt.Printf("Error flushing TCP metrics: %v\n", err)
            }
        }

        // Reset metrics
        metrics.RequestCount = 0
        metrics.ErrorCount = 0
        metrics.TCPCount = 0
        metrics.Latencies = metrics.Latencies[:0]
        metrics.TCPLatencies = metrics.TCPLatencies[:0]

        return true
    })
}