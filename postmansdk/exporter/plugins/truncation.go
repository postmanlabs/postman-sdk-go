package plugins

import (
	"encoding/json"
	"reflect"

	pmutils "github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
	"go.opentelemetry.io/otel/attribute"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

var DEFAULT_DATA_TRUNCATION_LEVEL = 2

func Truncate(span tracesdk.ReadOnlySpan) {
	pmutils.Log.Debug("Truncating data for span : %+v ", span)

	spanHttpBodyAttributesName := map[string]interface{}{
		"response": "http.response.body",
		"request":  "http.request.body",
	}

	spanAttributes := span.Attributes()

	for k, v := range spanAttributes {
		for attributeType, attributeName := range spanHttpBodyAttributesName {
			if string(v.Key) == attributeName {

				pmutils.Log.Debug("Running truncation for %+v at %+v \n", attributeType, attributeName)

				data := spanAttributes[k].Value.AsString()

				var finalData interface{}

				err := json.Unmarshal([]byte(data), &finalData)
				if err != nil {
					pmutils.Log.Debug(err)
				}

				truncatedData := trimBodyValuesToTypes(finalData, 1)

				jsonData, err := json.Marshal(truncatedData)
				if err != nil {
					pmutils.Log.Debug(err)
				}

				spanAttributes[k].Value = attribute.StringValue(string(jsonData))
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
			/*
				Using reflect.TypeOf(value).Kind() as it provides the most generic and direct comparison to types like
				reflect.Map orreflect.Slice.
				We enter this code block only when the current data type is slice. If, we find complex data further,
				recursive function is called, else the current data type is assigned. Since, we are limiting the
				truncation level to 2, that is also taken here.
			*/
			if currentLevel <= DEFAULT_DATA_TRUNCATION_LEVEL && value != nil && reflect.TypeOf(value).Kind() == reflect.Map {
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
			/*
				Similarly, we enter this code block only when the current data type is map. If, we find complex data further,
				recursive function is called, else the current data type is assigned. Since, we are limiting the
				truncation level to 2, that is also taken here.
			*/
			if currentLevel <= DEFAULT_DATA_TRUNCATION_LEVEL && value != nil && reflect.TypeOf(value).Kind() == reflect.Map {
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
