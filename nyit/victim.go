// Here's an example of a Go service that exposes Prometheus metrics. This service includes a simple HTTP server and demonstrates different types of Prometheus metrics (Counter, Gauge, and Histogram):

// ```go
package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define metrics
var (
	requestsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	memoryUsageGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "memory_usage_bytes",
		Help: "Current memory usage in bytes",
	})

	requestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: prometheus.DefBuckets,
	})
)

func main() {
	// Start background task to simulate memory usage
	go updateMemoryUsage()

	// Setup HTTP server
	http.HandleFunc("/", handleRoot)
	http.Handle("/metrics", promhttp.Handler())

	// Start server
	port := ":8080"
	println("Server running on port", port)
	http.ListenAndServe(port, nil)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Simulate work
	time.Sleep(100 * time.Millisecond)

	// Record metrics
	requestsCounter.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
	requestDuration.Observe(time.Since(start).Seconds())
	w.Write([]byte("Hello, World!"))
}

func updateMemoryUsage() {
	for {
		// Simulate memory usage between 50MB and 150MB
		mem := 50_000_000 + float64(time.Now().UnixNano()%100_000_000)
		memoryUsageGauge.Set(mem)
		time.Sleep(1 * time.Second)
	}
}

// ```

// ### Key Components:

// 1. **Metrics Definition**:
//    - `Counter`: Tracks total HTTP requests with labels (method, path, status)
//    - `Gauge`: Simulates memory usage (updated periodically)
//    - `Histogram`: Measures request duration with default buckets

// 2. **Dependencies**:
// ```bash
// go get github.com/prometheus/client_golang/prometheus
// ```

// 3. **Running the Service**:
// ```bash
// go run main.go
// ```

// 4. **Access Metrics**:
// ```bash
// curl http://localhost:8080/metrics
// ```

// Example metrics output:
// ```
// # HELP http_requests_total Total number of HTTP requests
// # TYPE http_requests_total counter
// http_requests_total{method="GET",path="/",status="200"} 5

// # HELP memory_usage_bytes Current memory usage in bytes
// # TYPE memory_usage_bytes gauge
// memory_usage_bytes 123456789

// # HELP http_request_duration_seconds Duration of HTTP requests in seconds
// # TYPE http_request_duration_seconds histogram
// http_request_duration_seconds_bucket{le="0.005"} 0
// http_request_duration_seconds_bucket{le="0.01"} 0
// ...
// ```

// ### To Extend This Service:

// 1. **Add More Metrics**:
// ```go
// cpuUsage := promauto.NewGauge(prometheus.GaugeOpts{
//     Name: "cpu_usage_percent",
//     Help: "Current CPU usage percentage",
// })
// ```

// 2. **Add Labels**:
// ```go
// errorCounter = promauto.NewCounterVec(prometheus.CounterOpts{
//     Name: "errors_total",
//     Help: "Total number of errors",
// }, []string{"type", "service"})
// ```

// 3. **Custom Registries** (for advanced use):
// ```go
// registry := prometheus.NewRegistry()
// registry.MustRegister(customMetric)
// ```

// 4. **Push Gateway Support** (for batch jobs):
// ```go
// push.New("push-gateway:9091", "my_job").Collector(requestsCounter)
// ```

// ### Monitoring with Prometheus:

// 1. Add this to your Prometheus config (`prometheus.yml`):
// ```yaml
// scrape_configs:
//   - job_name: 'my_service'
//     static_configs:
//       - targets: ['localhost:8080']
// ```

// 2. Create dashboards using Grafana or use Prometheus's built-in expression browser

// ### Notes:
// - Use `NewConstMetric` for metrics that can't be directly instrumented
// - Protect sensitive metrics with authentication middleware
// - Consider rate limits and cardinality when using labels
// - Use `Summary` instead of `Histogram` for pre-calculated quantiles

// This service provides a foundation for monitoring Go applications with Prometheus. The metrics can be extended based on your specific monitoring needs.
