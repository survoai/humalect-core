package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/Humalect/humalect-core/agent/constants"
)

func AddCustomDockerfile(UseDockerFromCodeFlag bool,
	dockerManifest string,
	sourceCodeRepositoryName string,
	commitId string,
	sourceCodeToken string,
) error {
	if !UseDockerFromCodeFlag {
		dockerFileName := "Dockerfile"
		var dockerCommands []string
		err := json.Unmarshal([]byte(dockerManifest), &dockerCommands)
		if err != nil {
			log.Fatal("Error unmarshalling DockerManifest JSON: ", err)
			return err
		}
		dockerFileContent := strings.Join(dockerCommands, "\r\n")

		if dockerManifest != "" && len(dockerFileContent) > 0 {

			dockerFilePath := fmt.Sprintf("%s/%s", constants.TempDirectoryName, dockerFileName)
			err = ioutil.WriteFile(dockerFilePath, []byte(dockerFileContent), 0o644)
			if err != nil {
				log.Fatal("Error writing Dockerfile: ", err)
				return err
			}
		}
	}
	return nil
}
