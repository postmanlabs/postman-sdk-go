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

const requestRedactionMap = `{
    "body": {
        "attribute_key":      "http.request.body",
        "redaction_function": "redact_body_data"
    },
    "headers": {
        "attribute_key":      "http.request.headers",
        "redaction_function": "redact_headers_data"
    },
    "queryUrl": {
        "attribute_key":      "http.url",
        "redaction_function": "redact_uristring_data"
    },
    "queryString": {
        "attribute_key":      "http.request.query",
        "redaction_function": "redact_query_data"
    },
    "targetUrl": {
        "attribute_key":      "http.target",
        "redaction_function": "redact_uristring_data"
    }
}`

const responseRedactionMap = `{
	"body": {
		"attribute_key":      "http.response.body",
		"redaction_function": "redact_body_data"
	},
	"headers": {
		"attribute_key":      "http.response.headers",
		"redaction_function": "redact_headers_data"
	}
}`

const defaultRedactionReplacementString = "*****"
