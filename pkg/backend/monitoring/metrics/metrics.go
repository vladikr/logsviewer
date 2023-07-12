package metrics

import (
	"github.com/machadovilaca/operator-observability/pkg/operatormetrics"
	"logsviewer/pkg/backend/env"
)

const metricPrefix = "logsviewer_"

var (
	metrics = [][]operatormetrics.Metric{
		instanceMetrics,
		usageMetrics,
	}

	constLabels = map[string]string{
		"instance": env.GetEnv("POD_NAME", "unknown"),
	}
)

func SetupMetrics() {
	err := operatormetrics.RegisterMetrics(metrics...)
	if err != nil {
		panic(err)
	}
}

func ListMetrics() []operatormetrics.Metric {
	return operatormetrics.ListMetrics()
}
