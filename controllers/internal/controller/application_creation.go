package controller

import (
	"context"
	"log"
	"regexp"
	"strings"

	k8sv1 "github.com/Humalect/humalect-core/api/v1"
	constants "github.com/Humalect/humalect-core/internal/controller/constants"
	helpers "github.com/Humalect/humalect-core/internal/controller/helpers"
	cloudhelpers "github.com/Humalect/humalect-core/internal/controller/helpers/cloud"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ApplicationReconciler) handleCreation(ctx context.Context, application *k8sv1.Application, DeploymentYamlManifest k8sv1.DeploymentYamlManifestType, ServiceYamlManifest k8sv1.ServiceYamlManifestType, IngressYamlManifest k8sv1.IngressYamlManifestType, Namespace string) (ctrl.Result, error) {
	// Create a slice of Object to store the objects you want to pass
	objects := []helpers.Object{
		&appsv1.Deployment{
			ObjectMeta: DeploymentYamlManifest.Metadata,
			Spec:       DeploymentYamlManifest.Spec,
		},
		&corev1.Service{
			ObjectMeta: ServiceYamlManifest.Metadata,
			Spec:       ServiceYamlManifest.Spec,
		},
		&networkingv1.Ingress{
			ObjectMeta: IngressYamlManifest.Metadata,
			Spec:       IngressYamlManifest.Spec,
		},
	}
	// Check your specific condition
	// TODO send deployment id here so that secret can be created with every deployment
	if application.Spec.SecretManagerName != "" {
		regexPattern := "[^a-z0-9-.]+"
		regex, err := regexp.Compile(regexPattern)
		secretMetadataObject := metav1.ObjectMeta{
			Name: strings.Trim(regex.ReplaceAllString(strings.ToLower(application.Spec.SecretManagerName), "-"), "-."),
			Labels: map[string]string{
				"managedBy":  application.Spec.ManagedBy,
				"identifier": application.Spec.K8sResourcesIdentifier,
				// "deploymentId": application.Spec.DeploymentId,
			},
			Namespace: Namespace,
		}
		SecretStringData, err := cloudhelpers.GetCloudSecretMap(application.Spec.SecretsProvider, application.Spec.AwsSecretCredentials.AccessKey,
			application.Spec.AwsSecretCredentials.SecretKey, application.Spec.AwsSecretCredentials.Region, application.Spec.AzureVaultCredentials.Token,
			application.Spec.SecretManagerName, application.Spec.CloudRegion, application.Spec.AzureVaultCredentials.Name, application.Spec.CloudProvider)
		if err != nil {
			log.Fatal(err.Error())
			application.Spec.WebhookData = helpers.UpdateStatusData(application.Spec.WebhookData, constants.CreatedKubernetesResources, false)
			helpers.SendWebhook(application.Spec.WebhookEndpoint, application.Spec.WebhookData, false, constants.CreatedKubernetesResources)
		}
		// Append the secret object to the objects slice
		objects = append(objects, &corev1.Secret{
			ObjectMeta: secretMetadataObject,
			StringData: SecretStringData,
		})
	}

	return helpers.CreateK8sResource(ctx, application, application.GetNamespace(), (*helpers.ApplicationReconciler)(r), objects...)
}
