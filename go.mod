module logsviewer

go 1.16

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/gorilla/websocket v1.5.0
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v0.23.1
	k8s.io/klog/v2 v2.40.1 // indirect
	sigs.k8s.io/yaml v1.3.0
)

replace (
	github.com/openshift/api => github.com/openshift/api v0.0.0-20191219222812-2987a591a72c
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20210112165513-ebc401615f47
	github.com/operator-framework/operator-lifecycle-manager => github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190128024246-5eb7ae5bdb7a
	k8s.io/kubectl => k8s.io/kubectl v0.23.1
	kubevirt.io/api => kubevirt.io/api v0.56.0
	kubevirt.io/client-go => kubevirt.io/client-go v0.56.0
	kubevirt.io/containerized-data-importer => kubevirt.io/containerized-data-importer v1.41.0
	kubevirt.io/containerized-data-importer-api => kubevirt.io/containerized-data-importer-api v1.54.0
)

exclude k8s.io/cluster-bootstrap v0.0.0
