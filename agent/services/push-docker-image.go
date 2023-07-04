package services

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
)

func PushDockerImage(cloudProvider string,
	artifactsRepoLink string,
) error {
	type ArtifactsRegistryCredentials struct {
		Username      string `json:"username"`
		Password      string `json:"password"`
		ServerAddress string `json:"serverAddress"`
	}

	var artifactsRegistryCredentials ArtifactsRegistryCredentials
	if cloudProvider == "aws" {
		ctx := context.Background()

		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			fmt.Println(err)
			return err
		}

		client := ecr.NewFromConfig(cfg)
		input := &ecr.GetAuthorizationTokenInput{}
		resp, err := client.GetAuthorizationToken(ctx, input)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if len(resp.AuthorizationData) == 0 {
			return errors.New("no authorization data found")
		}

		authData := resp.AuthorizationData[0]
		decodedToken, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
		if err != nil {
			fmt.Println(err)
			return err
		}

		tokenParts := strings.Split(string(decodedToken), ":")
		if len(tokenParts) < 2 {
			// return "", fmt.Errorf("invalid authorization token format")
			return errors.New("invalid authorization token format")
		}

		artifactsRegistryCredentials = ArtifactsRegistryCredentials{
			Username:      "awsEcrUserName",
			Password:      tokenParts[1],
			ServerAddress: "awsEcrRegistryUrl",
		}

	} else if cloudProvider == "azure" {

		url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ContainerRegistry/registries/%s/listCredentials?api-version=2019-05-01",
			"azureSubscriptionId", "azureResourceGroupName", "azureAcrRegistryName")

		client := &http.Client{}

		req, err := http.NewRequest("POST", url, strings.NewReader("{}"))
		if err != nil {
			fmt.Println(err)
			return err
		}

		req.Header.Set("Authorization", "Bearer "+"azureManagementScopeToken")
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return err
		}

		var data struct {
			Username  string `json:"username"`
			Passwords []struct {
				Value string `json:"value"`
			} `json:"passwords"`
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if data.Username == "" {
			return errors.New("Invalid Username")
		}

		artifactsRegistryCredentials = ArtifactsRegistryCredentials{
			Username:      data.Username,
			Password:      data.Passwords[0].Value,
			ServerAddress: "azureAcrRegistryName" + ".azurecr.io",
		}

	} else {
		fmt.Println("Error: Invalid cloudProvider")
		return errors.New("Error: Invalid cloudProvider")

	}

	// fmt.Println(artifactsRegistryCredentials)
	pushDockerImageContext := context.Background()
	pushDockerImageClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Println(err)
		return err
	}
	pushDockerImageClient.NegotiateAPIVersion(pushDockerImageContext)
	encodedAuth, _ := json.Marshal(artifactsRegistryCredentials)
	authStr := base64.URLEncoding.EncodeToString(encodedAuth)
	options := types.ImagePushOptions{
		RegistryAuth: authStr,
	}
	response, err := pushDockerImageClient.ImagePush(pushDockerImageContext, artifactsRepoLink, options)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer response.Close()

	err = handleProgressMessages(response)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("<======== Finished Pushing Docker file =======>")
	fmt.Println("Pushed image repo link:", artifactsRepoLink)

	return nil
}

func handleProgressMessages(response io.ReadCloser) error {
	scanner := bufio.NewScanner(response)
	for scanner.Scan() {
		var event jsonmessage.JSONMessage

		err := json.Unmarshal(scanner.Bytes(), &event)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if event.Status != "" {
			fmt.Println(event.Status)
		}

	}

	return scanner.Err()
}
