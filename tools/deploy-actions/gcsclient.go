package main

import (
	"context"
	"log"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// OpenStorage creates a client for the Google Cloud Storage API.
func OpenStorage(ctx context.Context) *storage.Client {
	var storageClient *storage.Client
	var err error

	if FlagGCPCredentialsFile != "" {
		options := option.WithCredentialsFile(FlagGCPCredentialsFile)
		storageClient, err = storage.NewClient(ctx, options)
	} else {
		storageClient, err = storage.NewClient(ctx)
	}

	if err != nil {
		log.Fatalf("Unable to create Google Cloud Storage client: %v", err)
	}

	return storageClient
}
