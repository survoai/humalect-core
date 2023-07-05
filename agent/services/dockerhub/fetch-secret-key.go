package dockerhub

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/Humalect/humalect-core/agent/constants"
	"github.com/Humalect/humalect-core/agent/services/aws"
	"github.com/Humalect/humalect-core/agent/services/azure"
)

func FetchDockerHubSecretKey(params constants.ParamsConfig) (string, error) {
	var dockerHubCreds constants.DockerHubCredentials
	json.Unmarshal([]byte(params.DockerHubCredentials), dockerHubCreds)
	if params.SecretsProvider == constants.CloudIdAWS || (params.SecretsProvider == "" && params.CloudProvider == constants.CloudIdAWS) {
		var awsSecretCredentials constants.AwsSecretCredentials
		json.Unmarshal([]byte(params.AwsSecretCredentials), awsSecretCredentials)
		region := ""
		if params.SecretsProvider == "" && params.CloudProvider == constants.CloudIdAWS {
			region = params.CloudRegion
		} else {
			region = awsSecretCredentials.Region
		}
		return aws.GetSecretValue(dockerHubCreds.SecretName, awsSecretCredentials.AccessKey, awsSecretCredentials.SecretKey, region, params.CloudProvider)
	} else if params.SecretsProvider == constants.CloudIdAzure || (params.SecretsProvider == "" && params.CloudProvider == constants.CloudIdAzure) {
		var azureVaultCredentials constants.AzureVaultCredentials
		json.Unmarshal([]byte(params.AzureVaultCredentials), azureVaultCredentials)
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
