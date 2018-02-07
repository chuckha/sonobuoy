package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	SonobuoyPod = "sonobuoy"
)

// OutOfClusterClient returns a kubernetes client that is accessing the
// cluster from outside the cluster.
func OutOfClusterClient() (kubernetes.Interface, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, fmt.Errorf("could not build config from kubeconfig: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not make a new clientset from config: %v", err)
	}
	return clientset, nil
}

// GetConfig returns a kubernetes client.
func GetConfig() (*rest.Config, error) {
	kubeconfig := locateKubeconfig()
	if len(kubeconfig) == 0 {
		return nil, errors.New("Could not locate kubeconfig")
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func locateKubeconfig() string {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		u, err := user.Current()
		if err != nil {
			return ""
		}
		kubeconfig = filepath.Join(u.HomeDir, ".kube", "config")
		// make sure this file exists
		_, err = os.Stat(kubeconfig)
		if err != nil {
			return ""
		}
	}
	return kubeconfig
}
