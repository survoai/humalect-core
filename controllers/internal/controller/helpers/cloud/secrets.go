package cloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	k8sv1 "github.com/Humalect/humalect-core/api/v1"
	constants "github.com/Humalect/humalect-core/internal/controller/constants"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func GetCloudSecretMap(application *k8sv1.Application) (map[string]string, error) {
	var secretsMap map[string]string
	var secretString string
	var err error
	if application.Spec.SecretsProvider == constants.CloudIdAWS || (application.Spec.SecretsProvider == "" && application.Spec.CloudProvider == constants.CloudIdAWS) {
		config, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(application.Spec.AwsSecretCredentials.AccessKey, application.Spec.AwsSecretCredentials.SecretKey, "")))
		if err != nil {
			log.Fatal(err)
			return map[string]string{}, err
		}

		secretString, err = getAwsSecretString(config, application.Spec.SecretManagerName, application.Spec.AwsSecretCredentials.Region)

	} else if application.Spec.SecretsProvider == constants.CloudIdAzure || (application.Spec.SecretsProvider == "" && application.Spec.CloudProvider == constants.CloudIdAzure) {
		secretString, err = getAzureSecretString(application.Spec.AzureVaultCredentials.Token, application.Spec.AzureVaultCredentials.Name, application.Spec.SecretManagerName)
		if err != nil {
			log.Fatal(err.Error())
			return map[string]string{}, err
		}
	} else {
		return map[string]string{}, errors.New("Invalid Secrets Provider for secrets")
	}
	err = json.Unmarshal([]byte(secretString), &secretsMap)
	if err != nil {
		log.Fatal(err.Error())
		return map[string]string{}, err
	}
	return secretsMap, nil
}

func getAwsSecretString(config aws.Config, secretName string, region string) (string, error) {
	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Fatal(err.Error())
		return "", err
	}

	var secretString string = *result.SecretString

	return secretString, nil
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
