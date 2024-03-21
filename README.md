# Troubleshooting as a service

LogsViewer aims to facilitate the troubleshooting of Kubernetes, OpenShift and related operators.

## Goal
The service provides a UI to reflect the relevant stored objects, such as Pods, Migration Objects, Nodes, Virtual Machines (Instances), etc.
Selecting an object will generate a focused query for kibana/ elasticsearch to presents the logs of all relevant components which are related to this object.

For example, selecting a Virtual Machine Instance will generate a query that will include the associated virt-launcher, relevant virt-handler and virt-controller and virt-api.
The query will be bound by the Virtual Machine lifecycle timeline.

All imported log files will be streamed into an integrated instance of elasticsearch.

The service operates on compressed collected must-gather files, but can be extended to work also with cluster dump or connect to a live cluster.

Logsviewer also run insights rules to detect known issues and reports them if identified.
 
## Running the service

Here are a few steps to quickly get the service up and running. For more
information and custom configurations about the lvctl command, head to the
[lvctl documentation](./tools/lvctl/README.md).

### Get a lvctl binary

```bash
	$ git clone https://github.com/vladikr/logsviewer
	$ cd logsviewer/
	$ make build -C tools/lvctl/
	$ mv tools/lvctl/bin/lvctl /usr/local/bin/
```

### Create an instance of the LogsViewer
```bash
  # To deploy LogsViewer in the current namespace with default configuration
  ./lvctl setup

  # Or if you want to import a must-gather file into the LogsViewer during the setup
  ./lvctl setup-import -namespace <namespace> -id <instance-id> -file <path-to/must-gather-file.tar.gz>
```

This commands will log the the route created to access the LogsViewer UI.

*Note:* It is possible to import the files manually through the UI after the instance was created.

### Import Additional Must Gather files to the same Logsviewer instance
```bash
  ./lvctl import -namespace <namespace> -id <instance-id> -file <path-to/must-gather-file.tar.gz>
```

*Note:*
  - You can upload more than one Must Gather to the same Logsviewer instance for the same cluster.
  - You can upload both OpenShift and other Must gathers for the same cluster.
  - By default, the instance will be deleted 48 hours after the last must-gather file was imported. Or after creation time if no must-gather file was imported. Check the '-deletion-condition' and '-deletion-delay' flags for more details.

### Manually delete an instance of the LogsViewer

```bash
  ./lvctl delete  -namespace <namespace> -id <instance-id>
```

### Log into the Logsviewer Pod UI
The UI link would be:
http://logsviewer-<namespace><instance-id>.example.com


## Demo

[logsviewerNew.webm](https://github.com/vladikr/logsviewer/assets/1035064/0a71f97e-b5c7-45b4-8a21-262c0a40806f)
