# LogsViewer CLI tool

## Build

```bash
make build
```

## Usage

```bash
bin/lvctl [command] [flags]
```

## Commands

- `bin/lvctl setup [flags]` - to create a new instance of the LogsViewer

- `bin/lvctl delete [flags]` - to delete an instance of the LogsViewer

- `bin/lvctl import [flags]` - to import a must-gather file into an instance of the LogsViewer

- `bin/lvctl setup-import [flags]` - to create a new instance of the LogsViewer and import a must-gather file

- `bin/lvctl help` - for more information about the CLI tool, its commands and flags

```bash
> bin/lvctl help

Control an instance of the logsViewer
Syntax: lvctl [command] [options]

Usage of setup:
  -id string
        The instance id
  -image string
        The LogsViewer image to use (default "quay.io/vladikr/logsviewer:devel")
  -kubeconfig string
        absolute path to the kubeconfig file (default "/home/machadovilaca/.kube/config")
  -namespace string
        The namespace (defaults to the current namespace)
  -storage-class string
        The storage class to use (default "ocs-storagecluster-ceph-rbd")

Usage of delete:
  -id string
        The instance id
  -kubeconfig string
        absolute path to the kubeconfig file (default "/home/machadovilaca/.kube/config")
  -namespace string
        The namespace (defaults to the current namespace)

Usage of import:
  -file string
        The must-gather file to import
  -id string
        The instance id
  -kubeconfig string
        absolute path to the kubeconfig file (default "/home/machadovilaca/.kube/config")
  -namespace string
        The namespace (defaults to the current namespace)

Usage of setup-import:
  -file string
        The must-gather file to import
  -id string
        The instance id
  -image string
        The LogsViewer image to use (default "quay.io/vladikr/logsviewer:devel")
  -kubeconfig string
        absolute path to the kubeconfig file (default "/home/machadovilaca/.kube/config")
  -namespace string
        The namespace (defaults to the current namespace)
  -storage-class string
        The storage class to use (default "ocs-storagecluster-ceph-rbd")
```
