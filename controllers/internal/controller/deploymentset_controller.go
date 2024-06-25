/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	k8sv1 "github.com/Humalect/humalect-core/api/v1"
	constants "github.com/Humalect/humalect-core/internal/controller/constants"
	helpers "github.com/Humalect/humalect-core/internal/controller/helpers"
	"github.com/joho/godotenv"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// DeploymentSetReconciler reconciles a DeploymentSet object
type DeploymentSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=k8s.humalect.com,resources=deploymentsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8s.humalect.com,resources=deploymentsets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8s.humalect.com,resources=deploymentsets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DeploymentSet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
type Object interface {
	metav1.Object
	runtime.Object
}

func sendDeploymentJobCreatedWebhook(deploymentSet k8sv1.DeploymentSet, success bool) {
	helpers.SendWebhook(deploymentSet.Spec.WebhookEndpoint, deploymentSet.Spec.WebhookData, success, constants.DeploymentJobCreated)
}

func createEmptyObject(obj Object) Object {
	objType := reflect.TypeOf(obj)
	emptyObj := reflect.New(objType.Elem()).Interface().(Object)
	return emptyObj
}

const (
	deploymentSetFinalizer = "finalizers.humalect.com/deploymentset"
)

func (r *DeploymentSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	godotenv.Load()

	deploymentSet := &k8sv1.DeploymentSet{}
	err := r.Get(ctx, req.NamespacedName, deploymentSet)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf("log for <depid:%s> <pipeid:%s> DeploymentSet not found, ignoring reconcile", deploymentSet.Spec.DeploymentId, deploymentSet.Spec.PipelineId))
			return ctrl.Result{}, nil
		}
		log.Error(err, fmt.Sprintf("log for <depid:%s> <pipeid:%s> ERROR: Failed to get Application, %v", deploymentSet.Spec.DeploymentId, deploymentSet.Spec.PipelineId, err))
		return ctrl.Result{}, err
	}

	if !deploymentSet.DeletionTimestamp.IsZero() {
		deploymentSet.ObjectMeta.Finalizers = removeString(deploymentSet.ObjectMeta.Finalizers, deploymentSetFinalizer)
		if err := r.Update(ctx, deploymentSet); err != nil {
			log.Error(err, fmt.Sprintf("log for <depid:%s> <pipeid:%s> ERROR: There is some error, %v", deploymentSet.Spec.DeploymentId, deploymentSet.Spec.PipelineId, err))
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	ingressYamlManifest, err := json.Marshal(deploymentSet.Spec.IngressYamlManifest)
	if err != nil {
		deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)
		sendDeploymentJobCreatedWebhook(*deploymentSet, false)
	}
	applicationSecretsConfig, err := json.Marshal(deploymentSet.Spec.ApplicationSecretsConfig)
	if err != nil {
		deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)
		sendDeploymentJobCreatedWebhook(*deploymentSet, false)
	}

	buildSecretsConfig, err := json.Marshal(deploymentSet.Spec.BuildSecretsConfig)
	if err != nil {
		deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)
		sendDeploymentJobCreatedWebhook(*deploymentSet, false)
	}
	serviceYamlManifest, err := json.Marshal(deploymentSet.Spec.ServiceYamlManifest)
	if err != nil {
		deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)
		sendDeploymentJobCreatedWebhook(*deploymentSet, false)
	}
	deploymentYamlManifest, err := json.Marshal(deploymentSet.Spec.DeploymentYamlManifest)
	if err != nil {
		deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)
		sendDeploymentJobCreatedWebhook(*deploymentSet, false)
	}
	DockerManifest, err := json.Marshal(deploymentSet.Spec.DockerManifest)
	if err != nil {
		deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)
		sendDeploymentJobCreatedWebhook(*deploymentSet, false)
	}
	log.Info(fmt.Sprintf("log for <depid:%s> <pipeid:%s> Creating Job", deploymentSet.Spec.DeploymentId, deploymentSet.Spec.PipelineId))
	agentImageTag, exists := os.LookupEnv("AGENT_IMAGE_TAG")

	if !exists {
		agentImageTag = "latest"
	}
	backoffLimit := int32(0)

	ecrCredentials, err := json.Marshal(deploymentSet.Spec.EcrCredentials)
	if err != nil {
		deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)
		sendDeploymentJobCreatedWebhook(*deploymentSet, false)
	}

	acrCredentials, err := json.Marshal(deploymentSet.Spec.AcrCredentials)
	if err != nil {
		deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)
		sendDeploymentJobCreatedWebhook(*deploymentSet, false)
	}

	dockerHubCredentials, err := json.Marshal(deploymentSet.Spec.DockerHubCredentials)
	if err != nil {
		deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)
		sendDeploymentJobCreatedWebhook(*deploymentSet, false)
	}

	awsSecretCredentials, err := json.Marshal(deploymentSet.Spec.AwsSecretCredentials)
	if err != nil {
		deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)
		sendDeploymentJobCreatedWebhook(*deploymentSet, false)
	}

	azureVaultCredentials, err := json.Marshal(deploymentSet.Spec.AzureVaultCredentials)
	if err != nil {
		deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)
		sendDeploymentJobCreatedWebhook(*deploymentSet, false)
	}

	jobObj := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-ds-%s-%s",
				deploymentSet.Spec.ManagedBy[:int(math.Min(float64(len(deploymentSet.Spec.ManagedBy)), float64(10)))],
				deploymentSet.Spec.CommitId[:int(math.Min(float64(len(deploymentSet.Spec.CommitId)), float64(5)))],
				deploymentSet.Spec.DeploymentId[:int(math.Min(float64(len(deploymentSet.Spec.DeploymentId)), float64(7)))]),
			Namespace: "humalect",
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            fmt.Sprintf("%s-ds-%s", deploymentSet.Spec.ManagedBy, deploymentSet.Spec.CommitId),
							Image:           "public.ecr.aws/survo/core-agent:" + agentImageTag,
							ImagePullPolicy: corev1.PullPolicy("Always"),
							Args: []string{
								fmt.Sprintf("--artifactsRegistryProvider=%s", deploymentSet.Spec.ArtifactsRegistryProvider),
								fmt.Sprintf("--secretsProvider=%s", deploymentSet.Spec.SecretsProvider),
								fmt.Sprintf("--ecrCredentials=%s", ecrCredentials),
								fmt.Sprintf("--acrCredentials=%s", acrCredentials),
								fmt.Sprintf("--dockerHubCredentials=%s", dockerHubCredentials),
								fmt.Sprintf("--awsSecretCredentials=%s", awsSecretCredentials),
								fmt.Sprintf("--azureVaultCredentials=%s", azureVaultCredentials),
								fmt.Sprintf("--cloudProvider=%s", deploymentSet.Spec.CloudProvider),
								fmt.Sprintf("--sourceCodeRepositoryName=%s", deploymentSet.Spec.SourceCodeRepositoryName),
								fmt.Sprintf("--sourceCodeProvider=%s", deploymentSet.Spec.SourceCodeProvider),
								fmt.Sprintf("--sourceCodeToken=%s", deploymentSet.Spec.SourceCodeToken),
								fmt.Sprintf("--sourceCodeOrgName=%s", deploymentSet.Spec.SourceCodeOrgName),
								fmt.Sprintf("--commitId=%s", deploymentSet.Spec.CommitId),
								fmt.Sprintf("--dockerManifest=%s", DockerManifest),
								fmt.Sprintf("--k8sAppName=%s", deploymentSet.Spec.K8sAppName),
								fmt.Sprintf("--artifactsRepositoryName=%s", deploymentSet.Spec.ArtifactsRepositoryName),
								fmt.Sprintf("--useDockerFromCodeFlag=%t", deploymentSet.Spec.UseDockerFromCodeFlag),
								fmt.Sprintf("--managedBy=%s", deploymentSet.Spec.ManagedBy),
								fmt.Sprintf("--cloudRegion=%s", deploymentSet.Spec.CloudRegion),
								fmt.Sprintf("--k8sResourcesIdentifier=%s", deploymentSet.Spec.K8sResourcesIdentifier),
								fmt.Sprintf("--buildSecretsConfig=%s", buildSecretsConfig),
								fmt.Sprintf("--applicationSecretsConfig=%s", applicationSecretsConfig),
								fmt.Sprintf("--namespace=%s", deploymentSet.Spec.Namespace),
								fmt.Sprintf("--deploymentId=%s", deploymentSet.Spec.DeploymentId),
								fmt.Sprintf("--ingressYamlManifest=%s", ingressYamlManifest),
								fmt.Sprintf("--serviceYamlManifest=%s", serviceYamlManifest),
								fmt.Sprintf("--deploymentYamlManifest=%s", deploymentYamlManifest),
								fmt.Sprintf("--pipelineId=%s", deploymentSet.Spec.PipelineId),
								fmt.Sprintf("--webhookEndpoint=%s", deploymentSet.Spec.WebhookEndpoint),
								fmt.Sprintf("--webhookData=%s", helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, true)),
							},
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: "humalect-sa",
				},
			},
			BackoffLimit: &backoffLimit,
		},
	}
	emptyObj := createEmptyObject(jobObj)

	jobObj.SetNamespace("humalect")

	if err := r.Get(ctx, client.ObjectKey{Name: jobObj.GetName(), Namespace: "humalect"}, emptyObj); err != nil {
		if errors.IsNotFound(err) {

			// controllerRef := metav1.NewControllerRef(deploymentSet, k8sv1.GroupVersion.WithKind("DeploymentSet"))
			// jobObj.SetOwnerReferences(append(jobObj.GetOwnerReferences(), *controllerRef))

			// if err := r.Create(ctx, jobObj); err != nil {
			// 	log.Error(err, "Failed to create", reflect.TypeOf(jobObj).String(), jobObj.GetName())
			// 	sendDeploymentJobCreatedWebhook(*deploymentSet, false)
			// 	return ctrl.Result{}, err
			// }
			// sendDeploymentJobCreatedWebhook(*deploymentSet, true)
			// log.Info("Created Resource", reflect.TypeOf(jobObj).String(), jobObj.GetName())
			// var kubeconfig *string
			// home := homedir.HomeDir()
			// kubeconfig = flag.String("kubecon1fig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")

			// Set up the default Kubeconfig file path.
			// if ; home != "" {
			// } else {
			// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
			// }
			// flag.Parse()

			config := helpers.GetK8sConfig()

			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				log.Error(err, fmt.Sprintf("log for <depid:%s> <pipeid:%s> ERROR: Error creating clientset, %v", deploymentSet.Spec.DeploymentId, deploymentSet.Spec.PipelineId, err))
				deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)

				sendDeploymentJobCreatedWebhook(*deploymentSet, false)
			}
			jobClient := clientset.BatchV1().Jobs("humalect")
			jobObj.SetNamespace("humalect")
			_, err = jobClient.Create(context.Background(), jobObj, metav1.CreateOptions{})
			if err != nil {
				deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)

				sendDeploymentJobCreatedWebhook(*deploymentSet, false)
			}
			deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, true)

			sendDeploymentJobCreatedWebhook(*deploymentSet, true)
		} else {
			deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)

			sendDeploymentJobCreatedWebhook(*deploymentSet, false)
			deploymentSet.Spec.WebhookData = helpers.UpdateStatusData(deploymentSet.Spec.WebhookData, constants.DeploymentJobCreated, false)

			log.Error(err, fmt.Sprintf("log for <depid:%s> <pipeid:%s> ERROR: Failed to get, %v", deploymentSet.Spec.DeploymentId, deploymentSet.Spec.PipelineId, err), reflect.TypeOf(emptyObj).String(), jobObj.GetName())

			return ctrl.Result{}, err
		}
	} else {
		// sendDeploymentJobCreatedWebhook(*deploymentSet, true)
		log.Info(fmt.Sprintf("log for <depid:%s> <pipeid:%s> Job already exists, skipping creation", deploymentSet.Spec.DeploymentId, deploymentSet.Spec.PipelineId), reflect.TypeOf(jobObj).String(), jobObj.GetName())
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sv1.DeploymentSet{}).
		Complete(r)
}
