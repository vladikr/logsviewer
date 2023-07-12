package metrics

import (
	"time"

	"github.com/machadovilaca/operator-observability/pkg/operatormetrics"
)

var (
	usageMetrics = []operatormetrics.Metric{
		mustGatherUploadTotal,
		lastMustGatherUploadTimestamp,
	}

	mustGatherUploadTotal = operatormetrics.NewCounter(
		operatormetrics.MetricOpts{
			Name:        metricPrefix + "must_gather_upload_total",
			Help:        "Number of must-gather uploads",
			ConstLabels: constLabels,
		},
	)

	lastMustGatherUploadTimestamp = operatormetrics.NewGauge(
		operatormetrics.MetricOpts{
			Name:        metricPrefix + "last_must_gather_upload_timestamp",
			Help:        "Timestamp of the last must-gather upload",
			ConstLabels: constLabels,
		},
	)
)

func NewMustGatherUploaded() {
	mustGatherUploadTotal.Inc()
	lastMustGatherUploadTimestamp.Set(float64(time.Now().Unix()))
}
