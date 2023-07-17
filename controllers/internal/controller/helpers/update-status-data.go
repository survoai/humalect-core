package helpers

import (
	"encoding/json"
	"fmt"
)

func UpdateStatusData(webhookDataString string, step string, success bool) string {
	var WebhookData map[string]interface{}
	err := json.Unmarshal([]byte(webhookDataString), &WebhookData)
	if err != nil {
		fmt.Println("Some error occured while parsing webhook data:= ", err)
	}
	if _, ok := WebhookData["statusData"]; ok {
		WebhookData["statusData"].(map[string]interface{})[step] = success
	} else {
		WebhookData["statusData"] = map[string]interface{}{}
		WebhookData["statusData"].(map[string]interface{})[step] = success
	}
	WebhookData["statusData"].(map[string]interface{})[step] = success
	jsonData, err := json.Marshal(WebhookData)
	return string(jsonData)
}
