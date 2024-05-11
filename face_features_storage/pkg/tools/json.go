package tools

import (
	"encoding/json"
)

func MarshalStruct(data interface{}) string {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "failed to marshal structure"
	}
	return string(bytes)
}
