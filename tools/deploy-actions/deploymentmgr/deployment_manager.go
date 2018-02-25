package deploymentmgr

import (
	"context"
	"encoding/json"
	"log"
	"os/exec"

	"github.com/pkg/errors"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	kuberesources "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"

	"fmt"
	"strings"
	"sync"

	"sort"

	"github.com/twineai/actions/tools/deploy-actions/action"
)

var serverOnce sync.Once
var _serverVersion string

func getServerVersion() string {
	serverOnce.Do(func() {
		_serverVersion = FlagActionServerVersion
		if _serverVersion == "" {
			cmd := fmt.Sprintf(`gcloud container images list-tags %s --filter="tags!=latest" --limit=1 | tail -n 1 | awk '{ print $2; }'`, FlagActionServerImageName)
			out, err := exec.Command("bash", "-ec", cmd).Output()
			if err != nil {
				log.Fatalf("Unable to get server version: %s", err)
			}

			rawTags := strings.Split(strings.TrimSpace(string(out)), ",")
			for _, tag := range rawTags {
				normalized := strings.TrimSpace(tag)
				if normalized != "latest" {
					log.Printf("No server version provided, using '%s'", normalized)
					_serverVersion = normalized
					return
				}
			}
		}

		log.Printf("Using server version: %s", _serverVersion)
	})

	return _serverVersion
}

func NewDeploymentManager(
	kubeClient kubernetes.Interface,
	namespace string,
	bucketName string,
	actions []action.Action,
) *deploymentManager {
	versionToUse := getServerVersion()

	return &deploymentManager{
		kubeClient:           kubeClient,
		actions:              actions,
		namespace:            namespace,
		bucketName:           bucketName,
		serverImageName:      FlagActionServerImageName,
		serverSetupImageName: FlagActionServerSetupImageName,
		serverVersion:        versionToUse,
	}
}

type deploymentManager struct {
	kubeClient kubernetes.Interface

	actions              []action.Action
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
			err = errors.Wrapf(err, "error configuring deployment for actions '%x'", mgr.actions)
		}
	}()

	old, err := mgr.getCurrentDeployment(ctx)
	if err != nil && !kubeerrors.IsNotFound(err) {
		return err
	}

	if kubeerrors.IsNotFound(err) {
		log.Printf("Creating deployment for actions: %x", mgr.actions)
		return mgr.createDeployment(ctx)
	} else {
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

	txt, err := json.Marshal(result)
	if err != nil {
		return err
	}

	log.Printf("Created deployment for actions '%s': %s", mgr.actions, string(txt))
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
		log.Printf("Deployment up to date for Actions: %x", mgr.actions)
		return nil
	}

	log.Printf("Applying patch: %s", string(patch))

	result, err := mgr.kubeClient.AppsV1beta2().
		Deployments(mgr.namespace).
		Patch(old.Name, types.StrategicMergePatchType, patch)

	if err != nil {
		return err
	}

	txt, err := json.Marshal(result)
	if err != nil {
		return err
	}

	log.Printf("Updated deployment for Action '%s': %s", mgr.actions, string(txt))
	return nil
}

func (mgr *deploymentManager) applyUpdates(
	deployment *appsv1beta2.Deployment,
) *appsv1beta2.Deployment {
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = map[string]string{}
	}

	deployment.Spec.Replicas = int32Ptr(int32(FlagInstanceCount))
	deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
		{
			Name: mgr.actionVolumeName(),
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "github-ssh-keys",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  "builder-ssh-key",
					DefaultMode: int32Ptr(0600),
				},
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
		{
			Name:      "github-ssh-keys",
			MountPath: "/root/.ssh",
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
	}

	actionNames := make([]string, len(mgr.actions))
	for i, action := range mgr.actions {
		fmt.Printf("Handling action: %v\n", action)
		actionNames[i] = action.Name
	}

	sort.Strings(actionNames)
	container.Args = actionNames

	fmt.Printf("Container args: %v\n", container.Args)
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
		{
			Name:  "LOG_LEVEL",
			Value: "debug",
		},
	}
	container.Ports = []corev1.ContainerPort{
		{
			Name:          "grpc",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: 8080,
		},
	}
	container.Resources = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    kuberesources.MustParse("500m"),
			corev1.ResourceMemory: kuberesources.MustParse("512Mi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    kuberesources.MustParse("125m"),
			corev1.ResourceMemory: kuberesources.MustParse("192Mi"),
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
					Volumes: []corev1.Volume{
						{
							Name: "github-ssh-keys",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "builder-ssh-key",
								},
							},
						},
					},
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
