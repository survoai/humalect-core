package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/Humalect/humalect-core/agent/constants"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

type AzureCreds struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
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
		panic(err)
	}
	createJobConfig, err := createKanikoConfigResources(clientset, params)
	if err != nil {
		log.Fatalf("Error creating resources for Job: %v", err)
		SendWebhook(params.WebhookEndpoint, params.WebhookData, false, constants.CreatedKanikoJob)
		panic(err)
	}

	job, err := getKanikoJobObject(createJobConfig, params)
	if err != nil {
		log.Fatalf("Error generating Job Yaml: %v", err)
		SendWebhook(params.WebhookEndpoint, params.WebhookData, false, constants.CreatedKanikoJob)
		panic(err)
	}

	jobClient := clientset.BatchV1().Jobs("humalect")
	createdJob, err := jobClient.Create(context.Background(), &job, metav1.CreateOptions{})
	if err != nil {
		SendWebhook(params.WebhookEndpoint, params.WebhookData, false, constants.CreatedKanikoJob)
		panic(err)
	}
	SendWebhook(params.WebhookEndpoint, params.WebhookData, true, constants.CreatedKanikoJob)
	createJobConfig.KanikoJobName = createdJob.GetName()
	return createJobConfig, nil
}

func createCloudProviderCredSecrets(clientset *kubernetes.Clientset, params constants.ParamsConfig) (string, error) {
	if params.CloudProvider == constants.CloudIdAzure {
		azureCreds, err := getOrgAzureCredsForAcr(params.AzureManagementScopeToken, params.AzureAcrRegistryName,
			params.AzureSubscriptionId,
			params.AzureResourceGroupName)
		if err != nil {
			log.Fatalf("Error getting Azure ACR creds: %v", err)
			return "", err
		}

		dockerRegistrySecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-ksec-%s-%s",
					params.ManagedBy[:int(math.Min(float64(len(params.ManagedBy)), float64(10)))],
					params.CommitId[:int(math.Min(float64(len(params.CommitId)), float64(5)))],
					params.DeploymentId[:int(math.Min(float64(len(params.DeploymentId)), float64(7)))]),
				Namespace: "humalect",
				Labels: map[string]string{
					"app": fmt.Sprintf("%s-build-push-dockerimage-%s-%s",
						params.ManagedBy[:int(math.Min(float64(len(params.ManagedBy)), float64(10)))],
						params.CommitId[:int(math.Min(float64(len(params.CommitId)), float64(5)))],
						params.DeploymentId[:int(math.Min(float64(len(params.DeploymentId)), float64(7)))]),
					"DeploymentId": params.DeploymentId,
					"ManagedBy":    params.ManagedBy,
					"CommitId":     params.CommitId,
				},
			},
			Type: corev1.SecretTypeDockerConfigJson,
			StringData: map[string]string{
				".dockerconfigjson": fmt.Sprintf(`{  
					"auths": {  
						"%s.azurecr.io": {  
							"username": "%s",  
							"password": "%s"  
						}  
					}  
				}`, params.AzureAcrRegistryName, azureCreds.Username, azureCreds.Password),
			},
		}

		// Create the Docker registry secret in the default Namespace
		createdSecret, err := clientset.CoreV1().Secrets("humalect").Create(context.Background(), dockerRegistrySecret, metav1.CreateOptions{})
		if err != nil {
			log.Fatalf("Error creating docker registry secret secret: %v", err)
		}

		log.Printf("Docker Registry Secret %s created in Namespace %s\n", createdSecret.Name, createdSecret.Namespace)

		return createdSecret.Name, nil

	} else {
		return "", nil
	}
}

func getOrgAzureCredsForAcr(AzureManagementScopeToken string, AzureAcrRegistryName string, AzureSubscriptionId string, AzureResourceGroupName string) (AzureCreds, error) {
	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ContainerRegistry/registries/%s/listCredentials?api-version=2019-05-01", AzureSubscriptionId, AzureResourceGroupName, AzureAcrRegistryName)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader("{}"))
	if err != nil {
		return AzureCreds{}, err
	}

	req.Header.Set("Authorization", "Bearer "+AzureManagementScopeToken)
	resp, err := client.Do(req)
	if err != nil {
		return AzureCreds{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Status Code: ", resp.StatusCode)
		fmt.Println("resp", resp)
		return AzureCreds{}, errors.New("non-200 status code received when tried to get Creds for Azure ACR")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return AzureCreds{}, err
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	username := result["username"].(string)
	password := result["passwords"].([]interface{})[0].(map[string]interface{})["value"].(string)

	if username == "" {
		return AzureCreds{}, errors.New("unable to fetch credentials for ACR")
	}

	return AzureCreds{
		Username: username,
		Password: password,
	}, nil
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
					"DeploymentId": params.DeploymentId,
					"ManagedBy":    params.ManagedBy,
					"CommitId":     params.CommitId,
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
	if params.CloudProvider == constants.CloudIdAzure {
		artifactsRepoUrl = fmt.Sprintf("%s.azurecr.io/%s:%s", params.AzureAcrRegistryName, params.
			ArtifactsRepositoryName, params.CommitId)
	} else if params.CloudProvider == constants.CloudIdAWS {
		artifactsRepoUrl = fmt.Sprintf("%s/%s:%s", params.AwsEcrRegistryUrl, params.
			ArtifactsRepositoryName, params.CommitId)
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
				Args: []string{
					fmt.Sprintf("--context=dir:///%s", kanikoWorkspaceName),
					fmt.Sprintf("--dockerfile=/%s/Dockerfile", kanikoWorkspaceName),
					fmt.Sprintf("--destination=%s", artifactsRepoUrl),
				},
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
					"DeploymentId": params.DeploymentId,
					"ManagedBy":    params.ManagedBy,
					"CommitId":     params.CommitId,
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
			},
			Namespace: "humalect",
		},
		Spec: jobSpec,
	}
	return job, nil
}

func createKanikoConfigResources(clientset *kubernetes.Clientset, params constants.ParamsConfig) (CreateJobConfig, error) {
	cloudProviderSecretName, err := createCloudProviderCredSecrets(clientset, params)
	if err != nil {
		log.Fatalf("Error Getting Secret Name: %v", err)
		return CreateJobConfig{}, err
	}

	dockerFileConfigName, err := getDockerFileConfig(clientset, params)
	if err != nil {
		log.Fatalf("Error creating Dockerfile Config: %v", err)
		return CreateJobConfig{}, err
	}
	return CreateJobConfig{CloudProviderSecretName: cloudProviderSecretName, DockerFileConfigName: dockerFileConfigName}, nil
}
