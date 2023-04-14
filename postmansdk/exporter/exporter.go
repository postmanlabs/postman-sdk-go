package exporter

import (
	"context"
	"fmt"

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

	for idx, span := range ss {
		defer func() {
			if r := recover(); r != nil {
				pmutils.Log.Debug("Issue faced while running plugins.")
			}
		}()

		if e.Sdkconfig.Options.TruncateData {
			plugins.Truncate(span)
		}

		if e.Sdkconfig.Options.RedactSensitiveData.RedactionEnable {
			rules := e.Sdkconfig.Options.RedactSensitiveData.Rules
			if rules == nil {
				rules = make(map[string]string)
			}

			plugins.Redaction(span, rules)
			pmutils.Log.Debug("Rules %+v", rules)
		}

		pmutils.Log.Debug(fmt.Printf("Span number:%d span:%+v", idx, span))
	}
	return e.Exporter.ExportSpans(ctx, ss)
}
