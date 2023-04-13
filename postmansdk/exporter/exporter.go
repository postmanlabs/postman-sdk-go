package exporter

import (
	"context"
	"fmt"
	"log"

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
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Issue faced while running plugins:", r)
			}
		}()

		if truncateDataFlag {
			plugins.Truncate(span)
		}

		if redactDataFlag && e.Config.Options.RedactSensitiveData.Available {
			rules := e.Config.Options.RedactSensitiveData.Rules
			if rules == nil {
				rules = make(map[string]string)
			}

			plugins.Redaction(span, rules)
			log.Printf("Rules %+v", rules)
		}
		log.Printf("Debug: span number:%d span:%+v", idx, span)
		pmutils.Log.Debug("Span number:%d span:%+v", idx, span)
	}
	return e.Exporter.ExportSpans(ctx, ss)
}
