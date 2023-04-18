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

	pmutils.Log.Debug("Spans to be exported are")

	var processedSpans []tracesdk.ReadOnlySpan
	var err error
	for _, span := range ss {
		defer func() {
			if r := recover(); r != nil {
				pmutils.Log.Debug("Issue faced while running plugins.")
			}
		}()

		if e.Sdkconfig.Options.TruncateData {
			err = plugins.Truncate(span)
			if err != nil {
				pmutils.Log.WithError(err).Error("Failure in truncation.")
				pmutils.Log.WithField("span", span).Info("Skipping the span.")
				continue
			}
		}

		if e.Sdkconfig.Options.RedactSensitiveData.Enable {
			err = plugins.Redact(span, e.Sdkconfig.Options.RedactSensitiveData.Rules)
			if err != nil {
				pmutils.Log.WithError(err).Error("Failure in redaction.")
				pmutils.Log.WithField("span", span).Info("Skipping the span.")
				continue
			}
		}

		processedSpans = append(processedSpans, span)
		pmutils.Log.WithField("span attributes - ", span.Attributes()).Info("Span - ")
	}
	return e.Exporter.ExportSpans(ctx, processedSpans)
}
