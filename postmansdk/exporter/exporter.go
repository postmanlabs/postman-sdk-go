package exporter

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	plugins "github.com/postmanlabs/postman-go-sdk/postmansdk/exporter/plugins"
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

	for _, span := range ss {
		defer func() {
			if r := recover(); r != nil {
				pmutils.Log.Debug("Issue faced while running plugins.")
			}
		}()

		if e.Sdkconfig.Options.TruncateData {
			plugins.Truncate(span)
		}

		if e.Sdkconfig.Options.RedactSensitiveData.Enable {
			plugins.Redact(span, e.Sdkconfig.Options.RedactSensitiveData.Rules)
		}

	}
	return e.Exporter.ExportSpans(ctx, ss)
}
