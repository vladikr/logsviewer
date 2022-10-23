package db

import (
	"fmt"
	"sync"
	"time"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"

//	v1 "kubevirt.io/api/core/v1"
    
    "logsviewer/pkg/backend/log"
)

func NewObjectStore() *ObjectStore {

	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "objectStore")
	//db := NewDatabaseInstance()
	c := &ObjectStore{
		Queue:             queue,
		lockDBConn:        &sync.Mutex{},
	}

	return c
}

type ObjectStore struct {
	Queue             workqueue.RateLimitingInterface
	storeDB           *databaseInstance
	lockDBConn   	  *sync.Mutex
    wg                sync.WaitGroup

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
		if err := c.storeDB.InitTables(); err != nil {
            log.Log.Println("failed to connect to database", err)
			if err := c.storeDB.DropTables(); err != nil {
                log.Log.Println("failed to drop tables", err)
			}
			c.storeDB.Shutdown()
			c.storeDB = nil
			return false
		}
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
		//c.Queue.AddRateLimited(obj)
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

func (d *ObjectStore) storePod(pod *k8sv1.Pod) error {
    jsonBytes, err := json.Marshal(pod)
    if err != nil {
        log.Log.Println("failed to marshal pod object ", pod, " err: ", err)
    }

    totalContainers, activeContainers := d.countPodContainers(pod)
    name := pod.GetObjectMeta().GetName()
    namespace := pod.GetObjectMeta().GetNamespace()
    uid := string(pod.GetObjectMeta().GetUID())
    kind := pod.GetObjectKind().GroupVersionKind().Kind
	storeObj := &Pod{
		Key:       fmt.Sprintf("%s/%s", name, namespace),
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
		UUID:      uid,
        Phase:     string(pod.Status.Phase),    
        ActiveContainers: activeContainers,
        TotalContainers: totalContainers,
        NodeName: string(pod.Spec.NodeName),
        CreationTime: pod.CreationTimestamp,
        Content: jsonBytes,
	}
	if err := d.storeDB.StorePod(storeObj); err != nil {
        log.Log.Println("failed to store obj  ", storeObj, " err: ", err)
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
	default:
		jsonBytes, err := json.Marshal(obj)
		if err != nil {
            log.Log.Println("failed to marshal obj", obj, " err: ", err)
		}
		log.Log.Println(jsonBytes)
	}
}

func (d *ObjectStore) Add(obj interface{}) {

	switch v := obj.(type) {
	case *k8sv1.Pod:
		podObj := obj.(*k8sv1.Pod)
        
        d.wg.Add(1)    
	    d.Queue.Add(podObj)
	default:
		log.Log.Println("Cannot store unsupported obj ", v)
    }
}
