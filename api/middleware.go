package api

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type RequestMetrics struct {
	mu          sync.Mutex
	Total       int64
	Success     int64
	Failures    int64
	ByRoute     map[string]int64
	LastRequest time.Time
}

func NewRequestMetrics() *RequestMetrics { return &RequestMetrics{ByRoute: map[string]int64{}} }

func (m *RequestMetrics) Observe(route string, status int, at time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Total++
	m.LastRequest = at
	if status >= 200 && status < 400 {
		m.Success++
	} else {
		m.Failures++
	}
	m.ByRoute[route]++
}

func (m *RequestMetrics) Snapshot() (int64, int64, int64, map[string]int64, time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	routes := map[string]int64{}
	for route, count := range m.ByRoute {
		routes[route] = count
	}
	return m.Total, m.Success, m.Failures, routes, m.LastRequest
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *statusWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(data)
}

func Instrument(next http.Handler, metrics *RequestMetrics) http.Handler {
	if metrics == nil {
		metrics = NewRequestMetrics()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		wrapped := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(w, r)
		route := r.URL.Path
		if strings.TrimSpace(route) == "" {
			route = "/"
		}
		metrics.Observe(route, wrapped.status, started)
		_ = started
	})
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Actor-ID, X-Actor-Role")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequestID(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if value == "" {
		value = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return value
}

func LimitBody(next http.Handler, max int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if max > 0 && r.ContentLength > max {
			writeJSON(w, http.StatusRequestEntityTooLarge, errorResponse{Error: "request body too large", Code: "body_limit"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
