package operatorrules

import (
	"fmt"

	"github.com/grafana/regexp"
	"github.com/hashicorp/go-multierror"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

var (
	recordingRuleValidator = defaultRecordingRuleValidation
	alertValidator         = defaultAlertValidation
)

// SetRecordingRuleValidator sets the validator for recording rules.
func SetRecordingRuleValidator(validator func(recordingRule *RecordingRule) error) {
	recordingRuleValidator = validator
}

// SetAlertValidator sets the validator for alerts.
func SetAlertValidator(validator func(alert *promv1.Rule) error) {
	alertValidator = validator
}

func defaultRecordingRuleValidation(recordingRule *RecordingRule) error {
	var result *multierror.Error

	if recordingRule.MetricsOpts.Name == "" {
		result = multierror.Append(result, fmt.Errorf("recording rule must have a name"))
	}

	if recordingRule.Expr.StrVal == "" {
		result = multierror.Append(result, fmt.Errorf("recording rule must have an expression"))
	}

	return result.ErrorOrNil()
}

// based on https://sdk.operatorframework.io/docs/best-practices/observability-best-practices/#alerts-style-guide
func defaultAlertValidation(alert *promv1.Rule) error {
	var result *multierror.Error

	if alert.Alert == "" || !isPascalCase(alert.Alert) {
		result = multierror.Append(result, fmt.Errorf("alert must have a name in PascalCase format"))
	}

	if alert.Expr.StrVal == "" {
		result = multierror.Append(result, fmt.Errorf("alert must have an expression"))
	}

	// Alerts MUST include a severity label indicating the alert’s urgency.
	result = multierror.Append(result, validateLabels(alert))

	// Alerts MUST include summary and description annotations.
	result = multierror.Append(result, validateAnnotations(alert))

	return result.ErrorOrNil()
}

func isPascalCase(s string) bool {
	pascalCasePattern := `^[A-Z][a-zA-Z0-9]*(?:[A-Z][a-zA-Z0-9]*)*$`
	pascalCaseRegex := regexp.MustCompile(pascalCasePattern)
	return pascalCaseRegex.MatchString(s)
}

func validateLabels(alert *promv1.Rule) error {
	severity := alert.Labels["severity"]
	if severity == "" || (severity != "critical" && severity != "warning" && severity != "info") {
		return fmt.Errorf("alert must have a severity label with value critical, warning, or info")
	}

	return nil
}

func validateAnnotations(alert *promv1.Rule) error {
	var result error

	summary := alert.Annotations["summary"]
	if summary == "" {
		result = multierror.Append(result, fmt.Errorf("alert must have a summary annotation"))
	}

	description := alert.Annotations["description"]
	if description == "" {
		result = multierror.Append(result, fmt.Errorf("alert must have a description annotation"))
	}

	return result
}
