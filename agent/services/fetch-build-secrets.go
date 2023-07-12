package services

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/Humalect/humalect-core/agent/constants"
	"github.com/Humalect/humalect-core/agent/services/aws"
	"github.com/Humalect/humalect-core/agent/services/azure"
)

func FetchBuildSecrets(params constants.ParamsConfig) (map[string]string, error) {
	var buildSecretsConfig []constants.SecretConfig
	json.Unmarshal([]byte(params.BuildSecretsConfig), &buildSecretsConfig)
	for _, secretConfig := range buildSecretsConfig {
		if params.SecretsProvider == constants.CloudIdAWS || (params.SecretsProvider == "" && params.CloudProvider == constants.CloudIdAWS) {
			var awsSecretCredentials constants.AwsSecretCredentials
			_ = json.Unmarshal([]byte(params.AwsSecretCredentials), &awsSecretCredentials)
			region := ""
			if params.SecretsProvider == "" && params.CloudProvider == constants.CloudIdAWS {
				region = params.CloudRegion
			} else {
				region = awsSecretCredentials.Region
			}
			secretData, err := aws.GetSecretValue(secretConfig.Name, awsSecretCredentials.AccessKey, awsSecretCredentials.SecretKey, region, params.CloudProvider)
			return secretData, err
		} else if params.SecretsProvider == constants.CloudIdAzure || (params.SecretsProvider == "" && params.CloudProvider == constants.CloudIdAzure) {
			var azureVaultCredentials constants.AzureVaultCredentials
			_ = json.Unmarshal([]byte(params.AzureVaultCredentials), &azureVaultCredentials)
			secretData, err := azure.GetSecretValue(azureVaultCredentials.Token, azureVaultCredentials.Name, secretConfig.Name)
			if err != nil {
				log.Fatalf("Error getting Build secret: %v", err)
				return map[string]string{}, err
			}
			return secretData, nil
		} else {
			return map[string]string{}, errors.New("No credentials provided")
		}
	}
	return map[string]string{}, nil
}
