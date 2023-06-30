package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/Humalect/humalect-core/agent/constants"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

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
	if params.ArtifactsRegistryProvider == constants.RegistryIdAWS {
		var ecrCredentials constants.EcrCredentials
		err := json.Unmarshal([]byte(params.EcrCredentials), &ecrCredentials)

		if err != nil {
			log.Fatalf("Failed to parse ECR credentials got error : %v", err)
			return "", err
		}

		ecrToken, err := getEcrLoginToken(ecrCredentials.AccessKey, ecrCredentials.SecretKey, ecrCredentials.Region)

		if err != nil {
			log.Fatalf("Error getting ECR token: %v", err)
			return "", err
		}

		dockerRegistrySecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-ksec-ecr-%s-%s",
					params.ManagedBy[:int(math.Min(float64(len(params.ManagedBy)), float64(10)))],
					params.CommitId[:int(math.Min(float64(len(params.CommitId)), float64(5)))],
					params.DeploymentId[:int(math.Min(float64(len(params.DeploymentId)), float64(7)))]),
				Namespace: "humalect",
				Labels: map[string]string{
					"app": fmt.Sprintf("%s-build-push-dockerecrimage-%s-%s",
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
						"%s": {  
							"username":"AWS",
							"password": "%s"  
						}  
					}  
				}`, ecrCredentials.RegistryUrl, ecrToken),
			},
		}

		createdSecret, err := clientset.CoreV1().Secrets("humalect").Create(context.Background(), dockerRegistrySecret, metav1.CreateOptions{})
		if err != nil {
			log.Fatalf("Error creating docker registry secret secret: %v", err)
		}
		log.Printf("Docker Registry Secret %s created in Namespace %s\n", createdSecret.Name, createdSecret.Namespace)
		return createdSecret.Name, nil

	} else if params.ArtifactsRegistryProvider == constants.RegistryIdDockerhub {

		secretData, err := getDockerHubSecretKey(params)

		if err != nil {
			log.Fatalf("Error getting dockerhub secret: %v", err)
			return "", err
		}

		dockerRegistrySecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-ksec-dh-%s-%s",
					params.ManagedBy[:int(math.Min(float64(len(params.ManagedBy)), float64(10)))],
					params.CommitId[:int(math.Min(float64(len(params.CommitId)), float64(5)))],
					params.DeploymentId[:int(math.Min(float64(len(params.DeploymentId)), float64(7)))]),
				Namespace: "humalect",
				Labels: map[string]string{
					"app": fmt.Sprintf("%s-build-push-dockerHubimage-%s-%s",
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
				".dockerconfigjson": strings.Join(strings.Fields(fmt.Sprintf(`{  
					"auths": {  
						"https://index.docker.io/v1/": {  
							"auth": "%s"  
						}  
					}  
				}`, secretData)), ""),
			},
		}

		createdSecret, err := clientset.CoreV1().Secrets("humalect").Create(context.Background(), dockerRegistrySecret, metav1.CreateOptions{})
		if err != nil {
			log.Fatalf("Error creating docker registry secret secret: %v", err)
		}

		log.Printf("Docker Registry Secret %s created in Namespace %s\n", createdSecret.Name, createdSecret.Namespace)

		return createdSecret.Name, nil

	} else if params.ArtifactsRegistryProvider == constants.RegistryIdAzure {

		var acrCredentials constants.AcrCredentials
		err := json.Unmarshal([]byte(params.AcrCredentials), &acrCredentials)

		if err != nil {
			log.Fatalf("Failed to parse ECR credentials got error : %v", err)
			return "", err
		}

		azureCreds, err := getOrgAzureCredsForAcr(acrCredentials.ManagementScopeToken, acrCredentials.RegistryName,
			acrCredentials.SubscriptionId, acrCredentials.ResourceGroupName)
		if err != nil {
			log.Fatalf("Error getting Azure ACR creds: %v", err)
			return "", err
		}

		// azureCreds, err := getOrgAzureCredsForAcr(params.AzureManagementScopeToken, params.AzureAcrRegistryName, params.AzureSubscriptionId, params.AzureResourceGroupName)
		dockerRegistrySecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-ksec-acr-%s-%s",
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
				}`, acrCredentials.RegistryName, azureCreds.Username, azureCreds.Password),
			},
		}
		createdSecret, err := clientset.CoreV1().Secrets("humalect").Create(context.Background(), dockerRegistrySecret, metav1.CreateOptions{})
		if err != nil {
			log.Fatalf("Error creating docker registry secret secret: %v", err)
		}

		log.Printf("Docker Registry Secret %s created in Namespace %s\n", createdSecret.Name, createdSecret.Namespace)
		return createdSecret.Name, nil
	}
	return "", nil
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
	if params.ArtifactsRegistryProvider == constants.RegistryIdAzure {
		artifactsRepoUrl = fmt.Sprintf("%s.azurecr.io/%s:%s", params.AzureAcrRegistryName, params.
			ArtifactsRepositoryName, params.CommitId)
	} else if params.ArtifactsRegistryProvider == constants.RegistryIdAWS {
		artifactsRepoUrl = fmt.Sprintf("%s/%s:%s", params.AwsEcrRegistryUrl, params.
			ArtifactsRepositoryName, params.CommitId)
	} else if params.ArtifactsRegistryProvider == constants.RegistryIdDockerhub {
		var dockerHubCreds constants.DockerHubCredentials
		err := json.Unmarshal([]byte(params.DockerHubCredentials), &dockerHubCreds)
		if err != nil {
			log.Fatalf("Failed to parse ECR credentials got error : %v", err)
		}

		artifactsRepoUrl = fmt.Sprintf("%s/%s:%s", dockerHubCreds.Username, params.ArtifactsRepositoryName, params.CommitId)

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

func getAwsSecretValue(secretName, accessKey, secretKey, region string) (string, error) {
	// Create a session object with the access key and secret key
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return "", err
	}

	// Create a Secrets Manager client
	svc := secretsmanager.New(sess)

	// Call the GetSecretValue API
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}
	result, err := svc.GetSecretValue(input)
	if err != nil {
		fmt.Println("Error getting secret value:", err)
		return "", err
	}

	// Extract the secret value and return it
	secretValue := *result.SecretString

	var secretData map[string]string
	err = json.Unmarshal([]byte(secretValue), &secretData)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return "", err
	}

	// Return the "dockerhub" key's value
	return secretData[constants.RegistryIdDockerhub], nil
}

func getAzureSecretString(azureVaultToken string, vaultName string, secretName string) (string, error) {
	// cred, err := azidentity.NewDefaultAzureCredential(nil)
	// if err != nil {
	// 	fmt.Println("Error creating Azure Credential:", err)
	// 	return "", err
	// }
	// fmt.Println("Here goes credentials")
	// fmt.Println(cred)

	// client, err := azsecrets.NewClient(vaultURL, cred, nil)
	// if err != nil {
	// 	fmt.Println("Error creating Azure Secret Client:", err)
	// 	return "", err
	// }

	// resp, err := client.GetSecret(context.Background(), secretName, "", nil)
	// if err != nil {
	// 	fmt.Println("Error retrieving secret value:", err)
	// 	return "", err
	// }

	// var secretValue string = *resp.Value
	// fmt.Println("Secret value goes here", secretValue)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	url := fmt.Sprintf("https://%s.vault.azure.net/secrets/%s?api-version=7.3", vaultName, secretName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", azureVaultToken))

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var responseJSON map[string]interface{}
	err = json.Unmarshal(body, &responseJSON)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response JSON: %v", err)
	}

	secretValue, ok := responseJSON["value"].(string)
	if !ok {
		return "", fmt.Errorf("value not found in response JSON")
	}

	return secretValue, nil
}

func getEcrLoginToken(accessKey string, secretKey string, region string) (string, error) {

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return "", err
	}

	ecrClient := ecr.New(sess)

	result, err := ecrClient.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		fmt.Println("Error getting ECR authorization token:", err)
		return "", err
	}

	decodedToken, err := base64.StdEncoding.DecodeString(*result.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		fmt.Println("Error decoding ECR authorization token:", err)
		return "", err
	}

	password := strings.Split(string(decodedToken), ":")[1]
	return password, nil
}

func getDockerHubSecretKey(params constants.ParamsConfig) (string, error) {

	var dockerHubCreds constants.DockerHubCredentials
	err := json.Unmarshal([]byte(params.DockerHubCredentials), &dockerHubCreds)

	if err != nil {
		log.Fatalf("Failed to parse DockerHub credentials got error : %v", err)
		return "", err
	}

	if params.SecretsProvider == constants.CloudIdAWS {

		var awsSecretCredentials constants.AwsSecretCredentials
		err := json.Unmarshal([]byte(params.AwsSecretCredentials), &awsSecretCredentials)

		if err != nil {
			log.Fatalf("Failed to parse Aws secrets credentials got error : %v", err)
			return "", err
		}
		return getAwsSecretValue(dockerHubCreds.SecretName, awsSecretCredentials.AccessKey, awsSecretCredentials.SecretKey, awsSecretCredentials.Region)
	} else if params.SecretsProvider == constants.CloudIdAzure {

		var azureVaultCredentials constants.AzureVaultCredentials
		err := json.Unmarshal([]byte(params.AzureVaultCredentials), &azureVaultCredentials)

		if err != nil {
			log.Fatalf("Failed to parse Aws credentials got error : %v", err)
			return "", err
		}
		secretData, err := getAzureSecretString(azureVaultCredentials.Token, azureVaultCredentials.Name, dockerHubCreds.SecretName)

		if err != nil {
			log.Fatalf("Error getting dockerhub secret: %v", err)
			return "", err
		}

		return secretData, nil

	} else {
		return "", errors.New("No credentials provided")
	}
}
