package plugins

import (
	"encoding/json"
	"regexp"

	pmutils "github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
	"go.opentelemetry.io/otel/attribute"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func Redaction(span tracesdk.ReadOnlySpan, rules map[string]string) {
	pmutils.Log.Info("Running redaction for span: %+v ", span)
	dr := DataRedaction{regexRedaction: make(map[string]*regexp.Regexp)}
	dr.compileRules(rules)
	dr.runRedaction(span)
}

type DataRedaction struct {
	regexRedaction map[string]*regexp.Regexp
}

func (dr *DataRedaction) compileRules(rules map[string]string) {
	combinedRules := make(map[string]string)
	for k, v := range defaultRedactionRules {
		combinedRules[k] = v
	}
	// Making sure that user rules are given priority.
	// In case of conflict, the items from rules will override the DEFAULT_REDACTION_RULES.

	for k, v := range rules {
		combinedRules[k] = v
	}

	for rlName, regexRuleCompiled := range combinedRules {
		dr.regexRedaction[rlName] = regexp.MustCompile("(?i)" + regexRuleCompiled)
	}
}

func (dr *DataRedaction) runRedaction(span tracesdk.ReadOnlySpan) {
	for _, requestSection := range []string{"request", "response"} {
		dr.redactData(requestSection, span)
	}
}

func (dr *DataRedaction) redactData(requestSection string, span tracesdk.ReadOnlySpan) {
	var redactionMap map[string]map[string]string
	var requestRedactionRuleSet map[string]map[string]string
	err := json.Unmarshal([]byte(requestRedactionMap), &requestRedactionRuleSet)
	if err != nil {
		pmutils.Log.WithError(err).Error("Issue while reading the request redaction map.")
	}

	var responseRedactionRuleSet map[string]map[string]string
	errr := json.Unmarshal([]byte(responseRedactionMap), &responseRedactionRuleSet)
	if errr != nil {
		pmutils.Log.WithError(err).Error("Issue while reading the response redaction map.")
	}

	if requestSection == "request" {
		redactionMap = requestRedactionRuleSet
	} else if requestSection == "response" {
		redactionMap = responseRedactionRuleSet
	} else {
		return
	}

	if redactionMap == nil {
		return
	}

	spanAttributes := span.Attributes()
	for key, value := range spanAttributes {
		for _, redactConfig := range redactionMap {
			if string(value.Key) == redactConfig["attribute_key"] {
				data := value.Value.AsString()
				if data == "" {
					continue
				}

				// go over each user defined rules from config and perfrom redaction.
				for _, regEx := range dr.regexRedaction {
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

					if data != redactedData {
						jsonStr, err := json.Marshal(redactedData)
						if err != nil {
							pmutils.Log.WithError(err).Error("Issue while pasring the redacted data.")
						}

						spanAttributes[key].Value = attribute.StringValue(string(jsonStr))
						data = redactedData
					}
				}
			}
		}
	}
}

func (dr *DataRedaction) obfuscateJSONString(jsonString string, regexCompiled *regexp.Regexp) string {
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
			// If the available data is not string, and of some complex type.
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

func (dr *DataRedaction) obfuscateString(textContent string, regexCompiled *regexp.Regexp) string {
	return regexCompiled.ReplaceAllString(textContent, defaultRedactionReplacementString)
}

func (dr *DataRedaction) redactHeadersData(data string, regExCompiled *regexp.Regexp) string {
	if regExCompiled == nil || data == "" {
		return data
	}
	return dr.obfuscateJSONString(data, regExCompiled)
}

func (dr *DataRedaction) redactBodyData(data string, regExCompiled *regexp.Regexp) string {
	if regExCompiled == nil || data == "" {
		return data
	}
	return dr.obfuscateString(data, regExCompiled)
}

func (dr *DataRedaction) redactQueryData(data string, regExCompiled *regexp.Regexp) string {
	if regExCompiled == nil || data == "" {
		return data
	}
	return dr.obfuscateJSONString(data, regExCompiled)
}

func (dr *DataRedaction) redactUriStringData(data string, regExCompiled *regexp.Regexp) string {
	if regExCompiled == nil || data == "" {
		return data
	}
	return dr.obfuscateString(data, regExCompiled)
}
