package Storage

import (
	"encoding/json"
	"fmt"
)

type JSON_TS struct{}

func (j *JSON_TS) Encode(input map[interface{}]interface{}) (string, error) {
	converted := make(map[string]interface{})
	for k, v := range input {
		switch key := k.(type) {
		case string:
			converted[key] = v
		default:
			converted[fmt.Sprintf("%v", key)] = v
		}
	}

	jsonData, err := json.Marshal(converted)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (j *JSON_TS) Decode(jsonData string) (map[interface{}]interface{}, error) {
	var temp map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &temp)
	if err != nil {
		return nil, err
	}

	result := make(map[interface{}]interface{})
	for k, v := range temp {
		result[k] = v
	}

	return result, nil
}

var JSON = new(JSON_TS)
