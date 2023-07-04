package k8s

import (
	"context"
	"fmt"
	"math"

	"github.com/Humalect/humalect-core/agent/constants"
	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateSecret(secretData map[string]string, params constants.ParamsConfig, clientset *kubernetes.Clientset) (string, error) {
	dockerRegistrySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-artifact-%s-%s",
				params.ManagedBy[:int(math.Min(float64(len(params.ManagedBy)), float64(10)))],
				params.CommitId[:int(math.Min(float64(len(params.CommitId)), float64(5)))],
				params.DeploymentId[:int(math.Min(float64(len(params.DeploymentId)), float64(7)))]),
			Namespace: params.Namespace,
			Labels: map[string]string{
				"app": fmt.Sprintf("%s-artifact-%s-%s",
					params.ManagedBy[:int(math.Min(float64(len(params.ManagedBy)), float64(10)))],
					params.CommitId[:int(math.Min(float64(len(params.CommitId)), float64(5)))],
					params.DeploymentId[:int(math.Min(float64(len(params.DeploymentId)), float64(7)))]),
				"deploymentId": params.DeploymentId,
				"managedBy":    params.ManagedBy,
				"commitId":     params.CommitId,
			},
		},
		Type:       corev1.SecretTypeDockerConfigJson,
		StringData: secretData,
	}

	createdSecret, err := clientset.CoreV1().Secrets("humalect").Create(context.Background(), dockerRegistrySecret, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	return createdSecret.Name, nil
}
