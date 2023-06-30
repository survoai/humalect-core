package tasks

import (
	"fmt"

	"github.com/Humalect/humalect-core/agent/constants"
	"github.com/Humalect/humalect-core/agent/services"
	"github.com/Humalect/humalect-core/agent/utils"
)

func Deploy(config *constants.ParamsConfig) error {
	// repoArchiveURL, err := services.CloneSourceCode(config.SourceCodeProvider, config.SourceCodeOrgName, config.SourceCodeRepositoryName, config.CommitId, config.SourceCodeToken)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// fmt.Println(repoArchiveURL)
	// err = services.AddCustomDockerfile(config.UseDockerFromCodeFlag, config.DockerManifest, config.SourceCodeRepositoryName, config.CommitId, config.SourceCodeToken)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// artifactsRepoLink, err := utils.GetArtifactsRepoLink(config.CloudProvider, config.AwsEcrRegistryUrl, config.ArtifactsRepositoryName, config.CommitId, config.AzureAcrRegistryName)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// fmt.Println(artifactsRepoLink)
	// err = services.BuildDockerImage(config.ArtifactsRepositoryName, artifactsRepoLink, config.CommitId)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// err = services.PushDockerImage(config.CloudProvider, config.AwsEcrUserName, config.AwsEcrRegistryUrl, config.AzureSubscriptionId, config.AzureResourceGroupName, config.AzureManagementScopeToken, config.AzureAcrRegistryName, artifactsRepoLink)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }

	kanikoJobResources, err := services.CreateKanikoJob(*config)
	if err != nil {
		config.WebhookData = utils.UpdateStatusData(config.WebhookData, constants.CreatedKanikoJob, false)
		services.SendWebhook(config.WebhookEndpoint, config.WebhookData, false, constants.CreatedKanikoJob)
		fmt.Println(err)
		return err
	}
	fmt.Println("Kaniko Job Created")
	status := services.WatchJobEvents("humalect", kanikoJobResources.KanikoJobName)
	if !status {
		fmt.Println("Kaniko Job Failed")
		config.WebhookData = utils.UpdateStatusData(config.WebhookData, constants.KanikoJobExecuted, false)
		services.SendWebhook(config.WebhookEndpoint, config.WebhookData, false, constants.KanikoJobExecuted)
		return nil
	}
	fmt.Println("Kaniko Job Completed")
	config.WebhookData = utils.UpdateStatusData(config.WebhookData, constants.KanikoJobExecuted, true)
	services.SendWebhook(config.WebhookEndpoint, config.WebhookData, true, constants.KanikoJobExecuted)

	// awsSecretCredentials, err := services.GetAwsSecretCredentials(config)
	// if err != nil {
	// 	return err
	// }

	// azureVaultCredentials, err := services.GetAzureVaultCredentials(config)
	// if err != nil {
	// 	return err
	// }

	// deploymentYamlManifest, err := services.GetDeploymentYamlManifest(config)
	// if err != nil {
	// 	return err
	// }
	// deploymentYamlManifest.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: kanikoJobResources.CloudProviderSecretName}}

	// serviceYamlManifest, err := services.GetServiceYamlManifest(config)
	// if err != nil {
	// 	return err
	// }

	// ingressYamlManifest, err := services.GetIngressYamlManifest(config)
	// if err != nil {
	// 	return err
	// }

	_, err = services.CreateK8sApplication(config, kanikoJobResources, utils.UpdateStatusData(config.WebhookData, constants.CreatedApplicationCrd, true))
	if err != nil {
		fmt.Println(err)
		config.WebhookData = utils.UpdateStatusData(config.WebhookData, constants.CreatedApplicationCrd, false)
		services.SendWebhook(config.WebhookEndpoint, config.WebhookData, false, constants.CreatedApplicationCrd)
		return err
	}
	fmt.Println("Application created")
	err = services.CleanupKanikoJobResources(kanikoJobResources)
	// TODO send webhook here
	if err != nil {
		fmt.Println(err)
		// services.SendWebhook(config.WebhookEndpoint, config.WebhookData, false, constants.CreatedApplicationCrd)
		return err
	}
	config.WebhookData = utils.UpdateStatusData(config.WebhookData, constants.CreatedApplicationCrd, true)
	services.SendWebhook(config.WebhookEndpoint, config.WebhookData, true, constants.CreatedApplicationCrd)
	return nil
}
