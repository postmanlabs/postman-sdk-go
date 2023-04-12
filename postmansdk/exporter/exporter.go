package exporter

import (
	"context"
	"log"

	plugins "github.com/postmanlabs/postman-go-sdk/postmansdk/exporter/plugins"
	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type PostmanExporter struct {
	otlptrace.Exporter
	ConfigOptions pminterfaces.PostmanSDKConfig
}

func (e *PostmanExporter) ExportSpans(ctx context.Context, ss []tracesdk.ReadOnlySpan) error {
	log.Printf("Configuration %+v", e.ConfigOptions)

	truncateData := e.ConfigOptions.ConfigOptions.TruncateData
	if truncateData {
		plugins.Truncation()
	}

	redactData := e.ConfigOptions.ConfigOptions.RedactSensitiveData["Enable"]
	if redactData == true {
		rules := e.ConfigOptions.ConfigOptions.RedactSensitiveData["Rules"]
		log.Printf("Rules %+v", rules)
		plugins.Redaction()
	}

	for idx, span := range ss {
		log.Printf("Debug: span number:%d span:%+v", idx, span)
	}
	return e.Exporter.ExportSpans(ctx, ss)
}
