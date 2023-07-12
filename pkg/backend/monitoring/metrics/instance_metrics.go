package metrics

import (
	"time"

	"github.com/machadovilaca/operator-observability/pkg/operatormetrics"
)

var (
	instanceMetrics = []operatormetrics.Metric{
		creationTimestamp,
	}

	creationTimestamp = operatormetrics.NewGauge(
		operatormetrics.MetricOpts{
			Name:        metricPrefix + "creation_timestamp",
			Help:        "Timestamp of the instance creation",
			ConstLabels: constLabels,
		},
	)
)

func InstanceCreated() {
	creationTimestamp.Set(float64(time.Now().Unix()))
}
