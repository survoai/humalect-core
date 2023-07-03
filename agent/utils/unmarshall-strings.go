package utils

import (
	"encoding/json"
)

func UnmarshalStrings(str string) interface{} {
	var emptyCreds interface{}
	json.Unmarshal([]byte(str), &emptyCreds)
	return emptyCreds
}
