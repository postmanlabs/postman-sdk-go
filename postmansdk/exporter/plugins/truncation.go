package plugins

import (
	"encoding/json"
	"fmt"
	"reflect"

	"go.opentelemetry.io/otel/attribute"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

var DEFAULT_DATA_TRUNCATION_LEVEL int = 2

func Truncation(span tracesdk.ReadOnlySpan) {
	fmt.Println("We are truncating the data")
	spanAttributes := span.Attributes()

	for key, value := range spanAttributes {
		for attributeType, attributeName := range spanHttpBodyAttributesName {
			if string(value.Key) == attributeName {

				fmt.Printf("Running truncation for %+v at %+v \n", attributeType, attributeName)

				data := spanAttributes[key].Value.AsString()

				var finalData interface{}

				err := json.Unmarshal([]byte(data), &finalData)
				if err != nil {
					panic(err)
				}

				truncatedData := trimBodyValuesToTypes(finalData, 1)

				jsonStr, err := json.Marshal(truncatedData)
				if err != nil {
					panic(err)
				}

				spanAttributes[key].Value = attribute.StringValue(string(jsonStr))
			}
		}
	}
}

func trimBodyValuesToTypes(data interface{}, currentLevel int) interface{} {

	if data == nil {
		return nil
	}

	switch data.(type) {
	case string:
		var parsedData interface{}
		err := json.Unmarshal([]byte(data.(string)), &parsedData)
		if err != nil {
			// If the data is not JSON parsable, it does not make sense to continue, as we only support
			// content-type = application/json at the moment.
			return data
		}
		data = parsedData
	case []interface{}:
		trimmedBody := make([]interface{}, 0)
		for _, value := range data.([]interface{}) {
			if currentLevel <= DEFAULT_DATA_TRUNCATION_LEVEL && value != nil && (reflect.TypeOf(value).Kind() == reflect.Map || reflect.TypeOf(value).Kind() == reflect.Slice) {
				trimmedBody = append(trimmedBody, trimBodyValuesToTypes(value, currentLevel+1))
			} else if value == nil {
				trimmedBody = append(trimmedBody, map[string]interface{}{"type": nil})
			} else {
				trimmedBody = append(trimmedBody, map[string]interface{}{"type": reflect.TypeOf(value).Kind().String()})
			}
		}
		return trimmedBody
	case map[string]interface{}:
		trimmedBody := make(map[string]interface{})
		for key, value := range data.(map[string]interface{}) {
			if currentLevel <= DEFAULT_DATA_TRUNCATION_LEVEL && value != nil && (reflect.TypeOf(value).Kind() == reflect.Map || reflect.TypeOf(value).Kind() == reflect.Slice) {
				trimmedBody[key] = trimBodyValuesToTypes(value, currentLevel+1)
			} else if value == nil {
				trimmedBody[key] = map[string]interface{}{"type": nil}
			} else {
				trimmedBody[key] = map[string]interface{}{"type": reflect.TypeOf(value).Kind().String()}
			}
		}
		return trimmedBody
	default:
		return data
	}
	return data
}
