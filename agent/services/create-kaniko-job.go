package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/Humalect/humalect-core/agent/constants"
	"github.com/Humalect/humalect-core/agent/services/aws"
	"github.com/Humalect/humalect-core/agent/services/azure"
	"github.com/Humalect/humalect-core/agent/services/dockerhub"
	"github.com/Humalect/humalect-core/agent/utils"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

type CreateJobConfig struct {
	CloudProviderSecretName string
	DockerFileConfigName    string
	KanikoJobName           string
}

const (
	kanikoWorkspaceName             = "workspace"
	dockerfileConfigDirectoryName   = "docker-config"
	kanikoDockerLocation            = "/kaniko/.docker/"
	kanikoDockerConfigName          = "kaniko-docker-config"
	cloudProviderRegistrySecretName = "cloud-provider-registry-secret"
	gitRepoVolumeName               = "git-repo"
)

func CreateKanikoJob(params constants.ParamsConfig) (CreateJobConfig, error) {
	config := GetK8sConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
		SendWebhook(params.WebhookEndpoint, params.WebhookData, false, constants.CreatedKanikoJob)
		return CreateJobConfig{}, errors.New("Error Starting Build")
	}
	createJobConfig, err := createKanikoConfigResources(clientset, params)
	if err != nil {
		log.Fatalf("Error creating resources for Job: %v", err)
		SendWebhook(params.WebhookEndpoint, params.WebhookData, false, constants.CreatedKanikoJob)
		return CreateJobConfig{}, errors.New("Error Starting Build")
	}

	job, err := getKanikoJobObject(createJobConfig, params)
	if err != nil {
		log.Fatalf("Error generating Job Yaml: %v", err)
		SendWebhook(params.WebhookEndpoint, params.WebhookData, false, constants.CreatedKanikoJob)
		return CreateJobConfig{}, errors.New("Error Starting Build")
	}

	jobClient := clientset.BatchV1().Jobs("humalect")
	createdJob, err := jobClient.Create(context.Background(), &job, metav1.CreateOptions{})
	if err != nil {
		SendWebhook(params.WebhookEndpoint, params.WebhookData, false, constants.CreatedKanikoJob)
		return CreateJobConfig{}, errors.New("Error Starting Build")
	}
	SendWebhook(params.WebhookEndpoint, params.WebhookData, true, constants.CreatedKanikoJob)
	createJobConfig.KanikoJobName = createdJob.GetName()
	return createJobConfig, nil
}

func createArtifactsSecret(clientset *kubernetes.Clientset, params constants.ParamsConfig) (string, error) {
	if params.ArtifactsRegistryProvider == constants.RegistryIdAWS {
		return aws.CreateEcrSecret(params, clientset)
	} else if params.ArtifactsRegistryProvider == constants.RegistryIdDockerhub {
		return dockerhub.CreateSecret(params, clientset)
	} else if params.ArtifactsRegistryProvider == constants.RegistryIdAzure || (params.ArtifactsRegistryProvider == "" && params.CloudProvider == constants.CloudIdAzure) {
		return azure.CreateAcrSecret(params, clientset)
	}
	return "", nil
}

func getCodeSourceSpecificGitUrl(params constants.ParamsConfig) string {
	gitRepoUrl := ""
	switch params.SourceCodeProvider {
	case constants.SourceGithub:
		gitRepoUrl = fmt.Sprintf("https://x-access-token:%s@github.com/%s.git", params.SourceCodeToken, params.SourceCodeRepositoryName)
	case constants.SourceGitlab:
		gitRepoUrl = strings.Replace(params.SourceCodeRepositoryName, "https://", fmt.Sprintf("https://oauth2:%s@", params.SourceCodeToken), 1)
	case constants.SourceBitbucket:
		gitRepoUrl = fmt.Sprintf("https://x-token-auth:%s@bitbucket.org/%s.git", params.SourceCodeToken, params.SourceCodeRepositoryName)
	default:
		log.Fatal("Invalid Source Code Provider received.")
	}
	return gitRepoUrl
}

func getDockerFileConfig(clientset *kubernetes.Clientset, params constants.ParamsConfig) (string, error) {
	if !params.UseDockerFromCodeFlag {
		var dockerCommands []string
		err := json.Unmarshal([]byte(params.DockerManifest), &dockerCommands)
		if err != nil {
			log.Fatalf("Failed to parse Dockerfile got error : %v", err)
			return "", err
		}

		dockerFileContent := strings.Join(dockerCommands, "\r\n")

		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-dockerfile-%s-%s",
					params.ManagedBy[:int(math.Min(float64(len(params.ManagedBy)), float64(10)))],
					params.CommitId[:int(math.Min(float64(len(params.CommitId)), float64(5)))],
					params.DeploymentId[:int(math.Min(float64(len(params.DeploymentId)), float64(7)))]),
				Namespace: "humalect",
				Labels: map[string]string{
					"app": fmt.Sprintf("%s-dockerfile-%s-%s",
						params.ManagedBy[:int(math.Min(float64(len(params.ManagedBy)), float64(10)))],
						params.CommitId[:int(math.Min(float64(len(params.CommitId)), float64(5)))],
						params.DeploymentId[:int(math.Min(float64(len(params.DeploymentId)), float64(7)))]),
					"deploymentId": params.DeploymentId,
					"managedBy":    params.ManagedBy,
					"commitId":     params.CommitId,
					"partOf":       "humalect-core",
					"resourceType": "dockerfile-config",
					"pipelineId":   params.PipelineId,
				},
			},
			Data: map[string]string{
				"Dockerfile": dockerFileContent,
			},
		}
		configMap, err = clientset.CoreV1().ConfigMaps("humalect").Create(context.Background(), configMap, metav1.CreateOptions{})
		if err != nil {
			fmt.Println("Error creating ConfigMap:", err)
			return "", err
		}
		return configMap.Name, nil

	}
	return "", nil
}

func getKanikoJobObject(
	createJobConfig CreateJobConfig,
	params constants.ParamsConfig,
) (batchv1.Job, error) {
	var artifactsRepoUrl string
	var imageTag=utils.MergeParseString(params.CommitId, params.PipelineId, 30)
	if params.ArtifactsRegistryProvider == constants.RegistryIdAzure || (params.ArtifactsRegistryProvider == "" && params.CloudProvider == constants.CloudIdAzure) {
		var acrCredentials constants.AcrCredentials
		_ = json.Unmarshal([]byte(params.AcrCredentials), &acrCredentials)
		artifactsRepoUrl = fmt.Sprintf("%s.azurecr.io/%s:%s", acrCredentials.RegistryName, params.
			ArtifactsRepositoryName, imageTag)
	} else if params.ArtifactsRegistryProvider == constants.RegistryIdAWS || (params.ArtifactsRegistryProvider == "" && params.CloudProvider == constants.CloudIdAWS) {
		var ecrCredentials constants.EcrCredentials
		_ = json.Unmarshal([]byte(params.EcrCredentials), &ecrCredentials)
		artifactsRepoUrl = fmt.Sprintf("%s/%s:%s", ecrCredentials.RegistryUrl, params.
			ArtifactsRepositoryName, imageTag)
	} else if params.ArtifactsRegistryProvider == constants.RegistryIdDockerhub {
		var dockerHubCreds constants.DockerHubCredentials
		_ = json.Unmarshal([]byte(params.DockerHubCredentials), &dockerHubCreds)
		artifactsRepoUrl = fmt.Sprintf("%s/%s:%s", dockerHubCreds.Username, params.ArtifactsRepositoryName, imageTag)

	} else {
		fmt.Println("Invalid Artifacts Registry Provider received.")
		return batchv1.Job{}, errors.New("Invalid Artifacts Registry Provider received.")
	}
	gitUrl := getCodeSourceSpecificGitUrl(params)
	prepareConfigVolumeMounts := []corev1.VolumeMount{
		{
			Name:      gitRepoVolumeName,
			MountPath: fmt.Sprintf("/%s", kanikoWorkspaceName),
		},
	}
	kanikoVolumes := []corev1.Volume{
		{
			Name:         gitRepoVolumeName,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		},
	}
	prepareConfigCommand := `git clone --no-checkout %[1]s /%[2]s && cd /%[2]s &&   git fetch --all &&  git checkout %[3]s `
	kanikoEnvVars := []corev1.EnvVar{}
	kanikoVolumeMounts := []corev1.VolumeMount{
		{
			Name:      gitRepoVolumeName,
			MountPath: fmt.Sprintf("/%s/", kanikoWorkspaceName),
		},
	}
	if len(createJobConfig.DockerFileConfigName) != 0 {
		prepareConfigCommand = prepareConfigCommand + fmt.Sprintf(` && cp /%s/Dockerfile /%s/Dockerfile --force`, dockerfileConfigDirectoryName, kanikoWorkspaceName)
		prepareConfigVolumeMounts = append(prepareConfigVolumeMounts, corev1.VolumeMount{
			Name:      dockerfileConfigDirectoryName,
			MountPath: fmt.Sprintf("/%s", dockerfileConfigDirectoryName),
		})
		kanikoVolumes = append(kanikoVolumes, corev1.Volume{
			Name: dockerfileConfigDirectoryName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: createJobConfig.DockerFileConfigName,
					},
				},
			},
		})

	}
	if len(createJobConfig.CloudProviderSecretName) != 0 {

		prepareConfigCommand = prepareConfigCommand + fmt.Sprintf(` && cp /%s/.dockerconfigjson /%s/config.json `, cloudProviderRegistrySecretName, kanikoDockerConfigName)
		prepareConfigVolumeMounts = append(prepareConfigVolumeMounts, corev1.VolumeMount{
			Name:      cloudProviderRegistrySecretName,
			MountPath: fmt.Sprintf("/%s", cloudProviderRegistrySecretName),
		})
		kanikoVolumes = append(kanikoVolumes, corev1.Volume{
			Name: cloudProviderRegistrySecretName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: createJobConfig.CloudProviderSecretName,
				},
			},
		})
		prepareConfigVolumeMounts = append(prepareConfigVolumeMounts, corev1.VolumeMount{
			Name:      kanikoDockerConfigName,
			MountPath: fmt.Sprintf("/%s", kanikoDockerConfigName),
		})
		kanikoEnvVars = append(kanikoEnvVars, corev1.EnvVar{
			Name:  "DOCKER_CONFIG",
			Value: kanikoDockerLocation,
		})
		kanikoVolumeMounts = append(kanikoVolumeMounts, corev1.VolumeMount{
			Name:      kanikoDockerConfigName,
			MountPath: kanikoDockerLocation,
		})
		kanikoVolumes = append(kanikoVolumes, corev1.Volume{
			Name:         kanikoDockerConfigName,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		})
	}
	buildArgs, err := getKanikoBuildArgs(params)
	if err != nil {
		return batchv1.Job{}, err
	}
	podSpec := corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name:  "prepare-config",
				Image: "alpine/git",
				Command: []string{
					"/bin/sh",
					"-c",
					fmt.Sprintf(prepareConfigCommand, gitUrl, kanikoWorkspaceName, params.CommitId),
				},
				VolumeMounts: prepareConfigVolumeMounts,
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "kaniko",
				Image:           "gcr.io/kaniko-project/executor:latest",
				ImagePullPolicy: corev1.PullAlways,
				Args: append(
					[]string{
						fmt.Sprintf("--context=dir:///%s", kanikoWorkspaceName),
						fmt.Sprintf("--dockerfile=/%s/Dockerfile", kanikoWorkspaceName),
						fmt.Sprintf("--destination=%s", artifactsRepoUrl),
					}, buildArgs...),
				Env:          kanikoEnvVars,
				VolumeMounts: kanikoVolumeMounts,
			},
		},
		RestartPolicy:      corev1.RestartPolicyNever,
		Volumes:            kanikoVolumes,
		ServiceAccountName: "humalect-sa",
	}
	backoffLimit := int32(0)
	jobSpec := batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"deploymentId": params.DeploymentId,
					"managedBy":    params.ManagedBy,
					"commitId":     params.CommitId,
					"pipelineId":   params.PipelineId,
					"partOf":       "humalect-core",
					"resourceType": "kaniko-job-pod",
				},
			},
			Spec: podSpec,
		},
		BackoffLimit: &backoffLimit,
	}
	job := batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-build-push-dockerimage-%s-%s",
				params.ManagedBy[:int(math.Min(float64(len(params.ManagedBy)), float64(10)))],
				params.CommitId[:int(math.Min(float64(len(params.CommitId)), float64(5)))],
				params.DeploymentId[:int(math.Min(float64(len(params.DeploymentId)), float64(7)))]),
			Labels: map[string]string{
				"app": fmt.Sprintf("%s-build-push-dockerimage-%s-%s",
					params.ManagedBy[:int(math.Min(float64(len(params.ManagedBy)), float64(10)))],
					params.CommitId[:int(math.Min(float64(len(params.CommitId)), float64(5)))],
					params.DeploymentId[:int(math.Min(float64(len(params.DeploymentId)), float64(7)))]),
				"DeploymentId": params.DeploymentId,
				"ManagedBy":    params.ManagedBy,
				"CommitId":     params.CommitId,
				"PartOf":       "humalect-core",
				"PipelineId":   params.PipelineId,
				"Type":         "kaniko-job",
			},
			Namespace: "humalect",
		},
		Spec: jobSpec,
	}
	return job, nil
}

func createKanikoConfigResources(clientset *kubernetes.Clientset, params constants.ParamsConfig) (CreateJobConfig, error) {
	cloudProviderSecretName, err := createArtifactsSecret(clientset, params)
	if err != nil {
		log.Fatalf("Error Creating Artifacts Secret: %v", err)
		// return CreateJobConfig{}, err
	}

	dockerFileConfigName, err := getDockerFileConfig(clientset, params)
	if err != nil {
		log.Fatalf("Error creating Dockerfile Config: %v", err)
		return CreateJobConfig{}, err
	}
	return CreateJobConfig{CloudProviderSecretName: cloudProviderSecretName, DockerFileConfigName: dockerFileConfigName}, nil
}

func getKanikoBuildArgs(params constants.ParamsConfig) ([]string, error) {
	secretData, err := FetchBuildSecrets(params)
	if err != nil {
		return []string{}, err
	}
	secretArgs := []string{}
	for key, value := range secretData {
		if value != "" {
			secretArgs = append(secretArgs, fmt.Sprintf("--build-arg=%s=%s", key, value))
		}
	}
	return secretArgs, err
}
