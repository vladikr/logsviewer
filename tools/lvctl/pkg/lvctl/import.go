package lvctl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (lg *LogsViewer) importMustGather() {
	err := lg.validateImportParams()
	if err != nil {
		klog.Exit("failed to validate import params: ", err)
	}

	klog.Infof("importing must-gather from file '%s'", lg.mustGatherFileName)

	err = lg.checkPodExistsAndIsReady()
	if err != nil {
		klog.Exit("failed to check pod: ", err)
	}

	klog.Info("importing must-gather file to LogsViewer...")
	err = lg.importMustGatherFileToLogsviewer()
	if err != nil {
		klog.Exit("failed to import must-gather file: ", err)
	}
	klog.Info("importing must-gather file to LogsViewer complete")
}

func (lg *LogsViewer) validateImportParams() error {
	if lg.instanceID == "" {
		return errors.New("instance ID must be specified")
	}

	if lg.mustGatherFileName == "" {
		return errors.New("must gather file name must be specified")
	}

	if !strings.HasSuffix(lg.mustGatherFileName, ".tar.gz") {
		return errors.New("must gather file name must be a .tar.gz file")
	}

	return nil
}

func (lg *LogsViewer) checkPodExistsAndIsReady() error {
	_, err := lg.k8sClient.CoreV1().Pods(lg.namespace).Get(context.TODO(), "logsviewer-"+lg.instanceID, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return lg.waitForPodToBeReady()
}

func (lg *LogsViewer) importMustGatherFileToLogsviewer() error {
	route, err := lg.routeClient.RouteV1().Routes(lg.namespace).Get(context.TODO(), "logsviewer-"+lg.instanceID, metav1.GetOptions{})
	if err != nil {
		klog.Exit("failed to get route: ", err)
	}
	url := "http://" + route.Status.Ingress[0].Host + "/uploadLogs"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	err = lg.loadMustGatherFile(writer)
	if err != nil {
		return err
	}

	return uploadMustGather(url, body, writer)
}

func (lg *LogsViewer) loadMustGatherFile(writer *multipart.Writer) error {
	file, err := os.Open(lg.mustGatherFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, file.Name()))
	h.Set("Content-Type", "application/gzip")
	part, err := writer.CreatePart(h)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	return nil
}

func uploadMustGather(url string, body *bytes.Buffer, writer *multipart.Writer) error {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)
		return fmt.Errorf("unexpected %d status code: %s", res.StatusCode, bodyString)
	}

	return nil
}
