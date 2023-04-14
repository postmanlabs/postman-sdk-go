package exporter

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
	pmutils "github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
)

type PostmanExporter struct {
	otlptrace.Exporter
	Sdkconfig *pminterfaces.PostmanSDKConfig
}

func (e *PostmanExporter) ExportSpans(ctx context.Context, ss []tracesdk.ReadOnlySpan) error {
	if e.Sdkconfig.IsSuppressed() {
		return nil
	}

	pmutils.Log.Debug("Spans to be exported are")
	for idx, span := range ss {
		pmutils.Log.Debug("Span number:%d span:%+v", idx, span)
	}
	return e.Exporter.ExportSpans(ctx, ss)
}
