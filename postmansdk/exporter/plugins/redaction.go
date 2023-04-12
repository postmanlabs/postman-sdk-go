package plugins

import (
	"encoding/json"
	"fmt"
	"regexp"

	"go.opentelemetry.io/otel/attribute"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func Redaction(span tracesdk.ReadOnlySpan, rules interface{}) {

	fmt.Println("We are redacting")
	dr := DataRedaction{regexRedaction: make(map[string]*regexp.Regexp)}
	// fmt.Println("Rules are combiled ----------------- ")
	dr.compileRules(rules)
	// fmt.Println(dr.regexRedaction)

	dr.runRedaction(span)

	// spanAttributes := span.Attributes()

	// for key, value := range spanAttributes {
	// 	fmt.Printf("Value: %+v\n", value)
	// 	data := spanAttributes[key].Value.AsString()

	// 	var finalData interface{}

	// 	fmt.Printf("Value: %+v\n", data)

	// 	err := json.Unmarshal([]byte(data), &finalData)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	fmt.Printf("We are here KB")

	// 	redactedData := dr.runRedaction(finalData)

	// 	jsonStr, err := json.Marshal(redactedData)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	spanAttributes[key].Value = attribute.StringValue(string(jsonStr))
	// }
}

type DataRedaction struct {
	regexRedaction map[string]*regexp.Regexp
}

func (dr *DataRedaction) compileRules(rules interface{}) {

	combinedRules := make(map[string]interface{})

	for k, v := range defaultRedactionRules {
		combinedRules[k] = v
	}
	// Making sure that user rules are given priority.
	// In case of conflict, the items from rules will override the DEFAULT_REDACTION_RULES.

	ruleMap := rules.(map[string]interface{})
	for k, v := range ruleMap {
		combinedRules[k] = v
	}

	for rlName, regexRuleCompiled := range combinedRules {
		fmt.Println(dr.regexRedaction)

		dr.regexRedaction[rlName] = regexp.MustCompile("(?i)" + regexRuleCompiled.(string))
	}
}

func (dr *DataRedaction) runRedaction(span tracesdk.ReadOnlySpan) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Unable to perform redaction error %v\n", r)
		}
	}()

	for _, requestSection := range []string{"request", "response"} {
		fmt.Printf("Request Section  Bansal : %+v \n", requestSection)
		dr.redactData(requestSection, span)
	}

	// attributes[POSTMAN_DATA_REDACTION_SPAN_ATTRIBUTE] = "true"
	// return attributes
}

func (dr *DataRedaction) redactData(requestSection string, span tracesdk.ReadOnlySpan) {
	var redactionMap map[string]map[string]string

	fmt.Printf("Atleast we are here xD  ")

	var requestRedactionRuleSet map[string]map[string]string
	err := json.Unmarshal([]byte(requestRedactionMap), &requestRedactionRuleSet)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Check me please %+v ", requestRedactionRuleSet)

	var responseRedactionRuleSet map[string]map[string]string
	errr := json.Unmarshal([]byte(responseRedactionMap), &responseRedactionRuleSet)
	if errr != nil {
		fmt.Printf("Unknowm panic")
		panic(err)
	}

	fmt.Printf("Did we reach here >>>>>. we are here xD  ")

	if requestSection == "request" {
		redactionMap = requestRedactionRuleSet
	} else if requestSection == "response" {
		redactionMap = responseRedactionRuleSet
	} else {
		return
	}

	fmt.Printf("Redaction Map Kartikay: %+v \n", redactionMap)

	/* eslint-disable keyword-spacing */

	if redactionMap == nil {
		return
	}

	spanAttributes := span.Attributes()

	for key, value := range spanAttributes {
		for _, redactConfig := range redactionMap {
			if string(value.Key) == redactConfig["attribute_key"] {
				fmt.Printf("Redaction Config: %+v %+v\n", key, value)
				data := value.Value.AsString()

				if data == "" {
					continue
				}

				// go over each user defined rules from config and perfrom redaction.
				for _, regEx := range dr.regexRedaction {
					fmt.Printf("Redaction Config: %+v %+v\n", redactConfig, regEx)

					redactedData := data

					switch redactConfig["redaction_function"] {
					case "redact_headers_data":
						redactedData = dr.redactHeadersData(data, regEx)
					case "redact_body_data":
						redactedData = dr.redactBodyData(data, regEx)
					case "redact_query_data":
						redactedData = dr.redactQueryData(data, regEx)
					case "redact_uristring_data":
						redactedData = dr.redactUriStringData(data, regEx)
					}

					fmt.Printf("Data -------  %+v", data)

					if data != redactedData {

						fmt.Printf("Aaaaaooooooo nachoooooooooo -------  %+v", redactedData)
						jsonStr, err := json.Marshal(redactedData)
						if err != nil {
							panic(err)
						}

						spanAttributes[key].Value = attribute.StringValue(string(jsonStr))

						data = redactedData
					}

				}
			}
		}
	}
}

func (dr *DataRedaction) __obfuscateJSONString(jsonString string, regexCompiled *regexp.Regexp) string {
	jsonObj := make(map[string]interface{})

	err := json.Unmarshal([]byte(jsonString), &jsonObj)
	if err != nil {
		return jsonString
	}

	for keyName, val := range jsonObj {
		var dataVal string

		if strVal, ok := val.(string); ok {
			dataVal = strVal
		} else {
			valBytes, err := json.Marshal(val)
			if err != nil {
				continue
			}

			dataVal = string(valBytes)
		}

		jsonObj[keyName] = regexCompiled.ReplaceAllString(dataVal, defaultRedactionReplacementString)
	}

	jsonBytes, err := json.Marshal(jsonObj)
	if err != nil {
		return jsonString
	}

	return string(jsonBytes)
}

func (dr *DataRedaction) __obfuscateString(textContent string, regexCompiled *regexp.Regexp) string {
	return regexCompiled.ReplaceAllString(textContent, defaultRedactionReplacementString)
}

func (dr *DataRedaction) redactHeadersData(data string, regExCompiled *regexp.Regexp) string {
	if regExCompiled == nil || data == "" {
		return data
	}
	return dr.__obfuscateJSONString(data, regExCompiled)
}

func (dr *DataRedaction) redactBodyData(data string, regExCompiled *regexp.Regexp) string {
	if regExCompiled == nil || data == "" {
		return data
	}
	return dr.__obfuscateString(data, regExCompiled)
}

func (dr *DataRedaction) redactQueryData(data string, regExCompiled *regexp.Regexp) string {
	if regExCompiled == nil || data == "" {
		return data
	}
	return dr.__obfuscateJSONString(data, regExCompiled)
}

func (dr *DataRedaction) redactUriStringData(data string, regExCompiled *regexp.Regexp) string {
	if regExCompiled == nil || data == "" {
		return data
	}
	return dr.__obfuscateString(data, regExCompiled)
}
