package lvctl

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (lg *LogsViewer) delete() {
	if lg.instanceID == "" {
		klog.Exit("instance ID must be specified")
	}

	err := lg.templateClient.TemplateV1().TemplateInstances(lg.namespace).Delete(context.TODO(), "logsviewer-"+lg.instanceID, metav1.DeleteOptions{})
	if err != nil {
		klog.Exit("failed to delete template instance: ", err)
	}
	klog.Infof("deleted template instance logsviewer-%s", lg.instanceID)

	err = lg.templateClient.TemplateV1().Templates(lg.namespace).Delete(context.TODO(), "logsviewer-"+lg.instanceID, metav1.DeleteOptions{})
	if err != nil {
		klog.Exit("failed to delete template: ", err)
	}
	klog.Infof("deleted template logsviewer-%s", lg.instanceID)
}
