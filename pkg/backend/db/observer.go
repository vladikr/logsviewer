package db

import (
	"fmt"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"strings"
	"sync"
	"time"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"

	kubevirtv1 "kubevirt.io/api/core/v1"

	"logsviewer/pkg/backend/log"
)

func NewObjectStore(storeDB *DatabaseInstance) *ObjectStore {

	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "objectStore")
	c := &ObjectStore{
		Queue:      queue,
		lockDBConn: &sync.Mutex{},
		storeDB:    storeDB,
	}

	return c
}

type ObjectStore struct {
	Queue      workqueue.RateLimitingInterface
	storeDB    *DatabaseInstance
	lockDBConn *sync.Mutex
	wg         sync.WaitGroup
}

func (c *ObjectStore) Run(threadiness int, stopCh chan struct{}) {
	defer c.Queue.ShutDown()
	//workers := 1

	// Start the actual work
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	c.wg.Wait()

}

func (c *ObjectStore) runWorker() {
	log.Log.Println("runWorker")
	for c.Execute() {
	}
}

func (c *ObjectStore) connectDatabaseIfNeeded() bool {
	if c.storeDB == nil {
		c.lockDBConn.Lock()
		defer c.lockDBConn.Unlock()

		dbInst, err := NewDatabaseInstance()
		if err != nil {
			log.Log.Println("failed to connect to database", err)
			return false
		}
		c.storeDB = dbInst
	}
	return true

}

func (c *ObjectStore) Execute() bool {
	if succedded := c.connectDatabaseIfNeeded(); !succedded {
		log.Log.Println("failed to connect to DB")
		return false
	}

	obj, quit := c.Queue.Get()
	if quit {
		return false
	}
	defer c.Queue.Done(obj)
	if err := c.execute(obj); err != nil {
		log.Log.Println("re-enqueuing object ", obj, " err: ", err)
	} else {
		c.Queue.Forget(obj)
		c.wg.Done()
	}

	return true
}

func (d *ObjectStore) execute(obj interface{}) error {
	// Make sure we re-enqueue the key to ensure this new VMI is processed
	// after the stale domain is removed
	//	d.Queue.AddAfter(controller.VirtualMachineInstanceKey(vmi), time.Second*5)
	d.processObject(obj)

	return nil
}

func (d *ObjectStore) countPodContainers(pod *k8sv1.Pod) (int, int) {
	totalContainers := len(pod.Spec.Containers)
	activeContainers := 0
	for _, container := range pod.Status.ContainerStatuses {
		if container.State.Running != nil {
			activeContainers++
		}
	}
	return totalContainers, activeContainers
}

func (d *ObjectStore) formatPodPVCs(pod *k8sv1.Pod) string {
	pvcs := []string{}

	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			pvcs = append(pvcs, volume.PersistentVolumeClaim.ClaimName)
		}
	}
	pvcsStr := strings.Join(pvcs, ",")
	return pvcsStr
}

func (d *ObjectStore) storePod(pod *k8sv1.Pod) error {
	jsonBytes, err := json.Marshal(pod)
	if err != nil {
		log.Log.Println("failed to marshal pod object ", pod, " err: ", err)
	}

	createdByUID := pod.Labels[kubevirtv1.CreatedByLabel]
	totalContainers, activeContainers := d.countPodContainers(pod)
	name := pod.GetObjectMeta().GetName()
	namespace := pod.GetObjectMeta().GetNamespace()
	uid := string(pod.GetObjectMeta().GetUID())
	kind := pod.GetObjectKind().GroupVersionKind().Kind
	pvcsList := d.formatPodPVCs(pod)
	storeObj := &Pod{
		Key:              fmt.Sprintf("%s/%s", name, namespace),
		Kind:             kind,
		Name:             name,
		Namespace:        namespace,
		UUID:             uid,
		Phase:            string(pod.Status.Phase),
		ActiveContainers: activeContainers,
		TotalContainers:  totalContainers,
		NodeName:         string(pod.Spec.NodeName),
		CreationTime:     pod.CreationTimestamp,
		PVCs:             pvcsList,
		Content:          jsonBytes,
		CreatedBy:        createdByUID,
	}
	if err := d.storeDB.StorePod(storeObj); err != nil {
		log.Log.Println("failed to store obj  ", storeObj, " err: ", err)
		return err
	}
	return nil
}

func (d *ObjectStore) storeNode(node *k8sv1.Node) error {
	jsonBytes, err := json.Marshal(node)
	if err != nil {
		log.Log.Println("failed to marshal node object ", node, " err: ", err)
	}

	nodeStatus := "NotReady"
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" {
			nodeStatus = "Ready"
		}
	}

	nodeInternalIP := ""
	nodeHostName := ""
	for _, address := range node.Status.Addresses {
		switch address.Type {
		case "InternalIP":
			nodeInternalIP = address.Address
		case "Hostname":
			nodeHostName = address.Address
		default:
			continue
		}
	}

	name := node.GetObjectMeta().GetName()
	storeObj := &Node{
		Name:                    name,
		Status:                  nodeStatus,
		SystemUUID:              node.Status.NodeInfo.SystemUUID,
		InternalIP:              nodeInternalIP,
		HostName:                nodeHostName,
		OsImage:                 node.Status.NodeInfo.OSImage,
		KernelVersion:           node.Status.NodeInfo.KernelVersion,
		KubletVersion:           node.Status.NodeInfo.KubeletVersion,
		ContainerRuntimeVersion: node.Status.NodeInfo.ContainerRuntimeVersion,
		Content:                 jsonBytes,
	}
	if err := d.storeDB.StoreNode(storeObj); err != nil {
		log.Log.Println("failed to store obj  ", storeObj, " err: ", err)
		return err
	}
	return nil
}

func (d *ObjectStore) storeVmi(vmi *kubevirtv1.VirtualMachineInstance) error {
	jsonBytes, err := json.Marshal(vmi)
	if err != nil {
		log.Log.Println("failed to marshal vmi object ", vmi, " err: ", err)
	}

	name := vmi.GetObjectMeta().GetName()
	namespace := vmi.GetObjectMeta().GetNamespace()
	uid := string(vmi.GetObjectMeta().GetUID())
	storeObj := &VirtualMachineInstance{
		Name:         name,
		Namespace:    namespace,
		UUID:         uid,
		Phase:        string(vmi.Status.Phase),
		Reason:       string(vmi.Status.Reason),
		NodeName:     string(vmi.Status.NodeName),
		CreationTime: vmi.CreationTimestamp,
		Status:       vmi.Status,
		Content:      jsonBytes,
	}
	if err := d.storeDB.StoreVmi(storeObj); err != nil {
		log.Log.Println("failed to store vmi obj  ", storeObj, " err: ", err)
		return err
	}
	return nil
}

func (d *ObjectStore) storeVmiMigration(vmim kubevirtv1.VirtualMachineInstanceMigration) error {
	jsonBytes, err := json.Marshal(vmim)
	if err != nil {
		log.Log.Println("failed to marshal vmi migration object ", vmim, " err: ", err)
		return err
	}

	name := vmim.GetObjectMeta().GetName()
	namespace := vmim.GetObjectMeta().GetNamespace()
	uid := string(vmim.GetObjectMeta().GetUID())
	storeObj := &VirtualMachineInstanceMigration{
		Name:         name,
		Namespace:    namespace,
		UUID:         uid,
		Phase:        string(vmim.Status.Phase),
		VMIName:      string(vmim.Spec.VMIName),
		CreationTime: vmim.CreationTimestamp,
		Content:      jsonBytes,
	}

	if migrationState := vmim.Status.MigrationState; migrationState != nil {
		storeObj.TargetPod = migrationState.TargetPod
		storeObj.CreationTime = *migrationState.StartTimestamp
		storeObj.EndTimestamp = *migrationState.EndTimestamp
		storeObj.SourceNode = migrationState.SourceNode
		storeObj.TargetNode = migrationState.TargetNode
		storeObj.Completed = migrationState.Completed
		storeObj.Failed = migrationState.Failed
	}
	if err := d.storeDB.StoreVmiMigration(storeObj); err != nil {
		log.Log.Println("failed to store vmi migration obj  ", storeObj, " err: ", err)
		return err
	}
	return nil
}

func (d *ObjectStore) storePVC(pvc *k8sv1.PersistentVolumeClaim) error {
	jsonBytes, err := json.Marshal(pvc)
	if err != nil {
		log.Log.Println("failed to marshal pvc object ", pvc, " err: ", err)
	}

	name := pvc.GetObjectMeta().GetName()
	namespace := pvc.GetObjectMeta().GetNamespace()
	uid := string(pvc.GetObjectMeta().GetUID())
	storage := pvc.Status.Capacity.Storage()
	capacity := storage.String()

	accessModes := ""
	if len(pvc.Status.AccessModes) > 0 {
		accessModes = string(pvc.Status.AccessModes[0])
		for i := 1; i < len(pvc.Status.AccessModes); i++ {
			accessModes += fmt.Sprintf(",%s", pvc.Status.AccessModes[i])
		}
	}

	storageClassName := ""
	if pvc.Spec.StorageClassName != nil {
		storageClassName = *pvc.Spec.StorageClassName
	}

	volumeMode := ""
	if pvc.Spec.VolumeMode != nil {
		volumeMode = string(*pvc.Spec.VolumeMode)
	}

	storeObj := &PersistentVolumeClaim{
		Name:             name,
		Namespace:        namespace,
		UUID:             uid,
		Phase:            string(pvc.Status.Phase),
		AccessModes:      accessModes,
		StorageClassName: storageClassName,
		VolumeName:       pvc.Spec.VolumeName,
		VolumeMode:       volumeMode,
		Reason:           "",
		Capacity:         capacity,
		CreationTime:     pvc.CreationTimestamp,
		Content:          jsonBytes,
	}
	if err := d.storeDB.StorePVC(storeObj); err != nil {
		log.Log.Println("failed to store obj  ", storeObj, " err: ", err)
		return err
	}
	return nil
}

func (d *ObjectStore) storeSubscription(sub *v1alpha1.Subscription) error {
	jsonBytes, err := json.Marshal(sub)
	if err != nil {
		log.Log.Println("failed to marshal subscription object ", sub, " err: ", err)
	}

	storeObj := &Subscription{
		UUID: string(sub.UID),

		Name:            sub.Spec.Package,
		Namespace:       sub.Namespace,
		Source:          sub.Spec.CatalogSource,
		SourceNamespace: sub.Spec.CatalogSourceNamespace,
		StartingCSV:     sub.Spec.StartingCSV,
		CurrentCSV:      sub.Status.CurrentCSV,
		InstalledCSV:    sub.Status.InstalledCSV,
		State:           string(sub.Status.State),

		CreationTime: sub.CreationTimestamp,
		Content:      jsonBytes,
	}
	if err := d.storeDB.StoreSubscription(storeObj); err != nil {
		log.Log.Println("failed to store subscription obj  ", storeObj, " err: ", err)
		return err
	}
	return nil
}

func (d *ObjectStore) processObject(obj interface{}) {

	switch obj.(type) {
	case *k8sv1.Pod:
		podObj := obj.(*k8sv1.Pod)
		if err := d.storePod(podObj); err == nil {
			log.Log.Println("stored obj  ", podObj)
		}
	case *k8sv1.Node:
		nodeObj := obj.(*k8sv1.Node)
		if err := d.storeNode(nodeObj); err == nil {
			log.Log.Println("stored obj  ", nodeObj)
		}
	case *kubevirtv1.VirtualMachineInstance:
		vmi := obj.(*kubevirtv1.VirtualMachineInstance)
		if err := d.storeVmi(vmi); err == nil {
			log.Log.Println("stored vmi obj  ", vmi)
		}
	case *kubevirtv1.VirtualMachineInstanceMigration:
		vmim := obj.(*kubevirtv1.VirtualMachineInstanceMigration)
		vmimCopy := vmim.DeepCopy()

		err := d.storeVmiMigration(*vmimCopy)
		if err == nil {
			log.Log.Println("stored vmi migration obj  ", vmimCopy)
		} else {
			log.Log.Println("failed to store vmi migration obj  ", vmimCopy)

		}
	case *k8sv1.PersistentVolumeClaim:

		pvc := obj.(*k8sv1.PersistentVolumeClaim)
		if err := d.storePVC(pvc); err == nil {
			log.Log.Println("stored pvc obj  ", pvc)
		}
	case *v1alpha1.Subscription:
		sub := obj.(*v1alpha1.Subscription)
		if err := d.storeSubscription(sub); err == nil {
			log.Log.Println("stored subscription obj  ", sub)
		}
	default:
		jsonBytes, err := json.Marshal(obj)
		if err != nil {
			log.Log.Println("failed to marshal obj", obj, " err: ", err)
		}
		log.Log.Println("failed to process unknown object ", jsonBytes)
	}
}

func (d *ObjectStore) Add(obj interface{}) {

	switch v := obj.(type) {
	case *k8sv1.Pod:
		podObj := obj.(*k8sv1.Pod)
		d.wg.Add(1)
		d.Queue.Add(podObj)
	case *k8sv1.Node:
		nodeObj := obj.(*k8sv1.Node)
		d.wg.Add(1)
		d.Queue.Add(nodeObj)
	case *kubevirtv1.VirtualMachineInstance:
		vmi := obj.(*kubevirtv1.VirtualMachineInstance)
		d.wg.Add(1)
		d.Queue.Add(vmi)
	case kubevirtv1.VirtualMachineInstanceMigration:
		vmim := obj.(kubevirtv1.VirtualMachineInstanceMigration)
		d.wg.Add(1)
		d.Queue.Add(&vmim)
	case *k8sv1.PersistentVolumeClaim:
		pvc := obj.(*k8sv1.PersistentVolumeClaim)
		d.wg.Add(1)
		d.Queue.Add(pvc)
	case *v1alpha1.Subscription:
		sub := obj.(*v1alpha1.Subscription)
		d.wg.Add(1)
		d.Queue.Add(sub)
	default:
		log.Log.Println("Cannot store unsupported obj ", v)
	}
}
