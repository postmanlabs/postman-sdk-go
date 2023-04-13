package exporter

import (
	"context"
	"log"

	plugins "github.com/postmanlabs/postman-go-sdk/postmansdk/exporter/plugins"
	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
	pmutils "github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type PostmanExporter struct {
	otlptrace.Exporter
	ConfigOptions pminterfaces.PostmanSDKConfig
}

func (e *PostmanExporter) ExportSpans(ctx context.Context, ss []tracesdk.ReadOnlySpan) error {
	log.Printf("Configuration %+v", e.ConfigOptions)

	truncateData := e.ConfigOptions.Options.TruncateData

	redactData := e.ConfigOptions.Options.RedactSensitiveData["Enable"]

	pmutils.Log.Debug("Spans to be exported are")

	for idx, span := range ss {

		if truncateData {
			plugins.Truncation(span)
		}

		if redactData == true {
			rules := e.ConfigOptions.Options.RedactSensitiveData["Rules"]
			log.Printf("Rules %+v", rules)
			// plugins.Redaction(span, rules)
		}

		pmutils.Log.Debug("Span number:%d span:%+v", idx, span)
	}
	return e.Exporter.ExportSpans(ctx, ss)
}
