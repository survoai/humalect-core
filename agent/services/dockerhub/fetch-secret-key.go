package dockerhub

import (
	"errors"
	"log"

	"github.com/Humalect/humalect-core/agent/constants"
	"github.com/Humalect/humalect-core/agent/services/aws"
	"github.com/Humalect/humalect-core/agent/services/azure"
	"github.com/Humalect/humalect-core/agent/utils"
)

func FetchDockerHubSecretKey(params constants.ParamsConfig) (string, error) {
	dockerHubCreds := utils.UnmarshalStrings(params.DockerHubCredentials).(constants.DockerHubCredentials)
	if params.SecretsProvider == constants.CloudIdAWS || (params.SecretsProvider == "" && params.CloudProvider == constants.CloudIdAWS) {
		awsSecretCredentials := utils.UnmarshalStrings(params.AwsSecretCredentials).(constants.AwsSecretCredentials)
		return aws.GetSecretValue(dockerHubCreds.SecretName, awsSecretCredentials.AccessKey, awsSecretCredentials.SecretKey, awsSecretCredentials.Region)
	} else if params.SecretsProvider == constants.CloudIdAzure || (params.SecretsProvider == "" && params.CloudProvider == constants.CloudIdAzure) {
		azureVaultCredentials := utils.UnmarshalStrings(params.AzureVaultCredentials).(constants.AzureVaultCredentials)
		secretData, err := azure.GetSecretValue(azureVaultCredentials.Token, azureVaultCredentials.Name, dockerHubCreds.SecretName)
		if err != nil {
			log.Fatalf("Error getting dockerhub secret: %v", err)
			return "", err
		}
		return secretData, nil
	} else {
		return "", errors.New("No credentials provided")
	}
}
