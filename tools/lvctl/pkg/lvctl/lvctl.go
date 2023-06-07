package lvctl

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	routeClient "github.com/openshift/client-go/route/clientset/versioned"
	templateClient "github.com/openshift/client-go/template/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

const (
	defaultLogsviewerImage = "quay.io/vladikr/logsviewer:devel"
	defaultStorageClass    = "ocs-storagecluster-ceph-rbd"
)

type LogsViewer struct {
	instanceID         string
	namespace          string
	storageClass       string
	image              string
	mustGatherFileName string

	kubeconfig     string
	k8sClient      *kubernetes.Clientset
	templateClient *templateClient.Clientset
	routeClient    *routeClient.Clientset
}

func Run() {
	lg := &LogsViewer{}

	setupCommand, deleteCommand, importCommand, setupImportCommand := lg.setupFlags()

	if len(os.Args) < 2 {
		klog.Exit("expected 'setup', 'delete', 'import' or 'setupImportCommand' subcommands")
	}

	lg.createClientset()

	switch os.Args[1] {
	case "help":
		fmt.Println("Control an instance of the logsViewer")
		fmt.Println("Syntax: lvctl [command] [options]")
		fmt.Println()
		setupCommand.Usage()
		fmt.Println()
		deleteCommand.Usage()
		fmt.Println()
		importCommand.Usage()
		fmt.Println()
		setupImportCommand.Usage()
	case "setup":
		mustSucceed(setupCommand.Parse(os.Args[2:]))
		lg.setup()
	case "delete":
		mustSucceed(deleteCommand.Parse(os.Args[2:]))
		lg.delete()
	case "import":
		mustSucceed(importCommand.Parse(os.Args[2:]))
		lg.importMustGather()
	case "setup-import":
		mustSucceed(setupImportCommand.Parse(os.Args[2:]))
		lg.setup()
		lg.importMustGather()
	default:
		klog.Exit("unknown command: ", os.Args[1])
	}
}

func (lg *LogsViewer) setupFlags() (setupCommand, deleteCommand, importCommand, setupImportCommand *flag.FlagSet) {
	setupCommand = flag.NewFlagSet("setup", flag.ExitOnError)
	lg.commonFlags(setupCommand)
	setupCommand.StringVar(&lg.storageClass, "storage-class", defaultStorageClass, "The storage class to use")
	setupCommand.StringVar(&lg.image, "image", defaultLogsviewerImage, "The LogsViewer image to use")

	deleteCommand = flag.NewFlagSet("delete", flag.ExitOnError)
	lg.commonFlags(deleteCommand)

	importCommand = flag.NewFlagSet("import", flag.ExitOnError)
	lg.commonFlags(importCommand)
	importCommand.StringVar(&lg.mustGatherFileName, "file", "", "The must-gather file to import")

	setupImportCommand = flag.NewFlagSet("setup-import", flag.ExitOnError)
	lg.commonFlags(setupImportCommand)
	setupImportCommand.StringVar(&lg.storageClass, "storage-class", defaultStorageClass, "The storage class to use")
	setupImportCommand.StringVar(&lg.image, "image", defaultLogsviewerImage, "The LogsViewer image to use")
	setupImportCommand.StringVar(&lg.mustGatherFileName, "file", "", "The must-gather file to import")

	return
}

func (lg *LogsViewer) commonFlags(flagSet *flag.FlagSet) {
	flagSet.StringVar(&lg.instanceID, "id", "", "The instance id")
	flagSet.StringVar(&lg.namespace, "namespace", "", "The namespace (defaults to the current namespace)")

	if home := homedir.HomeDir(); home != "" {
		flagSet.StringVar(&lg.kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	} else {
		flagSet.StringVar(&lg.kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
}

func (lg *LogsViewer) createClientset() {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: lg.kubeconfig},
		&clientcmd.ConfigOverrides{},
	)
	rawConfig, err := config.RawConfig()
	if err != nil {
		klog.Exit(err.Error())
	}
	if lg.namespace == "" {
		lg.namespace = rawConfig.Contexts[rawConfig.CurrentContext].Namespace
	}

	clientConfig, err := config.ClientConfig()
	if err != nil {
		klog.Exit(err.Error())
	}

	lg.k8sClient, err = kubernetes.NewForConfig(clientConfig)
	if err != nil {
		klog.Exit(err.Error())
	}

	lg.templateClient, err = templateClient.NewForConfig(clientConfig)
	if err != nil {
		klog.Exit(err.Error())
	}

	lg.routeClient, err = routeClient.NewForConfig(clientConfig)
	if err != nil {
		klog.Exit(err.Error())
	}
}

func mustSucceed(err error) {
	if err != nil {
		klog.Exit(err.Error())
	}
}
