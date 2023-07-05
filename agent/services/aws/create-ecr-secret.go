package aws

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Humalect/humalect-core/agent/constants"
	"github.com/Humalect/humalect-core/agent/services/k8s"
	"k8s.io/client-go/kubernetes"
)

func CreateEcrSecret(params constants.ParamsConfig, clientSet *kubernetes.Clientset) (string, error) {
	var ecrCredentials constants.EcrCredentials
	_ = json.Unmarshal([]byte(params.EcrCredentials), &ecrCredentials)
	ecrToken, err := getEcrLoginToken(ecrCredentials.AccessKey, ecrCredentials.SecretKey, ecrCredentials.Region)
	if err != nil {
		log.Fatalf("Error getting ECR token: %v", err)
		return "", err
	}
	secretData := fmt.Sprintf(`{  
			"auths": {  
				"%s": {  
					"username":"AWS",
					"password": "%s"  
				}  
			}  
		}`, ecrCredentials.RegistryUrl, ecrToken)
	return k8s.CreateSecret(map[string]string{
		".dockerconfigjson": secretData,
	}, params, clientSet)
}
