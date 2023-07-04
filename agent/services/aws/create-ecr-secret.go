package aws

import (
	"fmt"
	"log"

	"github.com/Humalect/humalect-core/agent/constants"
	"github.com/Humalect/humalect-core/agent/services/k8s"
	"github.com/Humalect/humalect-core/agent/utils"
	"k8s.io/client-go/kubernetes"
)

func CreateEcrSecret(params constants.ParamsConfig, clientSet *kubernetes.Clientset) (string, error) {
	ecrCredentials, _ := utils.UnmarshalStrings(params.EcrCredentials).(constants.EcrCredentials)
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
