package util

import (
	"encoding/json"
)

func ParseJson(body []byte, target interface{}) error {
	return json.Unmarshal(body, target)
}
