package services

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/Humalect/humalect-core/agent/constants"
)

func CloneSourceCode(sourceCodeProvider string,
	sourceCodeOrgName string,
	sourceCodeRepositoryName string,
	commitId string,
	sourceCodeToken string,
) (string, error) {
	fmt.Println("Started Cloning of source code.")

	var repoArchiveURL string
	switch sourceCodeProvider {
	case "github":
		repoArchiveURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/%s", sourceCodeOrgName, sourceCodeRepositoryName, commitId)
	case "gitlab":
		repoArchiveURL = fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/repository/archive/?sha=%s", sourceCodeRepositoryName, commitId)
	case "bitbucket":
		repoArchiveURL = fmt.Sprintf("https://bitbucket.org/%s/get/%s.tar.gz", sourceCodeRepositoryName, commitId)
	default:
		return "", errors.New("Error: invalid source code provider received")
	}

	command := fmt.Sprintf("mkdir %s || true && (curl -L -k \"%s\" -H \"Authorization: Bearer %s\" | tar -xz -C %s --strip-components 1)", constants.TempDirectoryName, repoArchiveURL, sourceCodeToken, constants.TempDirectoryName)
	fmt.Println(command)
	fmt.Println("Cloning to temp folder")
	cmd := exec.Command("sh", "-c", command)

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(stdoutStderr))
		return "", err
	}
	fmt.Println("Cloning execution complete")
	return repoArchiveURL, nil
}
