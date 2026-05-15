package metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type AdminMetrics struct {
	APIHealth              APIHealthMetrics
	AuthenticationSecurity AuthenticationSecurityMetrics
	BusinessOperations     BusinessOperationMetrics
	DatabasePerformance    DatabasePerformanceMetrics
}

type APIHealthMetrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	RequestsInFlight prometheus.Gauge
}

type AuthenticationSecurityMetrics struct {
	EventsTotal *prometheus.CounterVec
}

type BusinessOperationMetrics struct {
	OperationsTotal   *prometheus.CounterVec
	OperationDuration *prometheus.HistogramVec
}

type DatabasePerformanceMetrics struct {
	OpenConnections prometheus.Collector
	InUse           prometheus.Collector
	Idle            prometheus.Collector
	WaitCount       prometheus.Collector
	WaitDuration    prometheus.Collector
}

func NewMetrics(db *sql.DB) *AdminMetrics {
	m := &AdminMetrics{
		APIHealth:              newAPIHealthMetrics(),
		AuthenticationSecurity: newAuthenticationSecurityMetrics(),
		BusinessOperations:     newBusinessOperationMetrics(),
		DatabasePerformance:    newDatabasePerformanceMetrics(db),
	}

	prometheus.MustRegister(
		m.APIHealth.RequestsTotal,
		m.APIHealth.RequestDuration,
		m.APIHealth.RequestsInFlight,
		m.AuthenticationSecurity.EventsTotal,
		m.BusinessOperations.OperationsTotal,
		m.BusinessOperations.OperationDuration,
		m.DatabasePerformance.OpenConnections,
		m.DatabasePerformance.InUse,
		m.DatabasePerformance.Idle,
		m.DatabasePerformance.WaitCount,
		m.DatabasePerformance.WaitDuration,
	)

	return m
}

func newAPIHealthMetrics() APIHealthMetrics {
	return APIHealthMetrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "admin_api_requests_total",
				Help: "Total Admin API requests by transport, endpoint, and status.",
			},
			[]string{"transport", "endpoint", "code"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "admin_api_request_duration_seconds",
				Help:    "Admin API request duration by transport, endpoint, and status.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"transport", "endpoint", "code"},
		),
		RequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "admin_api_requests_in_flight",
				Help: "Current Admin API requests being handled.",
			},
		),
	}
}

func newAuthenticationSecurityMetrics() AuthenticationSecurityMetrics {
	return AuthenticationSecurityMetrics{
		EventsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "admin_auth_security_events_total",
				Help: "Authentication and authorization events handled by Admin service.",
			},
			[]string{"transport", "endpoint", "event", "result"},
		),
	}
}

func newBusinessOperationMetrics() BusinessOperationMetrics {
	return BusinessOperationMetrics{
		OperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "admin_business_operations_total",
				Help: "Business operations handled by Admin service.",
			},
			[]string{"operation", "transport", "result"},
		),
		OperationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "admin_business_operation_duration_seconds",
				Help:    "Business operation duration by Admin operation and result.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "transport", "result"},
		),
	}
}

func newDatabasePerformanceMetrics(db *sql.DB) DatabasePerformanceMetrics {
	return DatabasePerformanceMetrics{
		OpenConnections: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "admin_db_open_connections",
				Help: "Current number of established database connections.",
			},
			func() float64 { return float64(db.Stats().OpenConnections) },
		),
		InUse: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "admin_db_in_use_connections",
				Help: "Current number of database connections in use.",
			},
			func() float64 { return float64(db.Stats().InUse) },
		),
		Idle: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "admin_db_idle_connections",
				Help: "Current number of idle database connections.",
			},
			func() float64 { return float64(db.Stats().Idle) },
		),
		WaitCount: prometheus.NewCounterFunc(
			prometheus.CounterOpts{
				Name: "admin_db_connection_wait_total",
				Help: "Total times database requests waited for a connection.",
			},
			func() float64 { return float64(db.Stats().WaitCount) },
		),
		WaitDuration: prometheus.NewCounterFunc(
			prometheus.CounterOpts{
				Name: "admin_db_connection_wait_duration_seconds_total",
				Help: "Total time spent waiting for database connections.",
			},
			func() float64 { return db.Stats().WaitDuration.Seconds() },
		),
	}
}
