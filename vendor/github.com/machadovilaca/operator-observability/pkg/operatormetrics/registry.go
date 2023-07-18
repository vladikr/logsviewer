package operatormetrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var operatorRegistry = newRegistry()

type operatorRegisterer struct {
	registeredMetrics    map[string]Metric
	registeredCollectors map[string]registeredCollector
}

type registeredCollector struct {
	desc      *prometheus.Desc
	metric    Metric
	valueType prometheus.ValueType
}

func newRegistry() operatorRegisterer {
	return operatorRegisterer{
		registeredMetrics:    map[string]Metric{},
		registeredCollectors: map[string]registeredCollector{},
	}
}

// RegisterMetrics registers the metrics with the Prometheus registry.
func RegisterMetrics(allMetrics ...[]Metric) error {
	for _, metricList := range allMetrics {
		for _, metric := range metricList {
			err := prometheus.Register(metric.getCollector())
			if err != nil {
				return err
			}
			operatorRegistry.registeredMetrics[metric.GetOpts().Name] = metric
		}
	}

	return nil
}

// RegisterCollector registers the collector with the Prometheus registry.
func RegisterCollector(collectors ...Collector) error {
	for _, collector := range collectors {
		for _, metric := range collector.Metrics {
			err := createCollectorMetric(metric)
			if err != nil {
				return err
			}
		}

		err := prometheus.Register(collector)
		if err != nil {
			return err
		}
	}

	return nil
}

func createCollectorMetric(metric CollectorMetric) error {
	opts := metric.GetOpts()
	mType := metric.GetType()
	var valueType prometheus.ValueType

	switch mType {
	case CounterType:
		valueType = prometheus.CounterValue
	case GaugeType:
		valueType = prometheus.GaugeValue
	default:
		return fmt.Errorf("collector metric %q has invalid metric type: %q", opts.Name, mType)
	}

	operatorRegistry.registeredCollectors[opts.Name] = registeredCollector{
		desc:      prometheus.NewDesc(opts.Name, opts.Help, metric.Labels, opts.ConstLabels),
		metric:    metric,
		valueType: valueType,
	}

	return nil
}

// ListMetrics returns a list of all registered metrics.
func ListMetrics() []Metric {
	var result []Metric

	for _, rm := range operatorRegistry.registeredMetrics {
		result = append(result, rm)
	}

	for _, rc := range operatorRegistry.registeredCollectors {
		result = append(result, rc.metric)
	}

	return result
}
