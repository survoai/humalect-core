package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Humalect/humalect-core/agent/constants"
)

func SendWebhookRequest(webhookEndpoint string, data map[string]interface{}) (response *WebhookResponse, err error) {
	apiURL := webhookEndpoint

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return &WebhookResponse{
			Success: true,
			Status:  resp.StatusCode,
			Data:    body,
		}, nil
	}

	return &WebhookResponse{
		Success: false,
		Status:  resp.StatusCode,
		Data:    body,
	}, nil
}

func CreateSendWebhookRequest(webhookEndpoint string, data map[string]interface{}) (response *WebhookResponse, err error) {
	response, err = SendWebhookRequest(webhookEndpoint, data)
	if err != nil {
		fmt.Printf("Error Sending Webhook: %v\n", err)
	}

	if response.Success {
		fmt.Printf("Sent Webhook Status: %d\n", response.Status)
		fmt.Printf("Sent Webhook Body: %v\n", response.Data)
	} else {
		fmt.Printf("Error Response Received while Sending Webhook: %v\n", response.Data)
	}
	return response, err
}

func SendWebhook(WebhookEndpoint string, data string, success bool, state string) {
	if len(WebhookEndpoint) > 0 {
		var WebhookData map[string]interface{}
		err := json.Unmarshal([]byte(data), &WebhookData)
		if err != nil {
			fmt.Println("Some error occured while parsing webhook data:= ", err)
		}

		var strStatus string
		if success {
			strStatus = state
		} else {
			strStatus = constants.DeploymentFailed
		}
		if _, ok := WebhookData["statusData"]; ok {
			WebhookData["statusData"].(map[string]interface{})[state] = success
		} else {
			WebhookData["statusData"] = map[string]interface{}{}
			WebhookData["statusData"].(map[string]interface{})[state] = success
		}
		webhookData := map[string]interface{}{
			"type": constants.WebhookTypeDeploymentStatusUpdate,
			"data": map[string]interface{}{
				"queueName":  WebhookData["queueName"],
				"statusData": WebhookData["statusData"],
				"state":      WebhookData["state"],
				"status":     strStatus,
			},
		}
		CreateSendWebhookRequest(WebhookEndpoint, webhookData)
	}
}

type WebhookResponse struct {
	Success bool
	Status  int
	Data    []byte
}
