package services

import (
	"context"
	"flag"
	"fmt"

	"github.com/Humalect/humalect-core/agent/constants"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func CreateK8sApplication(secretsProvider string, awsSecretCredentials constants.AwsSecretCredentials, azureSecretCredentials constants.AzureSecretCredentials, k8sAppName string, managedBy string, cloudRegion string, cloudProvider string,
	k8sResourcesIdentifier string, deploymentYamlManifest constants.DeploymentYamlManifestType, serviceYamlManifest constants.ServiceYamlManifestType, ingressYamlManifest constants.IngressYamlManifestType, secretManagerName string, azureVaultToken string, azureVaultName string, namespace string, webhookEndpoint string, webhookData string, deploymentId string,
) (string, error) {
	// var kubeconfig *string
	// if home := os.Getenv("HOME"); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }

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
					"app.kubernetes.io/name":       k8sAppName,
					"app.kubernetes.io/instance":   k8sAppName,
					"app.kubernetes.io/part-of":    "humalect-core",
					"app.kubernetes.io/managed-by": managedBy,
					"app.kubernetes.io/created-by": "humalect-core",
					"deploymentId":                 deploymentId,
				},
				"name": k8sAppName,
				"finalizers": []interface{}{
					"finalizers.humalect.com/application",
				},
			},
			"spec": map[string]interface{}{
				"secretsProvider":        secretsProvider,
				"awsSecretCredentials":   awsSecretCredentials,
				"azureSecretCredentials": azureSecretCredentials,
				"cloudRegion":            cloudRegion,
				"cloudProvider":          cloudProvider,
				"k8sResourcesIdentifier": k8sResourcesIdentifier,
				"deploymentYamlManifest": deploymentYamlManifest,
				"serviceYamlManifest":    serviceYamlManifest,
				"ingressYamlManifest":    ingressYamlManifest,
				"secretManagerName":      secretManagerName,
				"managedBy":              managedBy,
				"azureVaultToken":        azureVaultToken,
				"azureVaultName":         azureVaultName,
				"namespace":              namespace,
				"webhookEndpoint":        webhookEndpoint,
				"webhookData":            webhookData,
			},
		},
	}

	// create the custom resource in the specified namespace
	ctx := context.TODO()
	existingResource, err := dynamicClient.Resource(applicationGVR).Namespace(namespace).Get(ctx, applicationInstance.GetName(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			applicationResource, err := dynamicClient.Resource(applicationGVR).Namespace(namespace).Create(ctx, applicationInstance, metav1.CreateOptions{})
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
	updatedResource, err := dynamicClient.Resource(applicationGVR).Namespace(namespace).Update(ctx, existingResource, metav1.UpdateOptions{})
	if err != nil {
		fmt.Println(err)
		return "", err
		// panic(err.Error())
	}
	fmt.Printf("Created custom resource %s in namespace %s\n", updatedResource.GetName(), namespace)
	return updatedResource.GetName(), nil
}
