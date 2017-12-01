package servicemgr

import (
	"context"
	"log"

	"github.com/pkg/errors"
	"github.com/twineai/actions/tools/deploy-actions/action"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type serviceManager struct {
	kubeClient kubernetes.Interface

	action     action.Action
	namespace  string
	bucketName string
}

func NewServiceManager(
	kubeClient kubernetes.Interface,
	namespace string,
	bucketName string,
	action action.Action,
) *serviceManager {
	return &serviceManager{
		kubeClient: kubeClient,
		namespace:  namespace,
		bucketName: bucketName,
		action:     action,
	}
}

func (mgr *serviceManager) Run(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "error configuring service for action '%s'", mgr.action.Name)
		}
	}()

	old, err := mgr.getCurrentService(ctx)
	if err != nil && !kubeerrors.IsNotFound(err) {
		return err
	}

	if old == nil {
		return mgr.createService(ctx)
	} else {
		log.Printf("Service already exists for action '%s'", mgr.action.Name)
		return nil
	}
}

//
// Kubernetes Helpers
//

func (mgr *serviceManager) getCurrentService(
	ctx context.Context,
) (*corev1.Service, error) {
	resultChan := make(chan *corev1.Service)
	errChan := make(chan error)

	go func() {
		serviceName := mgr.serviceName()
		result, err := mgr.kubeClient.CoreV1().
			Services(mgr.namespace).
			Get(serviceName, metav1.GetOptions{})

		if err != nil {
			errChan <- err
		} else {
			resultChan <- result
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errChan:
			return nil, err
		case result := <-resultChan:
			return result, nil
		}
	}
}

func (mgr *serviceManager) createService(ctx context.Context) error {
	service := mgr.buildBaseService()
	result, err := mgr.kubeClient.
		CoreV1().
		Services(mgr.namespace).
		Create(service)
	if err != nil {
		return err
	}

	log.Printf("Created service: %v", result)
	return nil
}

func (mgr *serviceManager) buildBaseService() *corev1.Service {
	labels := mgr.serviceLabels()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mgr.serviceName(),
			Namespace: mgr.namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"twine-action": labels["twine-action"],
			},
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:     "grpc",
					Protocol: corev1.ProtocolTCP,
					Port:     8080,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8080,
					},
				},
			},
		},
	}
}
