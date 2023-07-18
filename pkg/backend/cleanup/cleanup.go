package cleanup

import (
	"time"

	clientset "github.com/openshift/client-go/template/clientset/versioned"
	"k8s.io/client-go/rest"

	"logsviewer/pkg/backend/log"
)

type Cleanup struct {
	instance  string
	namespace string
	client    *clientset.Clientset
}

func StartCleanupJob(instance, namespace string) {
	if instance == "" || namespace == "" {
		log.Log.Println("Instance and namespace must be set (POD_NAME and POD_NAMESPACE env vars)")
		return
	}

	cleanup := newCleanup(instance, namespace)

	timer := time.NewTimer(1 * time.Hour)
	for {
		<-timer.C
		cleanup.run()
		timer.Reset(1 * time.Hour)
	}
}

func newCleanup(instance, namespace string) *Cleanup {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	client, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return &Cleanup{
		instance:  instance,
		namespace: namespace,
		client:    client,
	}
}
