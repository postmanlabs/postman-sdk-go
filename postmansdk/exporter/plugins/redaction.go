package plugins

import (
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

	for name, rule := range combinedRules {
		rCompiled, err := regexp.Compile("(?i)" + rule)
		if err != nil {
			pmutils.Log.WithError(err).Error("Issue while compiling the rules.")
		}
		dr.ruleNameRegexMap[name] = rCompiled
	}
}

func (dr *DataRedaction) redactData(span tracesdk.ReadOnlySpan) {
	spanAttributes := span.Attributes()
	for key, value := range spanAttributes {
		if _, ok := redactAttribute[string(value.Key)]; !ok {
			continue
		}
		data := value.Value.AsString()
		for _, regEx := range dr.ruleNameRegexMap {
			redactedData := obfuscateString(data, regEx)

			// Do nothing when no redaction is performed.
			if data == redactedData {
				continue
			}

			spanAttributes[key].Value = attribute.StringValue(redactedData)
			// Rules are applied in order on input.
			// If we don't update the value, only the last redaction rule is applied.
			data = redactedData
		}
	}
}

func obfuscateString(data string, regExCompiled *regexp.Regexp) string {
	if regExCompiled == nil || data == "" {
		return data
	}

	redactedData := regExCompiled.ReplaceAllString(data, defaultRedactionReplacementString)
	return redactedData
}
