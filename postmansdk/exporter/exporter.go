package exporter

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/internal/tracetransform"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

//	type MyExporter struct {
//		client string
//	}
type PostmanExporter struct {
	otlptrace.Exporter
}

func (e *PostmanExporter) ExportSpans(ctx context.Context, ss []tracesdk.ReadOnlySpan) error {
	protoSpans := tracetransform.Spans(ss)
	if len(protoSpans) == 0 {
		return nil
	}

	err := e.client.UploadTraces(ctx, protoSpans)
	if err != nil {
		return WrapTracesError(err)
	}
	return nil
}
