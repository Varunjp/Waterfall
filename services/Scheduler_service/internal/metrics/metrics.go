package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type SchedulerMetrics struct {
	JobsAssigned prometheus.Counter
	JobsFailed   prometheus.Counter
	JobsSuccess  prometheus.Counter
	PendingJobs  *prometheus.GaugeVec
	RunningJobs  prometheus.Gauge
}

func NewMetrics() *SchedulerMetrics {
	m := &SchedulerMetrics{
		JobsAssigned: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name:"scheduler_jobs_assigned_total",
				Help: "Total jobs assigned to workers",
			},
		),
		JobsFailed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "scheduler_jobs_failed_total",
			},
		),
		JobsSuccess: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name:"scheduler_jobs_success_total",
			},
		),
		PendingJobs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "scheduler_pending_jobs",
			},
			[]string{"job_id","job_type"},
		),
		RunningJobs: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:"scheduler_running_jobs",
			},
		),
	}

	prometheus.MustRegister(
		m.JobsAssigned,
		m.JobsFailed,
		m.JobsSuccess,
		m.PendingJobs,
		m.RunningJobs,
	)

	return m 
}