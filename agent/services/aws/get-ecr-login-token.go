package aws

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

func getEcrLoginToken(accessKey string, secretKey string, region string) (string, error) {
	if accessKey == "" || secretKey == "" || region == "" {
		return "", fmt.Errorf("Error getting ECR login token: Missing accessKey, secretKey or region")
	}
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
