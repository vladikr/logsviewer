package db

import (
	"encoding/json"
	"time"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

type (
	Pod struct {
		Key       string `json:"keyid"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		UUID      string `json:"uuid"`
        Phase     string `json:"phase"`
        ActiveContainers int `json:"activeContainers"`
        TotalContainers  int `json:"totalContainers"`
        NodeName         string `json:"nodeName"`
        CreationTime     metav1.Time `json:"creationTime"`
		Content json.RawMessage `json:"content"`
        CreatedBy        string `json:"createdBy"`
	}

	VirtualMachineInstance struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		UUID      string `json:"uuid"`
        Reason     string `json:"reason"`
        Phase     string `json:"phase"`
        NodeName         string `json:"nodeName"`
        CreationTime     metav1.Time `json:"creationTime"`
        //PodName   string `json:"podName"`
        //HandlerPod  string `json:"handlerName"`
        Status kubevirtv1.VirtualMachineInstanceStatus `json:"status,omitempty"`
		Content json.RawMessage `json:"content"`
	}

	VirtualMachineInstanceMigration struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		UUID      string `json:"uuid"`
        Phase     string `json:"phase"`
        VMIName         string `json:"vmiName"`
        // The target pod that the VMI is moving to
        TargetPod string `json:"targetPod,omitempty"`
        CreationTime metav1.Time `json:"creationTime"`
        EndTimestamp metav1.Time `json:"endTimestamp,omitempty"`
        SourceNode string `json:"sourceNode,omitempty"`
        // The target node that the VMI is moving to
        TargetNode string `json:"targetNode,omitempty"`
        // Indicates the migration completed
        Completed bool `json:"completed,omitempty"`
        // Indicates that the migration failed
        Failed bool `json:"failed,omitempty"`
		Content json.RawMessage `json:"content"`
	}

	Node struct {
		Name                        string `json:"name"`
		SystemUUID                  string `json:"systemUuid"`
        Status                      string `json:"status"`
        InternalIP                  string `json:"internalIP"`
        HostName                    string `json:"hostName"`
        OsImage                     string `json:"osImage"`
        KernelVersion               string `json:"kernelVersion"`
        KubletVersion               string `json:"kubletVersion"`
        ContainerRuntimeVersion     string `json:"containerRuntimeVersion"`
		Content                     json.RawMessage `json:"content"`
	}

	QueryResults struct {
        Namespace   string
        SourcePodUUID     string
        TargetPodUUID     string
        VMIUUID         string
        MigrationUUID         string
        SourcePod string
        TargetPod string
        StartTimestamp time.Time
        EndTimestamp time.Time
        SourceHandler string
        TargetHandler string
	}
)
