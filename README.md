# Troubleshooting as a serivce

LogsViewer aims to facilitate the troubleshooting of OpenShift Virtualization.
In the future the service can be extended and generalized.

## Main flow
The service operates on collected system logs file.
All imported log files will be streamed into an integrated instance of elasticsearch.

The service provides a UI to reflect the relevant stored objects, such as Virtual Machines (Instances), Pods, Migration Objects, Nodes, etc. 
Selecting an object will generate a focused query for kibana/ elasticsearch to presents the logs of all relevant components which are related to this object.
For example, selecting a Virtual Machine Instance will generate a query that will include the associated virt-launcher, relevant virt-handler and virt-controller and virt-api.
The query will be bound by the Virtual Machine lifecycle timeline.
 
## Running the service

The service should be deployed on a running OpenShift cluster 

- creating an instance
```bash

$ ./deployment/lvctl.sh --create

service/logsviewer-n482tc created     
route.route.openshift.io/logviewer-n482tc created
route.route.openshift.io/kibana-n482tc created
configmap/es-configmap-n482tc created      
configmap/kibana-configmap-n482tc created  
configmap/logstash-configmap-n482tc created
pod/logsviewer-n482tc created
persistentvolumeclaim/elasticsearch-n482tc created
Waiting for logsviewer-n482tc pod .............................................................................................................................................................DONE

NAME               HOST/PORT                                            PATH   SERVICES            PORT   TERMINATION   WILDCARD
kibana-n482tc      kibana-n482tc.apps.cnv2.engineering.redhat.com              logsviewer-n482tc   5601                 None
logviewer-n482tc   logsviewer-n482tc.apps.cnv2.engineering.redhat.com          logsviewer-n482tc   8080                 None
```

- delete the instance

```bash
$ ./deployment/lvctl.sh --delete --suffix=n482tc                                             

service "logsviewer-n482tc" deleted                        
route.route.openshift.io "logviewer-n482tc" deleted   
route.route.openshift.io "kibana-n482tc" deleted                                                                       
configmap "es-configmap-n482tc" deleted
configmap "kibana-configmap-n482tc" deleted
configmap "logstash-configmap-n482tc" deleted
pod "logsviewer-n482tc" deleted
persistentvolumeclaim "elasticsearch-n482tc" deleted
```

## Routes

## Collecting system logs

- Control plane logs

The relevant Openshift Virtualization logs can be collected using the following command.
However, this will not collect `any` Virtula Machines logs

```bash
  $ oc adm must-gather --image=registry.redhat.io/container-native-virtualization/cnv-must-gather-rhel8:v4.11.0
```

- Virtual Machine logs

As it is today, relevant Virtual Machine logs can be only collected for the whole namespace where virtual machines are running.

```bash
 oc adm must-gather -n [Namespace where Virtual Machines run]
```

## Import logs

The service consumes compressed must-gathers. 
This is the entry point for any operaion.
Head to the `Import` tab in the logsviewer UI to upload the logs.


