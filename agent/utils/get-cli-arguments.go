package utils

import (
	"flag"

	"github.com/Humalect/humalect-core/agent/constants"
)

func ParseCLIArguments() *constants.ParamsConfig {
	config := &constants.ParamsConfig{}
	flag.StringVar(&config.SecretsProvider, "secretsProvider", "", "This is a required parameter which can have 2 values either aws or azure and it represents the main secrets provider to be used.")
	flag.StringVar(&config.EcrCredentials, "ecrCredentials", "", "This is an optional parameter and it would only be passed when the registryProvider is aws and it represents the credentials for the AWS ECR registry.")
	flag.StringVar(&config.DockerHubCredentials, "dockerHubCredentials", "", "This is an optional parameter and it would only be passed when the registryProvider is dockerhub and it represents the credentials for the dockerhub registry.")
	flag.StringVar(&config.AcrCredentials, "acrCredentials", "", "This is an optional parameter and it would only be passed when the registryProvider is azure and it represents the credentials for the Azure ACR registry.")
	flag.StringVar(&config.AwsSecretCredentials, "awsSecretCredentials", "", "This is an optional parameter and it would only be passed when the registryProvider is aws and it represents the credentials for the AWS ECR registry.")
	flag.StringVar(&config.AzureSecretCredentials, "azureSecretCredentials", "", "This is an optional parameter and it would only be passed when the registryProvider is azure and it represents the credentials for the Azure ACR registry.")
	flag.StringVar(&config.CloudProvider, "cloudProvider", "", "This is a required parameter which can have 2 values either aws or azure and it represents the main cloud provider to be used.")
	flag.StringVar(&config.AzureManagementScopeToken, "azureManagementScopeToken", "", "This is the optional parameter which shows the token of Azure for https://management.azure.com/.default scope and this would only be passed when the cloudProvider is azure.")
	flag.StringVar(&config.SourceCodeRepositoryName, "sourceCodeRepositoryName", "", "The repository name of your github, gitlab or bitbucket repository")
	flag.StringVar(&config.SourceCodeProvider, "sourceCodeProvider", "", "This is a required parameter that represents the name of the version control system portal and it could be either github  or gitlab or bitbucket")
	flag.StringVar(&config.SourceCodeToken, "sourceCodeToken", "", "This is a required parameter that represents the token to pull the source code from the version control system portal namely github, gitlab and bitbucket.")
	flag.StringVar(&config.RegistryProvider, "registryProvider", "", "This is a required parameter that represents the name of the docker registry provider and it could be either dockerhub or azure or aws.")
	flag.StringVar(&config.SourceCodeOrgName, "sourceCodeOrgName", "", "This is a required parameter that represents the organization or user name for the version control system portal namely github, gitlab and bitbucket.")
	flag.StringVar(&config.CommitId, "commitId", "", "This is a required parameter that represents the commitId for the version control system portal namely github, gitlab and bitbucket.")
	flag.StringVar(&config.DockerManifest, "dockerManifest", "", "This is a required parameter and this represents the docker file for the source code that is to be used to build docker image. It is an array of strings with each string representing a line in the dockerfile.")
	flag.StringVar(&config.ArtifactsRepositoryName, "artifactsRepositoryName", "", "This is a required parameter that represents the name of the Artifacts repository that is to be used to push docker image.")
	flag.StringVar(&config.AwsEcrRegistryUrl, "awsEcrRegistryUrl", "", "This is an optional parameter and it would only be passed when the cloudProvider is aws and it represents the URl for the AWS ECR registry.")
	flag.StringVar(&config.AzureAcrRegistryName, "azureAcrRegistryName", "", "This is an optional parameter and it would only be passed when the cloudProvider is azure and it represents the name for the Azure ACR registry.")
	flag.StringVar(&config.AwsEcrUserName, "awsEcrUserName", "", "This is an optional parameter and it would only be passed when the cloudProvider is aws and it represents the username that is to be used to push docker image to ECR.")
	flag.StringVar(&config.AzureSubscriptionId, "azureSubscriptionId", "", "This is an optional parameter and it would only be passed when the cloudProvider is azure and it represents the subscription ID for azure.")
	flag.StringVar(&config.AzureResourceGroupName, "azureResourceGroupName", "", "This is an optional parameter and it would only be passed when the cloudProvider is azure and it represents the resource group name for azure ACR registry and secrets vault")
	flag.BoolVar(&config.UseDockerFromCodeFlag, "useDockerFromCodeFlag", false, "This is a required boolean parameter that is used to decide weather source code docker file is to be used or not.")
	flag.StringVar(&config.DeploymentYamlManifest, "deploymentYamlManifest", "", "This is a required parameter and it represents the Deployment Yaml Manifest for the project in the stringified JSON format.")
	flag.StringVar(&config.IngressYamlManifest, "ingressYamlManifest", "", "This is a required parameter and it represents the Ingress Yaml Manifest for the project in the stringified JSON format.")
	flag.StringVar(&config.ServiceYamlManifest, "serviceYamlManifest", "", "This is a required parameter and it represents the Service Yaml Manifest for the project in the stringified JSON format.")
	flag.StringVar(&config.K8sAppName, "k8sAppName", "", "This is a required parameter and it represents the application name which is to be deployed(it can be any string of your choice).")
	flag.StringVar(&config.ManagedBy, "managedBy", "", "The is an optional parameter and it represents the name of the entity that is responsible to manage the resources. It is set to humalect by default.")
	flag.StringVar(&config.CloudRegion, "cloudRegion", "", "This is a required parameter and it represents the region of the cloud provider in which the K8s cluster is deployed.")
	flag.StringVar(&config.K8sResourcesIdentifier, "k8sResourcesIdentifier", "", "The is a required parameter and it represents the unique identifier that is to be used to uniquely identify the K8s resources.")
	flag.StringVar(&config.AzureVaultToken, "azureVaultToken", "", "This is the optional parameter which shows the token of Azure for https://vault.azure.net/.default scope and this would only be passed when the cloudProvider is azure.")
	flag.StringVar(&config.SecretManagerName, "secretManagerName", "", "This is a required parameter and it represents the secret name for cloud provider secret.")
	flag.StringVar(&config.AzureVaultName, "azureVaultName", "", "This is an optional parameter and it represents the Azure vault name for the provided secret.")
	flag.StringVar(&config.Namespace, "namespace", "", "This is an optional parameter and represents the namespace in which you want to deploy the k8s resources.(it is set to default if not passed)")
	flag.StringVar(&config.DeploymentId, "deploymentId", "", "This is a required parameter and represents the unique deployment id for each deployment.")
	flag.StringVar(&config.WebhookEndpoint, "webhookEndpoint", "", "This is an optional parameter and represents the endpoint which can be used to send webhook notifications.")
	flag.StringVar(&config.WebhookData, "webhookData", "", "This is an optional parameter and represents the data that is to be sent to the webhook endpoint(in json string format).")

	flag.Parse()
	return config
}
