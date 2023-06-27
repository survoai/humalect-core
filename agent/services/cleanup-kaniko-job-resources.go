package services

import (
	"context"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// TODO send webhook here
func CleanupKanikoJobResources(jobConfig CreateJobConfig) error {
	config := GetK8sConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
		// SendWebhook(params.WebhookEndpoint, params.WebhookData, false, constants.CreatedKanikoJob)
		panic(err)
	}
	if len(jobConfig.CloudProviderSecretName) != 0 {
		err := deleteSecret(clientset, jobConfig.CloudProviderSecretName, "humalect")
		if err != nil {
			return err
		}
	}
	if len(jobConfig.DockerFileConfigName) != 0 {
		err := deleteConfigMap(clientset, jobConfig.DockerFileConfigName, "humalect")
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteSecret(clientset *kubernetes.Clientset, secretName, namespace string) error {
	ctx := context.TODO()

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	err := clientset.CoreV1().Secrets(namespace).Delete(ctx, secretName, deleteOptions)
	if err != nil {
		return err
	}
	fmt.Printf("Secret '%s' in namespace '%s' deleted successfully.\n", secretName, namespace)
	return nil
}

func deleteConfigMap(clientset *kubernetes.Clientset, configMapName, namespace string) error {
	ctx := context.TODO()

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	err := clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, configMapName, deleteOptions)
	if err != nil {
		return err
	}
	fmt.Printf("ConfigMap '%s' in namespace '%s' deleted successfully.\n", configMapName, namespace)
	return nil
}
