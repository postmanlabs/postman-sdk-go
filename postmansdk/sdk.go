package postmansdk

import (
	"net/http"
	"regexp"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// newExporter returns a console exporter.
func newExporter() (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		// stdouttrace.WithWriter(w),
		// Use human-readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
}

// func newResource()

// getFilters reads the ignoreIncomingRequests array and produces a otelgin.Option
// to fiter out requests that match the regex.
func getFilters() otelgin.Option {
	if len(ignoreIncomingRequests) == 0 {
		return nil
	}

	return otelgin.WithFilter(func(r *http.Request) bool {
		for _, f := range ignoreIncomingRequests {
			matches, _ := regexp.MatchString(f, r.URL.Path)

			if matches {
				return false
			}
		}

		return true
	})
}

// getMiddlewareOptions returns an array of otelgin.Option(s), these can include
// filters, propagators or formatters.
//
// If no option has been selected by the user, an empty array is returned.
func getMiddlewareOptions() []otelgin.Option {
	var middlewareOptions []otelgin.Option

	// Add filters
	f := getFilters()
	if f != nil {
		middlewareOptions = append(middlewareOptions, f)
	}

	return middlewareOptions
}
