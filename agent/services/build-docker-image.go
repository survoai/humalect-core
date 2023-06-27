package services

import (
	"context"
	"fmt"

	"github.com/Humalect/humalect-core/agent/constants"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

func BuildDockerImage(artifactsRepositoryName string,
	artifactsRepoLink string,
	commitId string,
) error {
	fmt.Println("Building Docker image for:", artifactsRepositoryName, commitId)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Println(err)
		return err
	}
	cli.NegotiateAPIVersion(ctx)

	buildContext, err := archive.TarWithOptions(constants.TempDirectoryName, &archive.TarOptions{})
	if err != nil {
		fmt.Println(err)
		return err
	}

	imageBuildResponse, err := cli.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Tags: []string{artifactsRepoLink},
	})
	if err != nil {
		fmt.Println("After image build")

		fmt.Println(err)
		return err
	}
	defer imageBuildResponse.Body.Close()
	return nil
}
