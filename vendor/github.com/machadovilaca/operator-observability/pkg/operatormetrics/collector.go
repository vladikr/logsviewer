package operatormetrics

import "github.com/prometheus/client_golang/prometheus"

// Collector registers a prometheus.Collector with a set of metrics in the
// Prometheus registry. The metrics are collected by calling the CollectCallback
// function.
type Collector struct {
	// Metrics is a list of metrics to be collected by the collector.
	Metrics []CollectorMetric

	// CollectCallback is a function that returns a list of CollectionResults.
	// The CollectionResults are used to populate the metrics in the collector.
	CollectCallback func() []CollectorResult
}

type CollectorMetric struct {
	Metric
	Labels []string
}

type CollectorResult struct {
	Metric Metric
	Labels []string
	Value  float64
}

func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, cm := range c.Metrics {
		rc, ok := operatorRegistry.registeredCollectors[cm.Metric.GetOpts().Name]
		if !ok {
			continue
		}
		ch <- rc.desc
	}
}

func (c Collector) Collect(ch chan<- prometheus.Metric) {
	collectedMetrics := c.CollectCallback()

	for _, cm := range collectedMetrics {
		rc, ok := operatorRegistry.registeredCollectors[cm.Metric.GetOpts().Name]
		if !ok {
			continue
		}

		mv, _ := prometheus.NewConstMetric(rc.desc, rc.valueType, cm.Value, cm.Labels...)
		ch <- mv
	}
}
