package utils

import (
	"errors"
	"fmt"
)

func GetArtifactsRepoLink(cloudProvider string,
	artifactsRepositoryName string,
	commitId string,
	azureAcrRegistryName string,
) (string, error) {
	var artifactsRepoLink string
	if cloudProvider == "aws" {
		artifactsRepoLink = fmt.Sprintf("%s/%s:%s", "awsEcrRegistryUrl", artifactsRepositoryName, commitId)
	} else if cloudProvider == "azure" {
		artifactsRepoLink = fmt.Sprintf("%s.azurecr.io/%s:%s", azureAcrRegistryName, artifactsRepositoryName, commitId)
	} else {
		fmt.Println("Error: Invalid cloudProvider")
		return "", errors.New("Error: invalid cloud Id")
	}
	return artifactsRepoLink, nil
}
