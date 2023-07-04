package azure

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type AzureCreds struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func FetchAcrCreds(AzureManagementScopeToken string, AzureAcrRegistryName string, AzureSubscriptionId string, AzureResourceGroupName string) (AzureCreds, error) {
	fmt.Println("AzureManagementScopeToken", AzureManagementScopeToken)
	fmt.Println("AzureAcrRegistryName", AzureAcrRegistryName)
	fmt.Println("AzureSubscriptionId", AzureSubscriptionId)
	fmt.Println("AzureResourceGroupName", AzureResourceGroupName)
	if AzureManagementScopeToken == "" || AzureAcrRegistryName == "" || AzureSubscriptionId == "" || AzureResourceGroupName == "" {
		return AzureCreds{}, errors.New("Azure Management Scope Token, Azure ACR Registry Name, Azure Subscription Id and Azure Resource Group Name are required.")
	}
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
