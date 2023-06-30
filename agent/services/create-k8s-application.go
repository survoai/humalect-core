package services

import (
	"context"
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

func CreateK8sApplication(paramsConfig *constants.ParamsConfig, kanikoJobResources CreateJobConfig, webhookData string) (string, error) {
	// var kubeconfig *string
	// if home := os.Getenv("HOME"); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }

	awsSecretCredentials, err := GetAwsSecretCredentials(paramsConfig)
	if err != nil {
		return "", err
	}

	azureVaultCredentials, err := GetAzureVaultCredentials(paramsConfig)
	if err != nil {
		return "", err
	}

	deploymentYamlManifest, err := GetDeploymentYamlManifest(paramsConfig)
	if err != nil {
		return "", err
	}
	deploymentYamlManifest.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: kanikoJobResources.CloudProviderSecretName}}

	serviceYamlManifest, err := GetServiceYamlManifest(paramsConfig)
	if err != nil {
		return "", err
	}

	ingressYamlManifest, err := GetIngressYamlManifest(paramsConfig)
	if err != nil {
		return "", err
	}

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
					"app.kubernetes.io/name":       paramsConfig.K8sAppName,
					"app.kubernetes.io/instance":   paramsConfig.K8sAppName,
					"app.kubernetes.io/part-of":    "humalect-core",
					"app.kubernetes.io/managed-by": paramsConfig.ManagedBy,
					"app.kubernetes.io/created-by": "humalect-core",
					"deploymentId":                 paramsConfig.DeploymentId,
				},
				"name": paramsConfig.K8sAppName,
				"finalizers": []interface{}{
					"finalizers.humalect.com/application",
				},
			},
			"spec": map[string]interface{}{
				"secretsProvider":        paramsConfig.SecretsProvider,
				"awsSecretCredentials":   awsSecretCredentials,
				"azureVaultCredentials":  azureVaultCredentials,
				"cloudRegion":            paramsConfig.CloudRegion,
				"cloudProvider":          paramsConfig.CloudProvider,
				"k8sResourcesIdentifier": paramsConfig.K8sResourcesIdentifier,
				"deploymentYamlManifest": deploymentYamlManifest,
				"serviceYamlManifest":    serviceYamlManifest,
				"ingressYamlManifest":    ingressYamlManifest,
				"secretManagerName":      paramsConfig.SecretManagerName,
				"managedBy":              paramsConfig.ManagedBy,
				"azureVaultToken":        paramsConfig.AzureVaultToken,
				"azureVaultName":         paramsConfig.AzureVaultName,
				"namespace":              paramsConfig.Namespace,
				"webhookEndpoint":        paramsConfig.WebhookEndpoint,
				"webhookData":            webhookData,
			},
		},
	}

	// create the custom resource in the specified namespace
	ctx := context.TODO()
	existingResource, err := dynamicClient.Resource(applicationGVR).Namespace(paramsConfig.Namespace).Get(ctx, applicationInstance.GetName(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			applicationResource, err := dynamicClient.Resource(applicationGVR).Namespace(paramsConfig.Namespace).Create(ctx, applicationInstance, metav1.CreateOptions{})
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
	updatedResource, err := dynamicClient.Resource(applicationGVR).Namespace(paramsConfig.Namespace).Update(ctx, existingResource, metav1.UpdateOptions{})
	if err != nil {
		fmt.Println(err)
		return "", err
		// panic(err.Error())
	}
	fmt.Printf("Created custom resource %s in namespace %s\n", updatedResource.GetName(), paramsConfig.Namespace)
	return updatedResource.GetName(), nil
}
