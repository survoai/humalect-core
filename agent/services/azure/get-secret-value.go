package azure

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func GetSecretValue(azureVaultToken string, vaultName string, secretName string) (string, error) {
	if azureVaultToken == "" || vaultName == "" || secretName == "" {
		return "", errors.New("Azure Vault Token or Vault Name or Secret Name  is empty in DockerHub")
	}
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
