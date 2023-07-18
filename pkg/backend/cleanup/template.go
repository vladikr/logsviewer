package cleanup

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/openshift/api/template/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"logsviewer/pkg/backend/log"
	"logsviewer/pkg/backend/monitoring/metrics"
)

func (c *Cleanup) run() {
	log.Log.Println("running cleanup job")

	template, err := c.getTemplate()
	if err != nil {
		log.Log.Println("failed to get template", "error", err)
		return
	}

	err = c.handleCleanup(template)
	if err != nil {
		log.Log.Println("failed to handle cleanup", "error", err)
		return
	}
}

func (c *Cleanup) getTemplate() (*v1.Template, error) {
	return c.client.TemplateV1().Templates(c.namespace).Get(context.Background(), c.instance, metav1.GetOptions{})
}

func (c *Cleanup) handleCleanup(template *v1.Template) error {
	deletionCondition, err := c.getDeletionCondition(template)
	if err != nil {
		return err
	}

	if deletionCondition == Never {
		log.Log.Println("deletion condition is never, skipping cleanup")
		return nil
	}

	rawDeletionDelay, err := c.getDeletionDelay(template)
	if err != nil {
		return err
	}

	deletionDelay, err := time.ParseDuration(rawDeletionDelay)
	if err != nil {
		return err
	}

	log.Log.Println("deletion condition is", "condition", deletionCondition, "delay", deletionDelay)
	return c.deleteIfMatches(template, deletionCondition, deletionDelay)
}

func (c *Cleanup) getDeletionCondition(template *v1.Template) (DeletionCondition, error) {
	if deletionCondition, ok := template.Labels[DeletionConditionLabel]; ok {
		if deletionCondition != string(Creation) && deletionCondition != string(LastMustGatherUpload) && deletionCondition != string(Never) {
			return "", fmt.Errorf("deletion condition label has invalid value: %s", deletionCondition)
		}
		return DeletionCondition(deletionCondition), nil
	} else {
		return "", fmt.Errorf("deletion condition label not found")
	}
}

func (c *Cleanup) getDeletionDelay(template *v1.Template) (string, error) {
	if deletionDelay, ok := template.Labels[DeletionDelayLabel]; ok {
		return deletionDelay, nil
	} else {
		return "", fmt.Errorf("deletion delay label not found")
	}
}

func (c *Cleanup) deleteIfMatches(template *v1.Template, deletionCondition DeletionCondition, deletionDelay time.Duration) error {
	if deletionCondition == Creation {
		return c.deleteIfCreation(template, deletionDelay)
	} else if deletionCondition == LastMustGatherUpload {
		return c.deleteIfLastMustGatherUpload(template, deletionDelay)
	}
	return nil
}

func (c *Cleanup) deleteIfCreation(template *v1.Template, deletionDelay time.Duration) error {
	creationTime := template.CreationTimestamp.Time
	now := time.Now()
	if now.Sub(creationTime) < deletionDelay {
		log.Log.Println("template is too young, skipping cleanup", "creationTime", creationTime, "now", now)
		return nil
	}

	return c.deleteTemplate(template)
}

func (c *Cleanup) deleteIfLastMustGatherUpload(template *v1.Template, deletionDelay time.Duration) error {
	lastMustGatherUploadTimestamp, err := metrics.GetLastMustGatherUploadTimestamp()
	if err != nil {
		if err == metrics.ErrNoMustGatherUploads {
			log.Log.Println("no must-gather uploads found, cleaning up based on creation time")
			return c.deleteIfCreation(template, deletionDelay)
		}
		return err
	}

	now := time.Now()
	if now.Sub(lastMustGatherUploadTimestamp) < deletionDelay {
		log.Log.Println("last must-gather upload was too recent, skipping cleanup", "lastMustGatherUploadTimestamp", lastMustGatherUploadTimestamp, "now", now)
		return nil
	}

	return c.deleteTemplate(template)
}

func (c *Cleanup) deleteTemplate(template *v1.Template) error {
	err := c.client.TemplateV1().Templates(c.namespace).Delete(context.Background(), template.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = c.client.TemplateV1().TemplateInstances(c.namespace).Delete(context.Background(), template.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	log.Log.Println("template deleted", "name", template.Name)
	return nil
}
