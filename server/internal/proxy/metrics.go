package proxy

import (
	"sync"
	"time"
)

type MetricsCollector struct {
	metrics sync.Map // map[string]*DomainMetrics
}

type DomainMetrics struct {
	RequestCount int
	ErrorCount   int
	Latencies    []float64
	mu           sync.Mutex
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
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

func (m *MetricsCollector) RecordError(domain string) {
	metricsVal, _ := m.metrics.LoadOrStore(domain, &DomainMetrics{})
	metrics := metricsVal.(*DomainMetrics)

	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.ErrorCount++
}

func (m *MetricsCollector) GetMetrics(domain string) (requests int, errors int) {
	metricsVal, ok := m.metrics.Load(domain)
	if !ok {
		return 0, 0
	}
	metrics := metricsVal.(*DomainMetrics)

	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	return metrics.RequestCount, metrics.ErrorCount
}
