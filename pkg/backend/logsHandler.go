package backend

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	yamlv3 "gopkg.in/yaml.v3"
	k8sv1 "k8s.io/api/core/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"logsviewer/pkg/backend/db"
	"logsviewer/pkg/backend/log"
)

type Pods struct {
	Metadata Metadata `yaml:"metadata"`
	Spec     Spec     `yaml:"spec"`
	Status   Status   `yaml:"status"`
}

type Metadata struct {
	Namespace       string           `yaml:"namespace,omitempty"`
	Name            string           `yaml:"name,omitempty"`
	OwnerReferences []OwnerReference `yaml:"ownerReferences,omitempty"`
	UID             string           `yaml:"uid,omitempty"`
}

type Spec struct {
	NodeName string `yaml:"nodeName,omitempty"`
}

type Status struct {
	HostIP string `yaml:"hostIP,omitempty"`
}

type OwnerReference struct {
	UID string
}

type EnrichmentData struct {
	HostName        string   `json:"host.name"`
	HostIP          string   `json:"host.ip"`
	UID             string   `json:"pod.uid"`
	OwnerReferences []string `json:"pod.ownerReferences,omitempty"`
}

type logsHandler struct {
	handlerLock sync.Mutex
	stopCh      chan struct{}
	objectStore *db.ObjectStore
	lookupData  map[string]EnrichmentData
}

func NewLogsHandler(storeDB *db.DatabaseInstance) *logsHandler {
	lookupData := make(map[string]EnrichmentData)
	stopCh := make(chan struct{}, 1)
	objStore := db.NewObjectStore(storeDB)

	go objStore.Run(1, stopCh)

	return &logsHandler{
		lookupData:  lookupData,
		objectStore: objStore,
		stopCh:      stopCh,
	}
}

func handleTarGz(srcFile string, targetPath string) error {
	if err := unTarGz(srcFile, targetPath); err != nil {
		return err
	}
	// delete source file
	if err := os.Remove(srcFile); err != nil {
		log.Log.Fatalln("failed to delete file ", srcFile, " - ", err)
	}
	log.Log.Println("removed file: ", srcFile)
	return nil
}

func unTarGz(srcFile string, targetPath string) error {
	var namespacePrefixPath []string

	gzipStream, err := os.Open(srcFile)
	defer gzipStream.Close()

	if err != nil {
		log.Log.Fatalln("failed to open file ", srcFile, " - ", err)
		return err
	}

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Log.Fatalln("failed create gzip stream  - ", err)
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Log.Fatalln("failed to get next file in tar  - ", err)
			return err
		}

		if strings.HasSuffix(header.Name, "/timestamp") {
			newTarget := filepath.Join(targetPath, "timestamp")
			err = os.MkdirAll(filepath.Dir(newTarget), os.ModePerm)
			if err != nil {
				log.Log.Fatalln("failed to create dir ", targetPath, " - ", err)
				return err
			}

			log.Log.Println("newTarget: ", newTarget)
			if _, err := os.Stat(newTarget); err == nil {
				log.Log.Println("file already exist, skip")
				continue
			}

			outFile, err := os.Create(newTarget)
			if err != nil {
				log.Log.Fatalln("failed create target ", newTarget, " - ", err)
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Log.Fatalln("failed to copy from src to target ", newTarget, " - ", err)
				return err
			}
			log.Log.Println("created file: ", newTarget)
			outFile.Close()
			continue
		}

		// extract only the namespaces dir
		if !strings.Contains(header.Name, "namespaces/") && !strings.Contains(header.Name, "cluster-scoped-resources/") {
			continue
		}

		if len(namespacePrefixPath) == 0 {

			log.Log.Println("Header name: ", header.Name)
			// find path to the namespaces directory
			sp := strings.Split(header.Name, "/")
			for _, ps := range sp {
				if ps == "namespaces" || ps == "cluster-scoped-resources" {
					break
				}
				namespacePrefixPath = append(namespacePrefixPath, ps)
				log.Log.Println("current pathToNamespacesDir: ", strings.Join(namespacePrefixPath[:], "/"))
			}
		}
		pathToNamespacesDir := strings.Join(namespacePrefixPath[:], "/")
		workingHeaderName := strings.TrimPrefix(header.Name, pathToNamespacesDir)
		newTarget := filepath.Join(targetPath, workingHeaderName)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(newTarget, 0755); err != nil {
				log.Log.Fatalln("failed to create dir  - ", err)
				return err
			}
		case tar.TypeReg:
			// if such file already exist, add/increase its version
			if _, err := os.Stat(newTarget); err == nil {

				ext := filepath.Ext(newTarget)
				filenameBase := strings.TrimSuffix(newTarget, ext)
				sp := strings.Split(filenameBase, "_")
				suffixIndexStr := sp[len(sp)-1]
				suffixIndex, err := strconv.Atoi(suffixIndexStr)
				if err != nil {
					filenameBase += "_1"
				} else {
					fileN := strings.TrimSuffix(filenameBase, fmt.Sprintf("_%d", suffixIndex))
					suffixIndex += 1
					filenameBase = fmt.Sprintf("%s_%d", fileN, suffixIndex)
				}
				newTarget = fmt.Sprintf("%s%s", filenameBase, ext)
			}
			outFile, err := os.Create(newTarget)

			if err != nil {
				log.Log.Fatalln("failed create target ", newTarget, " - ", err)
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Log.Fatalln("failed to copy from src to target ", newTarget, " - ", err)
				return err
			}
			outFile.Close()

		default:
			log.Log.Fatalf(
				"uknown header type: %s in %s",
				header.Typeflag,
				header.Name)
			return err
		}

	}
	log.Log.Println("Extracted file: ", srcFile)
	return nil
}

func (l *logsHandler) loadExistingEnrichmentData() error {
	// read the existing enrichment data file
	jsonFile, err := os.Open(ENRICHMENT_DATA_FILE)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	json.Unmarshal([]byte(byteValue), &l.lookupData)

	return nil
}

func (l *logsHandler) processEnrichmentData(pod *Pods) {
	// generate enrichment data
	key := fmt.Sprintf("%s/%s", pod.Metadata.Namespace, pod.Metadata.Name)

	enrichmentData := EnrichmentData{
		HostName: pod.Spec.NodeName,
		HostIP:   pod.Status.HostIP,
		UID:      pod.Metadata.UID,
	}
	if pod.Metadata.OwnerReferences != nil {
		for _, ref := range pod.Metadata.OwnerReferences {
			enrichmentData.OwnerReferences = append(enrichmentData.OwnerReferences, ref.UID)
		}
	}
	l.lookupData[key] = enrichmentData

	return
}

func (l *logsHandler) storePodData(yamlFile []byte) error {
	pod := k8sv1.Pod{}

	err := yaml.Unmarshal(yamlFile, &pod)
	if err != nil {
		log.Log.Fatalln("failed to unmarshal yaml  - ", err)
		return err
	}

	l.objectStore.Add(&pod)
	return nil
}

// store single VM object from file
func (l *logsHandler) storeVMData(yamlFile []byte) error {
	vm := kubevirtv1.VirtualMachine{}

	err := yaml.Unmarshal(yamlFile, &vm)
	if err != nil {
		log.Log.Fatalln("failed to unmarshal vm yaml  - ", err)
		return err
	}

	l.objectStore.Add(&vm)
	return nil
}

func (l *logsHandler) storeVMIData(yamlFile []byte) error {
	vmi := kubevirtv1.VirtualMachineInstance{}

	err := yaml.Unmarshal(yamlFile, &vmi)
	if err != nil {
		log.Log.Fatalln("failed to unmarshal vmi yaml  - ", err)
		return err
	}

	l.objectStore.Add(&vmi)
	return nil
}

func (l *logsHandler) storeVMIMData(yamlFile []byte) error {
	vmim := kubevirtv1.VirtualMachineInstanceMigration{}

	err := yaml.Unmarshal(yamlFile, &vmim)
	if err != nil {
		log.Log.Fatalln("failed to unmarshal vmi migration yaml  - ", err)
		return err
	}

	l.objectStore.Add(vmim)
	return nil
}

func (l *logsHandler) storeVMIMListData(yamlFile []byte) error {
	vmimList := kubevirtv1.VirtualMachineInstanceMigrationList{}

	err := yaml.Unmarshal(yamlFile, &vmimList)
	if err != nil {
		log.Log.Fatalln("failed to unmarshal vmi migration yaml  - ", err)
		return err
	}

	for _, vmim := range vmimList.Items {
		l.objectStore.Add(vmim)
	}
	return nil
}

func (l *logsHandler) storePVCData(yamlFile []byte) error {
	pvc := k8sv1.PersistentVolumeClaim{}

	err := yaml.Unmarshal(yamlFile, &pvc)
	if err != nil {
		log.Log.Fatalln("failed to unmarshal pvc yaml  - ", err)
		return err
	}

	l.objectStore.Add(&pvc)
	return nil
}

func (l *logsHandler) storePVCListData(yamlFile []byte) error {
	pvcList := k8sv1.PersistentVolumeClaimList{}

	err := yaml.Unmarshal(yamlFile, &pvcList)
	if err != nil {
		log.Log.Fatalln("failed to unmarshal pvc yaml  - ", err)
		return err
	}

	for _, pvc := range pvcList.Items {
		l.objectStore.Add(&pvc)
	}
	return nil
}

func (l *logsHandler) storeNodeData(yamlFile []byte) error {
	node := k8sv1.Node{}

	err := yaml.Unmarshal(yamlFile, &node)
	if err != nil {
		log.Log.Fatalln("failed to unmarshal node yaml  - ", err)
		return err
	}

	l.objectStore.Add(&node)
	return nil
}

func (l *logsHandler) processImportedMustGather(filename string, insightsData string) error {
	l.handlerLock.Lock()
	defer l.handlerLock.Unlock()

	gatherTime, err := getMustGatherTimestamp()
	if err != nil {
		return err
	}

	l.objectStore.Add(&db.ImportedMustGather{
		Name:         filename,
		ImportTime:   time.Now(),
		GatherTime:   gatherTime,
		InsightsData: insightsData,
	})
	return nil
}

func getMustGatherTimestamp() (time.Time, error) {
	timestampFile, err := os.Open(fmt.Sprintf("/space/timestamp"))
	if err != nil {
		return time.Time{}, err
	}

	scanner := bufio.NewScanner(timestampFile)
	scanner.Scan()
	timestamp := scanner.Text()
	timestampFile.Close()

	return timestampStringToTime(timestamp)
}

func timestampStringToTime(timestamp string) (time.Time, error) {
	const layout = "2006-01-02 15:04:05.999999999 -0700 MST m=+0.000000000"
	const layoutNoMono = "2006-01-02 15:04:05.999999999 -0700 MST"

	t, err := time.Parse(layout, timestamp)
	if err != nil {
        // remove the monolithic offset if we can't parse it
        parts := strings.Split(timestamp, " m=")
        
        t, err = time.Parse(layoutNoMono, parts[0])
        if err != nil {
            return time.Time{}, err
        }
	}

	return t, nil
}

func (l *logsHandler) processPodYAMLs() error {
	var pod Pods
	l.handlerLock.Lock()
	defer l.handlerLock.Unlock()

	l.loadExistingEnrichmentData()

	//TODO: make path configurable
	layouts, err := filepath.Glob("/space/namespaces/*/pods/*/*.yaml")
	if err != nil {
		return (err)
	}

	for _, filename := range layouts {

		// read pod yaml
		yamlFile, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(yamlFile, &pod)
		if err != nil {
			return err
		}
		l.processEnrichmentData(&pod)
		l.storePodData(yamlFile)
	}

	js1, _ := json.Marshal(l.lookupData)
	_ = ioutil.WriteFile(ENRICHMENT_DATA_FILE, js1, 0644)

	log.Log.Println("finished writting lookupData")
	return nil
}

// store multiple VM objects from a single file.
func (l *logsHandler) storeVMListData(yamlFile []byte) error {
	vmList := kubevirtv1.VirtualMachineList{}

	err := yaml.Unmarshal(yamlFile, &vmList)
	if err != nil {
		log.Log.Fatalln("failed to unmarshal vm yaml list - ", err)
		return err
	}

	for _, vm := range vmList.Items {
		l.objectStore.Add(vm)
	}
	return nil
}

func (l *logsHandler) processCombinedVirtualMachineInstanceYAMLs(yamlFile []byte) error {

	dec := yamlv3.NewDecoder(bytes.NewReader(yamlFile))

	for {
		var vmi kubevirtv1.VirtualMachineInstance
		if dec.Decode(&vmi) != nil {
			break
		}
		l.objectStore.Add(&vmi)
	}
	return nil
}

func (l *logsHandler) processVirtualMachineYAMLs() error {
	// different versions of the must-gather collect the VM yamls differently

	l.handlerLock.Lock()
	defer l.handlerLock.Unlock()

	//TODO: make path configurable
	layouts, err := filepath.Glob("/space/namespaces/*/kubevirt.io/virtualmachines/*.yaml")
	if err != nil {
		return (err)
	}

	for _, filename := range layouts {
		yamlFile, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		l.storeVMData(yamlFile)
	}

	if len(layouts) == 0 {

		combinedYamls, err := filepath.Glob("/space/namespaces/*/kubevirt.io/virtualmachines.yaml")
		if err != nil {
			return (err)
		}
		for _, filename := range combinedYamls {
			yamlFile, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}
			l.storeVMListData(yamlFile)
		}
	}

	log.Log.Println("finished processing VMI YAMLs")
	return nil
}

func (l *logsHandler) processVirtualMachineInstanceYAMLs() error {
	// different versions of the must-gather collect the VMI yamls differently

	l.handlerLock.Lock()
	defer l.handlerLock.Unlock()

	//TODO: make path configurable
	layouts, err := filepath.Glob("/space/namespaces/*/kubevirt.io/virtualmachineinstances/*.yaml")
	if err != nil {
		return (err)
	}

	for _, filename := range layouts {
		yamlFile, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		l.storeVMIData(yamlFile)
	}

	if len(layouts) == 0 {

		combinedYamls, err := filepath.Glob("/space/namespaces/*/kubevirt.io/virtualmachineinstances.yaml")
		if err != nil {
			return (err)
		}
		for _, filename := range combinedYamls {
			yamlFile, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}
			l.processCombinedVirtualMachineInstanceYAMLs(yamlFile)
		}
	}

	log.Log.Println("finished processing VMI YAMLs")
	return nil
}

func (l *logsHandler) processVirtualMachineInstanceMigrationsYAMLs() error {
	// different versions of the must-gather collect the VMI yamls differently

	l.handlerLock.Lock()
	defer l.handlerLock.Unlock()

	//TODO: make path configurable
	layouts, err := filepath.Glob("/space/namespaces/*/kubevirt.io/virtualmachineinstancemigrations/*.yaml")
	if err != nil {
		return (err)
	}

	for _, filename := range layouts {
		yamlFile, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		l.storeVMIMData(yamlFile)
	}

	if len(layouts) == 0 {

		combinedYamls, err := filepath.Glob("/space/namespaces/*/kubevirt.io/virtualmachineinstancemigrations.yaml")
		if err != nil {
			return (err)
		}
		for _, filename := range combinedYamls {
			yamlFile, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}
			l.storeVMIMListData(yamlFile)
		}
	}

	log.Log.Println("finished processing VMIM YAMLs")
	return nil
}

func (l *logsHandler) processNodeYAMLs() error {
	l.handlerLock.Lock()
	defer l.handlerLock.Unlock()

	//TODO: make path configurable
	layouts, err := filepath.Glob("/space/cluster-scoped-resources/core/nodes/*.yaml")
	if err != nil {
		return (err)
	}

	for _, filename := range layouts {
		log.Log.Println("processing node YAML: ", filename)
		yamlFile, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		l.storeNodeData(yamlFile)
	}

	log.Log.Println("finished processing node YAMLs")
	return nil
}

func (l *logsHandler) processPersistentVolumeClaimYAMLs() error {
	// different versions of the must-gather collect the PVC yamls differently

	l.handlerLock.Lock()
	defer l.handlerLock.Unlock()

	//TODO: make path configurable
	layouts, err := filepath.Glob("/space/namespaces/*/core/persistentvolumeclaims/*.yaml")
	if err != nil {
		return (err)
	}

	for _, filename := range layouts {
		yamlFile, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		l.storePVCData(yamlFile)
	}

	if len(layouts) == 0 {

		combinedYamls, err := filepath.Glob("/space/namespaces/*/core/persistentvolumeclaims.yaml")
		if err != nil {
			return (err)
		}
		for _, filename := range combinedYamls {
			yamlFile, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}
			l.storePVCListData(yamlFile)
		}
	}

	log.Log.Println("finished processing PVC YAMLs")
	return nil
}

func (l *logsHandler) processSubscriptionsYAMLs() error {
	l.handlerLock.Lock()
	defer l.handlerLock.Unlock()

	layouts, err := filepath.Glob("/space/namespaces/*/subscriptions")
	if err != nil {
		return err
	}

	if len(layouts) == 0 {
		log.Log.Println("no subscriptions found")
	}

	for _, filename := range layouts {
		log.Log.Println("processing subscriptions YAML: ", filename)
		yamlFile, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		yamlFile = bytes.TrimRight(yamlFile, " \n-")

		err = l.storeSubscriptionData(yamlFile)
		if err != nil {
			return err
		}
	}

	log.Log.Println("finished processing node YAMLs")
	return nil
}

func (l *logsHandler) storeSubscriptionData(yamlFile []byte) error {
	subscription := v1alpha1.Subscription{}

	err := yaml.Unmarshal(yamlFile, &subscription)
	if err != nil {
		log.Log.Fatalln("failed to unmarshal subscription yaml  - ", err)
		return err
	}

	l.objectStore.Add(&subscription)
	return nil
}
