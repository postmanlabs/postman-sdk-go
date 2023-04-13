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
	pmutils.Log.Info("Configuration %+v", e.Config)

	truncateDataFlag := e.Config.Options.TruncateData
	redactDataFlag := e.Config.Options.RedactSensitiveData.RedactionEnable

	pmutils.Log.Debug("Spans to be exported are %+v", ss)

	for idx, span := range ss {
		defer func() {
			if r := recover(); r != nil {
				pmutils.Log.Debug("Issue faced while running plugins.")
			}
		}()

		if truncateDataFlag {
			plugins.Truncate(span)
		}

		if redactDataFlag {
			rules := e.Config.Options.RedactSensitiveData.Rules
			if rules == nil {
				rules = make(map[string]string)
			}

			plugins.Redaction(span, rules)
			pmutils.Log.Debug("Rules %+v", rules)
		}

		pmutils.Log.Debug("Span number:%d span:%+v", idx, span)
	}
	return e.Exporter.ExportSpans(ctx, ss)
}
