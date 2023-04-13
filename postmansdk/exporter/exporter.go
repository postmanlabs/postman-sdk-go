package exporter

import (
	"context"

	plugins "github.com/postmanlabs/postman-go-sdk/postmansdk/exporter/plugins"
	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
	pmutils "github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type PostmanExporter struct {
	otlptrace.Exporter
	Config pminterfaces.PostmanSDKConfig
}

func (e *PostmanExporter) ExportSpans(ctx context.Context, ss []tracesdk.ReadOnlySpan) error {
	pmutils.Log.Debug("Configuration %+v", e.Config)

	truncateDataFlag := e.Config.Options.TruncateData
	redactDataFlag := e.Config.Options.RedactSensitiveData.Enable

	pmutils.Log.Debug("Spans to be exported are")

	for idx, span := range ss {

		if truncateDataFlag {
			plugins.Truncate(span)
		}

		if redactDataFlag {
			rules := e.Config.Options.RedactSensitiveData.Rules
			pmutils.Log.Debug("Rules %+v", rules)
			// plugins.Redaction(span, rules)
		}

		pmutils.Log.Debug("Span number:%d span:%+v", idx, span)
	}
	return e.Exporter.ExportSpans(ctx, ss)
}
