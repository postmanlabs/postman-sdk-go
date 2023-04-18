package plugins

import (
	"errors"
	"regexp"

	pmutils "github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
	"go.opentelemetry.io/otel/attribute"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func Redact(span tracesdk.ReadOnlySpan, rules map[string]string) error {
	pmutils.Log.WithField("span", span).Info("Running redaction for span")
	dr := DataRedaction{ruleNameRegexMap: make(map[string]*regexp.Regexp)}
	err := dr.compileRules(rules)
	if err != nil {
		return err
	}

	err = dr.redactData(span)
	if err != nil {
		return err
	}
	return nil
}

type DataRedaction struct {
	ruleNameRegexMap map[string]*regexp.Regexp
}

func (dr *DataRedaction) compileRules(rules map[string]string) error {
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
			return err
		}
		dr.ruleNameRegexMap[name] = rCompiled
	}
	return nil
}

func (dr *DataRedaction) redactData(span tracesdk.ReadOnlySpan) error {
	var err error
	// precautionary measure - no explicit error thrown, so handling using defer.
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("issue faced while running redactData")
		}
	}()

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
	if err != nil {
		return err
	}
	return nil
}

func obfuscateString(data string, regExCompiled *regexp.Regexp) string {
	if regExCompiled == nil || data == "" {
		return data
	}

	redactedData := regExCompiled.ReplaceAllString(data, defaultRedactionReplacementString)
	return redactedData
}
