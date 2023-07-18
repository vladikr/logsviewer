package operatorrules

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/hashicorp/go-multierror"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func AddToScheme(scheme *runtime.Scheme) error {
	var result error

	result = multierror.Append(result, promv1.AddToScheme(scheme))
	result = multierror.Append(result, rbacv1.AddToScheme(scheme))

	return result
}
