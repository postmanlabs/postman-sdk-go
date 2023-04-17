package plugins

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

var redactAttribute = map[string]bool{
	"http.request.body":     true,
	"http.request.headers":  true,
	"http.url":              true,
	"http.request.query":    true,
	"http.target":           true,
	"http.response.body":    true,
	"http.response.headers": true,
}

const defaultRedactionReplacementString = "*****"
