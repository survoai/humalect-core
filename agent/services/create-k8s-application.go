package services

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/Humalect/humalect-core/agent/constants"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func CreateK8sApplication(params *constants.ParamsConfig, kanikoJobResources CreateJobConfig, webhookData string) (string, error) {
	var awsSecretCredentials constants.AwsSecretCredentials
	var azureVaultCredentials constants.AzureVaultCredentials
	var deploymentYamlManifest constants.DeploymentYamlManifestType
	var serviceYamlManifest constants.ServiceYamlManifestType
	var ingressYamlManifest constants.IngressYamlManifestType
	json.Unmarshal([]byte(params.AwsSecretCredentials), &awsSecretCredentials)
	json.Unmarshal([]byte(params.AzureVaultCredentials), &azureVaultCredentials)
	json.Unmarshal([]byte(params.DeploymentYamlManifest), &deploymentYamlManifest)
	json.Unmarshal([]byte(params.ServiceYamlManifest), &serviceYamlManifest)
	json.Unmarshal([]byte(params.IngressYamlManifest), &ingressYamlManifest)

	deploymentYamlManifest.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: kanikoJobResources.CloudProviderSecretName}}

	flag.Parse()
	config := GetK8sConfig()
	// create the dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// specify the custom resource definition
	applicationGVR := schema.GroupVersionResource{
		Group:    "k8s.humalect.com",
		Version:  "v1",
		Resource: "applications",
	}

	applicationInstance := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "k8s.humalect.com/v1",
			"kind":       "Application",
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":       params.K8sAppName,
					"app.kubernetes.io/instance":   params.K8sAppName,
					"app.kubernetes.io/part-of":    "humalect-core",
					"app.kubernetes.io/managed-by": params.ManagedBy,
					"app.kubernetes.io/created-by": "humalect-core",
					"deploymentId":                 params.DeploymentId,
					"managedBy":                    params.ManagedBy,
				},
				"name": params.K8sAppName,
				"finalizers": []interface{}{
					"finalizers.humalect.com/application",
				},
			},
			"spec": map[string]interface{}{
				"secretsProvider":        params.SecretsProvider,
				"awsSecretCredentials":   awsSecretCredentials,
				"azureVaultCredentials":  azureVaultCredentials,
				"cloudRegion":            params.CloudRegion,
				"cloudProvider":          params.CloudProvider,
				"k8sResourcesIdentifier": params.K8sResourcesIdentifier,
				"deploymentYamlManifest": deploymentYamlManifest,
				"serviceYamlManifest":    serviceYamlManifest,
				"ingressYamlManifest":    ingressYamlManifest,
				"secretManagerName":      params.SecretManagerName,
				"managedBy":              params.ManagedBy,
				"namespace":              params.Namespace,
				"webhookEndpoint":        params.WebhookEndpoint,
				"webhookData":            webhookData,
			},
		},
	}

	// create the custom resource in the specified namespace
	ctx := context.TODO()
	existingResource, err := dynamicClient.Resource(applicationGVR).Namespace(params.Namespace).Get(ctx, applicationInstance.GetName(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			applicationResource, err := dynamicClient.Resource(applicationGVR).Namespace(params.Namespace).Create(ctx, applicationInstance, metav1.CreateOptions{})
			if err != nil {
				fmt.Println(err)
				return "", err
				// panic(err.Error())
			}
			return applicationResource.GetName(), nil
		} else {
			return "", err
		}
	}
	applicationInstance.SetResourceVersion(existingResource.GetResourceVersion())
	for key, value := range applicationInstance.Object {
		existingResource.Object[key] = value
	}
	updatedResource, err := dynamicClient.Resource(applicationGVR).Namespace(params.Namespace).Update(ctx, existingResource, metav1.UpdateOptions{})
	if err != nil {
		fmt.Println(err)
		return "", err
		// panic(err.Error())
	}
	fmt.Printf("Created custom resource %s in namespace %s\n", updatedResource.GetName(), params.Namespace)
	return updatedResource.GetName(), nil
}
