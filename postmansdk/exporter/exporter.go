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

	pmutils.Log.Debug("Exporting spans")

	var processedSpans []tracesdk.ReadOnlySpan
	for _, span := range ss {
		if e.Sdkconfig.Options.TruncateData {
			err := plugins.Truncate(span)
			if err != nil {
				pmutils.Log.WithError(err).Error("Truncation failed, span won't be sent to backend")
				continue
			}
		}

		if e.Sdkconfig.Options.RedactSensitiveData.Enable {
			err := plugins.Redact(span, e.Sdkconfig.Options.RedactSensitiveData.Rules)
			if err != nil {
				pmutils.Log.WithError(err).Error("Redaction Failed, span won't be sent to backend")
				continue
			}
		}

		processedSpans = append(processedSpans, span)
	}
	return e.Exporter.ExportSpans(ctx, processedSpans)
}
