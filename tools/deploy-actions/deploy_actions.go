package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"cloud.google.com/go/storage"
	"github.com/namsral/flag"
	"google.golang.org/api/iterator"
	"k8s.io/client-go/kubernetes"

	"github.com/twineai/actions/tools/deploy-actions/action"
	"github.com/twineai/actions/tools/deploy-actions/deploymentmgr"
	"github.com/twineai/actions/tools/deploy-actions/servicemgr"
)

func isAction(objAttrs *storage.ObjectAttrs) bool {
	return filepath.Ext(objAttrs.Name) == ".tgz"
}

func actionName(objAttrs *storage.ObjectAttrs) string {
	path := objAttrs.Name
	ext := filepath.Ext(path)
	name := path[0 : len(path)-len(ext)]
	return name
}

func work(
	ctx context.Context,
	kubeClient kubernetes.Interface,
	ns string,
	bucketName string,
	action action.Action,
) error {
	workerCtx, _ := context.WithCancel(ctx)

	select {
	case <-workerCtx.Done():
		return workerCtx.Err()
	default:
		deploymentMgr := deploymentmgr.NewDeploymentManager(kubeClient, ns, bucketName, action)
		err := deploymentMgr.Run(workerCtx)
		if err != nil {
			return err
		}

		serviceMgr := servicemgr.NewServiceManager(kubeClient, ns, bucketName, action)
		err = serviceMgr.Run(workerCtx)
		if err != nil {
			return err
		}

		return nil
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Don't print any sort of timestamp information.
	log.SetFlags(0)

	flag.Parse()

	ns := strings.TrimSpace(FlagNamespace)
	if len(ns) == 0 {
		fmt.Fprintln(os.Stderr, "Argument 'namespace' is required")
		flag.Usage()
		os.Exit(1)
	}

	bucketName := strings.TrimSpace(FlagBucket)
	if len(bucketName) == 0 {
		fmt.Fprintln(os.Stderr, "Argument 'namespace' is required")
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("Bucket: %s", bucketName)
	log.Printf("Namespace: %s", ns)

	storageClient := OpenStorage(ctx)
	kubeClient := OpenKubeClient(ctx, FlagKubeconfig)
	bucket := storageClient.Bucket(bucketName)
	wg := sync.WaitGroup{}

	actions := make(chan action.Action, 100)
	errors := make(chan error, 100)

	// Listen for cancellation signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, ok := <-signalChan

		if ok {
			log.Printf("Signal received, cancelling…")
			cancel()
		}
	}()

	// Start our worker pool
	for i := 0; i < FlagWorkerCount; i++ {
		go func() {
			for action := range actions {
				wg.Add(1)
				err := work(ctx, kubeClient, ns, bucketName, action)
				errors <- err
				wg.Done()
			}
		}()
	}

	jobCount := 0
	iter := bucket.Objects(ctx, nil)
scheduleLoop:
	for {
		// See if we've been cancelled before scheduling any more work.
		select {
		case <-ctx.Done():
			close(actions)
			break scheduleLoop
		default:
			break
			// Do nothing
		}

		objAttrs, iterErr := iter.Next()

		if iterErr == iterator.Done {
			close(actions)
			break
		}

		if iterErr != nil {
			log.Fatalf("Error listing objects in bucket '%s': %v", bucketName, iterErr)
		}

		if !isAction(objAttrs) {
			continue
		}

		actionName := actionName(objAttrs)
		actionId := fmt.Sprintf("%s_%s_%d", bucketName, objAttrs.Name, objAttrs.Generation)
		actions <- action.Action{
			Name: actionName,
			Id:   actionId,
		}

		jobCount++
	}

	if jobCount > 0 {
		log.Printf("Waiting on %d operations to finish…", jobCount)
	waitLoop:
		for {
			select {
			case err := <-errors:
				if err != nil && err != context.Canceled {
					log.Printf("Error: %v", err)
				}

				// If that was the last response from the work queue, then we can bail.
				if jobCount--; jobCount == 0 {
					log.Printf("All operations finished")
					break waitLoop
				}
			}
		}

		wg.Wait()
	}

	return
}
