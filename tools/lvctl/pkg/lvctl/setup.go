package lvctl

import (
	"context"
	_ "embed"
	"fmt"

	templatev1 "github.com/openshift/api/template/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

//go:embed resources/elk_pod_template.yaml
var tpl string

func (lg *LogsViewer) setup() {
	lg.generateInstanceID()
	klog.Infof("setting up LogsViewer %s", lg.instanceID)

	template, err := loadTemplate()
	if err != nil {
		klog.Exit("failed to load template: ", err)
	}

	template, err = lg.createTemplateInstance(template)
	if err != nil {
		klog.Exit("failed to create template: ", err)
	}

	err = lg.instantiateTemplate(template)
	if err != nil {
		klog.Exit("failed to instantiate template: ", err)
	}

	err = lg.waitTemplateInstanceToBeReady()
	if err != nil {
		klog.Exit("failed to wait for template instance to be ready: ", err)
	}

	klog.Info("waiting for LogsViewer pod to be ready...")
	err = lg.waitForPodToBeReady()
	if err != nil {
		klog.Exit("failed to wait for pod to be ready: ", err)
	}

	route, err := lg.routeClient.RouteV1().Routes(lg.namespace).Get(context.TODO(), "logsviewer-"+lg.instanceID, metav1.GetOptions{})
	if err != nil {
		klog.Exit("failed to get route: ", err)
	}

	klog.Infof("logsViewer %s is available at %s", lg.instanceID, route.Status.Ingress[0].Host)
	klog.Infof("logsViewer %s setup complete", lg.instanceID)
}

func (lg *LogsViewer) generateInstanceID() {
	if lg.instanceID != "" {
		return
	}

	lg.instanceID = rand.String(6)
}

func (lg *LogsViewer) createTemplateInstance(template *templatev1.Template) (*templatev1.Template, error) {
	template.Name = template.Name + "-" + lg.instanceID

	template.Parameters = []templatev1.Parameter{
		{
			Name:  "SUFFIX",
			Value: lg.instanceID,
		},
		{
			Name:  "STORAGE_CLASS",
			Value: lg.storageClass,
		},
		{
			Name:  "LOGSVIEWER_IMAGE",
			Value: lg.image,
		},
	}

	template, err := lg.templateClient.TemplateV1().Templates(lg.namespace).Create(context.TODO(), template, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	klog.Infof("template '%s' created", template.Name)

	return template, nil
}

func (lg *LogsViewer) instantiateTemplate(template *templatev1.Template) error {
	ti, err := lg.templateClient.TemplateV1().TemplateInstances(lg.namespace).Create(
		context.TODO(),
		&templatev1.TemplateInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name: "logsviewer-" + lg.instanceID,
			},
			Spec: templatev1.TemplateInstanceSpec{
				Template: *template,
			},
		},
		metav1.CreateOptions{},
	)
	if err != nil {
		return err
	}

	klog.Infof("template Instance '%s' created", ti.Name)

	return nil
}

func (lg *LogsViewer) waitTemplateInstanceToBeReady() error {
	klog.Info("waiting for Template Instance to be ready...")

	watcher, err := lg.templateClient.TemplateV1().TemplateInstances(lg.namespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: "metadata.name=logsviewer-" + lg.instanceID,
	})
	if err != nil {
		return err
	}

	for event := range watcher.ResultChan() {
		ti, ok := event.Object.(*templatev1.TemplateInstance)
		if !ok {
			klog.Warning("unexpected object type")
			continue
		}

		for _, condition := range ti.Status.Conditions {
			if condition.Type == templatev1.TemplateInstanceReady {
				if condition.Status == corev1.ConditionTrue {
					klog.Info("template Instance is ready")
					return nil
				}
			}

			if condition.Type == templatev1.TemplateInstanceInstantiateFailure {
				if condition.Status == corev1.ConditionTrue {
					return fmt.Errorf("template Instance failed: %s", condition.Message)
				}
			}
		}
	}

	return nil
}

func (lg *LogsViewer) waitForPodToBeReady() error {
	watcher, err := lg.k8sClient.CoreV1().Pods(lg.namespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: "metadata.name=logsviewer-" + lg.instanceID,
	})
	if err != nil {
		return err
	}

	var phase corev1.PodPhase
	for event := range watcher.ResultChan() {
		pod, ok := event.Object.(*corev1.Pod)
		if !ok {
			klog.Warning("unexpected object type")
			continue
		}

		if pod.Status.Phase == corev1.PodRunning {
			if !logsViewerContainerReady(pod) {
				klog.Info("logsViewer container is not ready yet")
			} else {
				klog.Info("logsViewer is ready")
				break
			}
		} else if pod.Status.Phase != phase {
			klog.Infof("logsViewer pod is not ready yet, current phase: %s", pod.Status.Phase)
			phase = pod.Status.Phase
		}
	}

	return nil
}

func logsViewerContainerReady(pod *corev1.Pod) bool {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == "logsviewer" {
			return containerStatus.Ready
		}
	}

	return false
}

func loadTemplate() (*templatev1.Template, error) {
	template := &templatev1.Template{}
	err := yaml.Unmarshal([]byte(tpl), template)
	if err != nil {
		return nil, err
	}

	return template, nil
}
