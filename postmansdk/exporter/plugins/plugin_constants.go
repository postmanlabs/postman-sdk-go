package plugins

import "regexp"

var spanHttpBodyAttributesName = map[string]string{
	"response": "http.response.body",
	"request":  "http.request.body",
}

var defaultRedactionRules = map[string]string{
	"pmPostmanAPIKey":    `PMAK-[a-f0-9]{24}-[a-f0-9]{34}`,
	"pmPostmanAccessKey": `PMAT-[0-9a-z]{26}`,
	"pmBasicAuth":        `Basic [a-zA-Z0-9]{3,1000}(?:[^a-z0-9+({})!@#$%^&|*=]{0,2})`,
	"pmBearerToken":      `Bearer [a-z0-9A-Z-._~+/]{15,1000}`,
}

type redactFunction func(data string, regExCompiled *regexp.Regexp) string

var redactionMap = map[string]redactFunction{
	"http.request.body":     redactBodyData,
	"http.request.headers":  redactHeadersData,
	"http.url":              redactUriStringData,
	"http.request.query":    redactQueryData,
	"http.target":           redactUriStringData,
	"http.response.body":    redactBodyData,
	"http.response.headers": redactHeadersData,
}

const defaultRedactionReplacementString = "*****"
