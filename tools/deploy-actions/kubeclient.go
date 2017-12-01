package main

import (
	"context"
	"log"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func OpenKubeClient(ctx context.Context, configName string) kubernetes.Interface {
	var config *rest.Config
	var err error

	if configName != "" {
		config, err = clientcmd.BuildConfigFromFlags("", configName)
		if err != nil {
			log.Fatalf("unable to create client config named '%s': %v", configName, err)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("unable to create in-cluster config: %v", err)
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("unable to create kubernetes client: %v", err)
	}

	return client
}
