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

Here are a few steps to quickly get the service up and running. For more
information and custom configurations about the lvctl command, head to the
[lvctl documentation](./tools/lvctl/README.md).

### Get a lvctl binary

```bash
  # To get the latest released lvctl binary
  LATEST_TAG=$(curl https://api.github.com/repos/kubevirt/kubevirt/releases/latest | grep -i "tag_name" | awk -F '"' '{print $4}')
  wget https://github.com/vladikr/logsviewer/releases/download/${LATEST_TAG}/lvctl
  chmod +x lvctl
  
  # Or to build the lvctl binary from source
  make build -C tools/lvctl/
  cp tools/lvctl/bin/lvctl . 
```

### Create an instance of the LogsViewer

```bash
  # To deploy LogsViewer in the current namespace with default configuration
  ./lvctl setup

  # Or if you want to import a must-gather file into the LogsViewer during the setup
  ./lvctl setup-import -file <path-to/must-gather-file.tar.gz>
```

This commands will log the instance id and the route created to access the
LogsViewer UI.

> **Warning**
> By default, the instance will be deleted 48 hours after the last must-gather
> file was imported. Or after creation time if no must-gather file was imported.
> Check the '-deletion-condition' and '-deletion-delay' flags for more details.

### Import a new must-gather file into an instance of the LogsViewer

```bash
  ./lvctl import -id <instance-id> -file <path-to/must-gather-file.tar.gz>
```

### Delete an instance of the LogsViewer

```bash
  ./lvctl delete -id <instance-id>
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
This is the entry point for any operation.
Head to the `Import` tab in the logsviewer UI to upload the logs.

## Demo

[logsviewerNew.webm](https://github.com/vladikr/logsviewer/assets/1035064/0a71f97e-b5c7-45b4-8a21-262c0a40806f)
