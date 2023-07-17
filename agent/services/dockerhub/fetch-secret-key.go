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
	_ = json.Unmarshal([]byte(params.DockerHubCredentials), &dockerHubCreds)
	if params.SecretsProvider == constants.CloudIdAWS || (params.SecretsProvider == "" && params.CloudProvider == constants.CloudIdAWS) {
		var awsSecretCredentials constants.AwsSecretCredentials
		_ = json.Unmarshal([]byte(params.AwsSecretCredentials), &awsSecretCredentials)
		region := ""
		if params.SecretsProvider == "" && params.CloudProvider == constants.CloudIdAWS {
			region = params.CloudRegion
		} else {
			region = awsSecretCredentials.Region
		}
		secretData, err := aws.GetSecretValue(dockerHubCreds.SecretName, awsSecretCredentials.AccessKey, awsSecretCredentials.SecretKey, region, params.CloudProvider)
		return secretData[constants.RegistryIdDockerhub], err
	} else if params.SecretsProvider == constants.CloudIdAzure || (params.SecretsProvider == "" && params.CloudProvider == constants.CloudIdAzure) {
		var azureVaultCredentials constants.AzureVaultCredentials
		_ = json.Unmarshal([]byte(params.AzureVaultCredentials), &azureVaultCredentials)
		secretData, err := azure.GetSecretValue(azureVaultCredentials.Token, azureVaultCredentials.Name, dockerHubCreds.SecretName)
		if err != nil {
			log.Fatalf("Error getting dockerhub secret: %v", err)
			return "", err
		}
		return secretData[constants.RegistryIdDockerhub], nil
	} else {
		return "", errors.New("No credentials provided")
	}
}
