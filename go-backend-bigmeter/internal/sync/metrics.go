package sync

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	syncDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sync_job_duration_seconds",
			Help:    "Duration of sync jobs (per branch)",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"job", "branch", "status"},
	)

	syncRows = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_rows_total",
			Help: "Rows processed by sync jobs (upserted/zeroed)",
		},
		[]string{"job", "branch", "type"},
	)

	syncBatches = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_batches_total",
			Help: "Number of batches processed by sync jobs",
		},
		[]string{"job", "branch"},
	)
)

func observeJob(job, branch, status string, start time.Time) {
	syncDuration.WithLabelValues(job, branch, status).Observe(time.Since(start).Seconds())
}

func addRows(job, branch, typ string, n int) {
	if n <= 0 {
		return
	}
	syncRows.WithLabelValues(job, branch, typ).Add(float64(n))
}

func incBatches(job, branch string, n int) {
	if n <= 0 {
		return
	}
	syncBatches.WithLabelValues(job, branch).Add(float64(n))
}
