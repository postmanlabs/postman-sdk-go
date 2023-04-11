package exporter

import (
	"context"
	"log"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
	plugins "github.com/postmanlabs/postman-go-sdk/postmansdk/exporter/plugins"
)

type PostmanExporter struct {
	otlptrace.Exporter
	ConfigOptions pminterfaces.PostmanSDKConfig
}

func (e *PostmanExporter) ExportSpans(ctx context.Context, ss []tracesdk.ReadOnlySpan) error {
	log.Printf("Configuration %+v", e.ConfigOptions)

	plugins.Redaction()
	plugins.Truncation()

	for idx, span := range ss {
		log.Printf("Debug: span number:%d span:%+v", idx, span)
	}
	return e.Exporter.ExportSpans(ctx, ss)
}
