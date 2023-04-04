package exporter

import (
	"context"
	"log"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type PostmanExporter struct {
	otlptrace.Exporter
}

func (e *PostmanExporter) ExportSpans(ctx context.Context, ss []tracesdk.ReadOnlySpan) error {
	log.Printf("DEBUG: Spans to be exported are \n")
	for idx, span := range ss {
		log.Printf("Debug: span number:%d span:%+v", idx, span)
	}
	return e.Exporter.ExportSpans(ctx, ss)
}
