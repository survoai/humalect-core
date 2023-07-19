package dockerhub

import (
	"fmt"
	"log"
	"strings"

	"github.com/Humalect/humalect-core/agent/constants"
	"github.com/Humalect/humalect-core/agent/services/k8s"
	"k8s.io/client-go/kubernetes"
)

func CreateSecret(params constants.ParamsConfig, clientSet *kubernetes.Clientset) (string, error) {
	secretKey, err := FetchDockerHubSecretKey(params)
	if err != nil {
		log.Fatalf("Error getting dockerhub secret: %v", err)
		return "", err
	}
	secretData := strings.Join(strings.Fields(fmt.Sprintf(`{
			"auths": {
				"https://index.docker.io/v1/": {
					"auth": "%s"
				}
			}
		}`, secretKey)), "")

	return k8s.CreateSecret(map[string]string{
		".dockerconfigjson": secretData,
	}, params, clientSet, "humalect")
}
