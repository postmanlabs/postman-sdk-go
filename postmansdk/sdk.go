package postmansdk

import (
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
