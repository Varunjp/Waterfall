package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(body)
}

func (m *AdminMetrics) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}

		m.APIHealth.RequestsInFlight.Inc()
		defer m.APIHealth.RequestsInFlight.Dec()

		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = r.URL.Path
		}
		code := strconv.Itoa(status)
		result := resultFromHTTPStatus(status)

		m.APIHealth.RequestsTotal.WithLabelValues("http", r.Method+" "+route, code).Inc()
		m.APIHealth.RequestDuration.WithLabelValues("http", r.Method+" "+route, code).Observe(time.Since(start).Seconds())

		if operation := businessOperationForHTTPRoute(route); operation != "" {
			m.BusinessOperations.OperationsTotal.WithLabelValues(operation, "http", result).Inc()
			m.BusinessOperations.OperationDuration.WithLabelValues(operation, "http", result).Observe(time.Since(start).Seconds())
		}

		m.recordHTTPSecurityEvent(route, status)
	})
}

func resultFromHTTPStatus(status int) string {
	if status < http.StatusBadRequest {
		return "success"
	}
	return "error"
}

func (m *AdminMetrics) recordHTTPSecurityEvent(route string, status int) {
	if status == http.StatusUnauthorized {
		m.AuthenticationSecurity.EventsTotal.WithLabelValues("http", route, "authentication", "denied").Inc()
		return
	}
	if status == http.StatusForbidden {
		m.AuthenticationSecurity.EventsTotal.WithLabelValues("http", route, "authorization", "denied").Inc()
	}
}

func businessOperationForHTTPRoute(route string) string {
	switch route {
	case "/billing/checkout", "/billing/cancel", "/billing/subscription", "/billing/webhook":
		return "billing"
	default:
		return ""
	}
}
