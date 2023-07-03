package aws

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Humalect/humalect-core/agent/constants"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

func GetSecretValue(secretName, accessKey, secretKey, region string) (string, error) {
	// Create a session object with the access key and secret key
	if secretName == "" || accessKey == "" || secretKey == "" || region == "" {
		return "", errors.New("Secrets Name or Access Key or Secret Key or Region is empty in DockerHub")
	}
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
