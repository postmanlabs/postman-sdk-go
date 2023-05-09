package plugins

import (
	"encoding/json"
	"reflect"

	pmutils "github.com/postmanlabs/postman-sdk-go/postmansdk/utils"
	"go.opentelemetry.io/otel/attribute"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

var DEFAULT_DATA_TRUNCATION_LEVEL = 2

func Truncate(span tracesdk.ReadOnlySpan) error {
	spanAttributes := span.Attributes()

	for k, v := range spanAttributes {
		if _, ok := attrNameTruncate[string(v.Key)]; !ok {
			continue
		}

		data := spanAttributes[k].Value.AsString()

		truncatedData, err := runTruncate(data)
		if err != nil {
			pmutils.Log.WithError(err).Error("Failed to run Truncation.")
			return err
		}

		spanAttributes[k].Value = attribute.StringValue(truncatedData)
	}
	return nil
}

func runTruncate(data string) (string, error) {
	var jdata interface{}

	err := json.Unmarshal([]byte(data), &jdata)
	if err != nil {
		pmutils.Log.WithError(err).Error("Failed to umarshall data.")
		// Supporting only content-type=application/json
		return data, err
	}

	// 1 represents the starting level.
	truncatedData := trimBodyValuesToTypes(jdata, 1)

	jsonData, err := json.Marshal(truncatedData)
	if err != nil {
		pmutils.Log.WithError(err).Error("Failed to marshall data.")
		return data, err
	}

	return string(jsonData), nil
}

func trimBodyValuesToTypes(data interface{}, currentLevel int) interface{} {

	if data == nil {
		return nil
	}

	switch data := data.(type) {
	case []interface{}:
		trimmedBody := make([]interface{}, 0)

		for _, value := range data {

			if checkRecursive(value, currentLevel) {
				trimmedBody = append(trimmedBody, trimBodyValuesToTypes(value, currentLevel+1))
			} else {
				trimmedBody = append(trimmedBody, getDataType(value))
			}
		}

		return trimmedBody

	case map[string]interface{}:
		trimmedBody := make(map[string]interface{})

		for key, value := range data {

			if checkRecursive(value, currentLevel) {
				trimmedBody[key] = trimBodyValuesToTypes(value, currentLevel+1)
			} else {
				trimmedBody[key] = getDataType(value)
			}
		}

		return trimmedBody

	default:
		return data
	}
}

func getDataType(value interface{}) map[string]interface{} {
	if value == nil {
		return map[string]interface{}{"type": nil}
	}
	//TODO: Precise reasoning for reflect.TypeOf vs reflect.Valueof
	return map[string]interface{}{"type": reflect.TypeOf(value).Kind().String()}
}

func checkRecursive(value interface{}, level int) bool {
	if level > DEFAULT_DATA_TRUNCATION_LEVEL {
		return false
	}
	if value == nil {
		return false
	}

	dtype := reflect.TypeOf(value).Kind()

	if dtype == reflect.Map || dtype == reflect.Slice {
		return true
	}

	return false

}
