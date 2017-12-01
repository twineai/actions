package deploymentmgr

import (
	"context"
	"encoding/json"
	"log"

	"github.com/pkg/errors"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"

	"github.com/twineai/actionserver/tools/deploy-actions/action"
)

func NewDeploymentManager(
	kubeClient kubernetes.Interface,
	namespace string,
	bucketName string,
	action action.Action,
) *deploymentManager {
	return &deploymentManager{
		kubeClient:           kubeClient,
		action:               action,
		namespace:            namespace,
		bucketName:           bucketName,
		serverImageName:      FlagActionServerImageName,
		serverSetupImageName: FlagActionServerSetupImageName,
		serverVersion:        FlagActionServerVersion,
	}
}

type deploymentManager struct {
	kubeClient kubernetes.Interface

	action               action.Action
	namespace            string
	bucketName           string
	serverVersion        string
	serverImageName      string
	serverSetupImageName string
}

//
// Deployment Management
//

func (mgr *deploymentManager) Run(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "error configuring deployment for action '%s'", mgr.action.Name)
		}
	}()

	old, err := mgr.getCurrentDeployment(ctx)
	if err != nil && !kubeerrors.IsNotFound(err) {
		return err
	}

	if kubeerrors.IsNotFound(err) {
		log.Printf("Creating deployment for action: %s", mgr.action.Name)
		return mgr.createDeployment(ctx)
	} else {
		txt, err := json.Marshal(old)
		if err != nil {
			return err
		}

		log.Printf("updating existing: %s", string(txt))

		return mgr.updateDeployment(ctx, old)
	}

	return nil
}

//
// Kubernetes Helpers
//

func (mgr *deploymentManager) getCurrentDeployment(ctx context.Context) (*appsv1beta2.Deployment, error) {
	resultChan := make(chan *appsv1beta2.Deployment)
	errChan := make(chan error)

	go func() {
		deploymentName := mgr.deploymentName()
		result, err := mgr.kubeClient.AppsV1beta2().
			Deployments(mgr.namespace).
			Get(deploymentName, metav1.GetOptions{})

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

func (mgr *deploymentManager) createDeployment(ctx context.Context) error {
	// We take a context here for the future when we can pass it along to kubernetes, but for our purposes, once the
	// operation starts, it can't be cancelled.

	deployment := mgr.buildBaseDeployment()
	deployment = mgr.applyUpdates(deployment)
	result, err := mgr.kubeClient.AppsV1beta2().Deployments(mgr.namespace).Create(deployment)
	if err != nil {
		return errors.Wrap(err, "error creating deployment for action")
		return err
	}

	log.Printf("Created deployment: %v", result)
	return nil
}

func (mgr *deploymentManager) updateDeployment(
	ctx context.Context,
	old *appsv1beta2.Deployment,
) error {
	// We take a context here for the future when we can pass it along to kubernetes, but for our purposes, once the
	// operation starts, it can't be cancelled.

	oldJson, err := json.Marshal(old)
	if err != nil {
		return err
	}

	newJson, err := json.Marshal(mgr.applyUpdates(old.DeepCopy()))
	if err != nil {
		return err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(oldJson, newJson, appsv1beta2.Deployment{})
	if err != nil {
		return err
	}

	if len(patch) == 0 || string(patch) == "{}" {
		log.Printf("Deployment up to date for Action: %s", mgr.action.Name)
		return nil
	}

	log.Printf("Applying patch: %s", string(patch))

	result, err := mgr.kubeClient.AppsV1beta2().
		Deployments(mgr.namespace).
		Patch(old.Name, types.StrategicMergePatchType, patch)

	if err != nil {
		return err
	}

	log.Printf("Updated deployment for Action '%s': %v", mgr.action.Name, result)
	return nil
}

func (mgr *deploymentManager) applyUpdates(
	deployment *appsv1beta2.Deployment,
) *appsv1beta2.Deployment {
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = map[string]string{}
	}

	deployment.Spec.Template.Annotations["twine.ai/twine-action-id"] = mgr.action.Id
	deployment.Spec.Replicas = int32Ptr(1)
	deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
		{
			Name: mgr.actionVolumeName(),
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	mgr.populateSetupContainer(&deployment.Spec.Template.Spec.InitContainers[0])
	mgr.populateServerContainer(&deployment.Spec.Template.Spec.Containers[0])

	return deployment
}

func (mgr *deploymentManager) populateSetupContainer(container *corev1.Container) {
	container.Name = "setup"
	container.Image = mgr.setupImageName()
	container.ImagePullPolicy = corev1.PullIfNotPresent
	container.VolumeMounts = []corev1.VolumeMount{
		{
			Name:      mgr.actionVolumeName(),
			MountPath: mgr.actionVolumePath(),
		},
	}
	container.Env = []corev1.EnvVar{
		{
			Name:  "BUCKET_NAME",
			Value: mgr.bucketName,
		},
		{
			Name:  "ACTION_DIR",
			Value: mgr.actionVolumePath(),
		},
		{
			Name:  "ACTIONS",
			Value: mgr.action.Name,
		},
	}
}

func (mgr *deploymentManager) populateServerContainer(container *corev1.Container) {
	container.Name = "actionserver"
	container.Image = mgr.imageName()
	container.ImagePullPolicy = corev1.PullIfNotPresent
	container.VolumeMounts = []corev1.VolumeMount{
		{
			Name:      mgr.actionVolumeName(),
			MountPath: mgr.actionVolumePath(),
		},
	}
	container.Env = []corev1.EnvVar{
		{
			Name:  "PORT",
			Value: "8080",
		},
		{
			Name:  "ACTION_DIR",
			Value: mgr.actionVolumePath(),
		},
	}
	container.Ports = []corev1.ContainerPort{
		{
			Name:          "grpc",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: 8080,
		},
	}
}

func (mgr *deploymentManager) buildBaseDeployment() *appsv1beta2.Deployment {
	labels := mgr.deploymentLabels()

	return &appsv1beta2.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mgr.deploymentName(),
			Namespace: mgr.namespace,
			Labels:    labels,
		},
		Spec: appsv1beta2.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   mgr.namespace,
					Labels:      labels,
					Annotations: map[string]string{},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name: "setup",
						},
					},
					Containers: []corev1.Container{
						{
							Name: "actionserver",
						},
					},
				},
			},
		},
	}
}
