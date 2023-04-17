package plugins

import (
	"encoding/json"
	"regexp"

	pmutils "github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
	"go.opentelemetry.io/otel/attribute"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func Redact(span tracesdk.ReadOnlySpan, rules map[string]string) {
	pmutils.Log.Info("Running redaction for span: %+v ", span)
	dr := DataRedaction{ruleNameRegexMap: make(map[string]*regexp.Regexp)}
	dr.compileRules(rules)
	dr.redactData(span)
}

type DataRedaction struct {
	ruleNameRegexMap map[string]*regexp.Regexp
}

func (dr *DataRedaction) compileRules(rules map[string]string) {
	combinedRules := make(map[string]string)
	for k, v := range defaultRedactionRules {
		combinedRules[k] = v
	}
	// User given rules over-ride the default rules.
	for k, v := range rules {
		combinedRules[k] = v
	}

	for rlName, regexRuleCompiled := range combinedRules {
		dr.ruleNameRegexMap[rlName], _ = regexp.Compile("(?i)" + regexRuleCompiled)
	}
}

func (dr *DataRedaction) redactData(span tracesdk.ReadOnlySpan) {
	spanAttributes := span.Attributes()
	for key, value := range spanAttributes {
		attrFunction, attExists := redactionMap[string(value.Key)]
		if attExists {
			data := value.Value.AsString()
			if data == "" {
				continue
			}

			// go over each user defined rules from config and perfrom redaction.
			for _, regEx := range dr.ruleNameRegexMap {
				redactedData := attrFunction(data, regEx)

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

func obfuscateJSONString(jsonString string, regexCompiled *regexp.Regexp) string {
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

func obfuscateString(textContent string, regexCompiled *regexp.Regexp) string {
	return regexCompiled.ReplaceAllString(textContent, defaultRedactionReplacementString)
}

func redactHeadersData(data string, regExCompiled *regexp.Regexp) string {
	if regExCompiled == nil || data == "" {
		return data
	}
	return obfuscateJSONString(data, regExCompiled)
}

func redactBodyData(data string, regExCompiled *regexp.Regexp) string {
	if regExCompiled == nil || data == "" {
		return data
	}
	return obfuscateString(data, regExCompiled)
}

func redactQueryData(data string, regExCompiled *regexp.Regexp) string {
	if regExCompiled == nil || data == "" {
		return data
	}
	return obfuscateJSONString(data, regExCompiled)
}

func redactUriStringData(data string, regExCompiled *regexp.Regexp) string {
	if regExCompiled == nil || data == "" {
		return data
	}
	return obfuscateString(data, regExCompiled)
}
