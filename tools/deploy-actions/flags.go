package main

import (
	"github.com/namsral/flag"
)

var (
	FlagKubeconfig         string
	FlagNamespace          string
	FlagGCPCredentialsFile string
	FlagBucket             string
	FlagWorkerCount        int
)

func init() {
	flag.StringVar(
		&FlagKubeconfig, "kubeconfig", "",
		"The kubeconfig to use. If empty the in-cluster config will be used")

	flag.StringVar(
		&FlagNamespace, "namespace", "",
		"The namespace to use")

	flag.StringVar(
		&FlagBucket, "gcs-bucket-name", "twineai-actions-prod-d37dc540",
		"The GCS bucket containing actions")

	flag.StringVar(
		&FlagGCPCredentialsFile, "gcp-credentials-file", "",
		"Path to the GCP service account key file")

	flag.IntVar(
		&FlagWorkerCount, "worker-count", 3,
		"Number of workers to run")
}
