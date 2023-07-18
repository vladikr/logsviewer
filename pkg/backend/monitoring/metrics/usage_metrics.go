package metrics

import (
	"errors"
	"time"

	"github.com/machadovilaca/operator-observability/pkg/operatormetrics"
	ioprometheusclient "github.com/prometheus/client_model/go"
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

	ErrNoMustGatherUploads = errors.New("no must-gather uploads")
)

func NewMustGatherUploaded() {
	mustGatherUploadTotal.Inc()
	lastMustGatherUploadTimestamp.Set(float64(time.Now().Unix()))
}

func GetLastMustGatherUploadTimestamp() (time.Time, error) {
	dto := &ioprometheusclient.Metric{}
	err := lastMustGatherUploadTimestamp.Write(dto)
	if err != nil {
		return time.Time{}, err
	}

	value := int64(dto.GetGauge().GetValue())
	if value == 0 {
		return time.Time{}, ErrNoMustGatherUploads
	}

	return time.Unix(value, 0), nil
}
