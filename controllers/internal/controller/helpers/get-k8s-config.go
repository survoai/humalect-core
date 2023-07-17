package services

import (
	"log"
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetK8sConfig() *rest.Config {
	var config *rest.Config
	var err error

	// Try to load the in-cluster configuration
	config, err = rest.InClusterConfig()

	if err != nil {
		// If in-cluster configuration is not available, use the kubeconfig file
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			// Fall back to the default kubeconfig file location if the KUBECONFIG environment variable is not set
			kubeconfig = os.ExpandEnv("$HOME/.kube/config")
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %v", err)
			panic(err)
		}
	}

	return config
}
