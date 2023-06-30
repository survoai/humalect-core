package services

import (
	"encoding/json"
	"fmt"

	"github.com/Humalect/humalect-core/agent/constants"
	"github.com/Humalect/humalect-core/agent/utils"
)

func GetEcrCredentials(config *constants.ParamsConfig) (constants.EcrCredentials, error) {
	var ecrCredentials constants.EcrCredentials
	err := json.Unmarshal([]byte(config.EcrCredentials), &ecrCredentials)
	if err != nil {
		fmt.Println(err)
		config.WebhookData = utils.UpdateStatusData(config.WebhookData, constants.CreatedApplicationCrd, false)
		SendWebhook(config.WebhookEndpoint, config.WebhookData, false, constants.CreatedApplicationCrd)
	}
	return ecrCredentials, err
}

func GetDockerHubCredentials(config *constants.ParamsConfig) (constants.DockerHubCredentials, error) {
	var dockerHubCredentials constants.DockerHubCredentials
	err := json.Unmarshal([]byte(config.DockerHubCredentials), &dockerHubCredentials)
	if err != nil {
		fmt.Println(err)
		config.WebhookData = utils.UpdateStatusData(config.WebhookData, constants.CreatedApplicationCrd, false)
		SendWebhook(config.WebhookEndpoint, config.WebhookData, false, constants.CreatedApplicationCrd)
	}
	return dockerHubCredentials, err
}

func GetAcrCredentials(config *constants.ParamsConfig) (constants.AcrCredentials, error) {
	var acrCredentials constants.AcrCredentials
	err := json.Unmarshal([]byte(config.AcrCredentials), &acrCredentials)
	if err != nil {
		fmt.Println(err)
		config.WebhookData = utils.UpdateStatusData(config.WebhookData, constants.CreatedApplicationCrd, false)
		SendWebhook(config.WebhookEndpoint, config.WebhookData, false, constants.CreatedApplicationCrd)
	}
	return acrCredentials, err
}

func GetAwsSecretCredentials(config *constants.ParamsConfig) (constants.AwsSecretCredentials, error) {
	var awsSecretCredentials constants.AwsSecretCredentials
	err := json.Unmarshal([]byte(config.AwsSecretCredentials), &awsSecretCredentials)
	if err != nil {
		fmt.Println(err)
		config.WebhookData = utils.UpdateStatusData(config.WebhookData, constants.CreatedApplicationCrd, false)
		SendWebhook(config.WebhookEndpoint, config.WebhookData, false, constants.CreatedApplicationCrd)
	}
	return awsSecretCredentials, err
}

func GetAzureVaultCredentials(config *constants.ParamsConfig) (constants.AzureVaultCredentials, error) {
	var azureVaultCredentials constants.AzureVaultCredentials
	err := json.Unmarshal([]byte(config.AzureVaultCredentials), &azureVaultCredentials)
	if err != nil {
		fmt.Println(err)
		config.WebhookData = utils.UpdateStatusData(config.WebhookData, constants.CreatedApplicationCrd, false)
		SendWebhook(config.WebhookEndpoint, config.WebhookData, false, constants.CreatedApplicationCrd)
	}
	return azureVaultCredentials, err
}

func GetDeploymentYamlManifest(config *constants.ParamsConfig) (constants.DeploymentYamlManifestType, error) {
	var deploymentYamlManifest constants.DeploymentYamlManifestType
	err := json.Unmarshal([]byte(config.DeploymentYamlManifest), &deploymentYamlManifest)
	if err != nil {
		fmt.Println(err)
		config.WebhookData = utils.UpdateStatusData(config.WebhookData, constants.CreatedApplicationCrd, false)
		SendWebhook(config.WebhookEndpoint, config.WebhookData, false, constants.CreatedApplicationCrd)
	}
	return deploymentYamlManifest, err
}

func GetServiceYamlManifest(config *constants.ParamsConfig) (constants.ServiceYamlManifestType, error) {
	var serviceYamlManifest constants.ServiceYamlManifestType
	err := json.Unmarshal([]byte(config.ServiceYamlManifest), &serviceYamlManifest)
	if err != nil {
		fmt.Println(err)
		config.WebhookData = utils.UpdateStatusData(config.WebhookData, constants.CreatedApplicationCrd, false)
		SendWebhook(config.WebhookEndpoint, config.WebhookData, false, constants.CreatedApplicationCrd)
	}
	return serviceYamlManifest, err
}

func GetIngressYamlManifest(config *constants.ParamsConfig) (constants.IngressYamlManifestType, error) {
	var ingressYamlManifest constants.IngressYamlManifestType
	err := json.Unmarshal([]byte(config.IngressYamlManifest), &ingressYamlManifest)
	if err != nil {
		fmt.Println(err)
		config.WebhookData = utils.UpdateStatusData(config.WebhookData, constants.CreatedApplicationCrd, false)
		SendWebhook(config.WebhookEndpoint, config.WebhookData, false, constants.CreatedApplicationCrd)
	}
	return ingressYamlManifest, err
}
