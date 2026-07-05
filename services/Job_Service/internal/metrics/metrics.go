package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

type JobMetrics struct {
	JobsCreatedTotal         prometheus.Counter
	JobsCreationFailedTotal  prometheus.Counter
	JobCreateDurationSeconds prometheus.Histogram
	JobPayloadSizeBytes      prometheus.Histogram
	KafkaPublishTotal        prometheus.Counter
	KafkaPublishFailedTotal  prometheus.Counter
	ActiveRequests           prometheus.Gauge
}

func NewMetrics() *JobMetrics {
	m := &JobMetrics{
		JobsCreatedTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "jobs_created_total",
				Help: "Total jobs successfully created by the Job Service.",
			},
		),
		JobsCreationFailedTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "jobs_creation_failed_total",
				Help: "Total job creation attempts that failed.",
			},
		),
		JobCreateDurationSeconds: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "job_create_duration_seconds",
				Help:    "Duration of job creation operations.",
				Buckets: prometheus.DefBuckets,
			},
		),
		JobPayloadSizeBytes: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "job_payload_size_bytes",
				Help:    "Size of job payloads in bytes.",
				Buckets: []float64{64, 128, 256, 512, 1024, 2048, 4096, 8192},
			},
		),
		KafkaPublishTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "kafka_publish_total",
				Help: "Total Kafka publishes successfully completed.",
			},
		),
		KafkaPublishFailedTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "kafka_publish_failed_total",
				Help: "Total Kafka publishes that failed.",
			},
		),
		ActiveRequests: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_requests",
				Help: "Number of active requests being processed.",
			},
		),
	}

	prometheus.MustRegister(
		m.JobsCreatedTotal,
		m.JobsCreationFailedTotal,
		m.JobCreateDurationSeconds,
		m.JobPayloadSizeBytes,
		m.KafkaPublishTotal,
		m.KafkaPublishFailedTotal,
		m.ActiveRequests,
	)

	return m
}

func (m *JobMetrics) RecordJobCreation(size int, succeeded bool) {
	if succeeded {
		m.JobsCreatedTotal.Inc()
	} else {
		m.JobsCreationFailedTotal.Inc()
	}
	m.ObservePayloadSize(size)
}

func (m *JobMetrics) RecordKafkaPublish(succeeded bool) {
	if succeeded {
		m.KafkaPublishTotal.Inc()
	} else {
		m.KafkaPublishFailedTotal.Inc()
	}
}

func (m *JobMetrics) RecordActiveRequest(delta float64) {
	m.ActiveRequests.Add(delta)
}

func (m *JobMetrics) ObserveCreateDuration(d time.Duration) {
	m.JobCreateDurationSeconds.Observe(d.Seconds())
}

func (m *JobMetrics) ObservePayloadSize(size int) {
	m.JobPayloadSizeBytes.Observe(float64(size))
}

func RunServer(ctx context.Context, addr string, logger any) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{Addr: addr, Handler: mux}
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	return server.ListenAndServe()
}

func (m *JobMetrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		m.RecordActiveRequest(1)
		defer m.RecordActiveRequest(-1)
		return handler(ctx, req)
	}
}
