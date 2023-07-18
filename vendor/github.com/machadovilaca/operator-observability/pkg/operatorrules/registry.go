package operatorrules

import (
	"fmt"

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

var operatorRegistry = newRegistry()

type operatorRegisterer struct {
	registeredRecordingRules []RecordingRule
	registeredAlerts         []promv1.Rule
}

func newRegistry() operatorRegisterer {
	return operatorRegisterer{
		registeredRecordingRules: []RecordingRule{},
	}
}

// RegisterRecordingRules registers the given recording rules.
func RegisterRecordingRules(recordingRules ...[]RecordingRule) error {
	for _, recordingRuleList := range recordingRules {
		for _, recordingRule := range recordingRuleList {
			if err := recordingRuleValidator(&recordingRule); err != nil {
				return fmt.Errorf("invalid recording rule %s: %w", recordingRule.MetricsOpts.Name, err)
			}

			operatorRegistry.registeredRecordingRules = append(operatorRegistry.registeredRecordingRules, recordingRule)
		}
	}

	return nil
}

// RegisterAlerts registers the given alerts.
func RegisterAlerts(alerts ...[]promv1.Rule) error {
	for _, alertList := range alerts {
		for _, alert := range alertList {
			if err := alertValidator(&alert); err != nil {
				return fmt.Errorf("invalid alert %s: %w", alert.Alert, err)
			}

			operatorRegistry.registeredAlerts = append(operatorRegistry.registeredAlerts, alert)
		}
	}

	return nil
}

// ListRecordingRules returns the registered recording rules.
func ListRecordingRules() []RecordingRule {
	return operatorRegistry.registeredRecordingRules
}

// ListAlerts returns the registered alerts.
func ListAlerts() []promv1.Rule {
	return operatorRegistry.registeredAlerts
}
