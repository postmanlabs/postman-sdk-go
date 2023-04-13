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
	ConfigOptions pminterfaces.PostmanSDKConfig
}

func (e *PostmanExporter) ExportSpans(ctx context.Context, ss []tracesdk.ReadOnlySpan) error {
	log.Printf("Configuration %+v", e.ConfigOptions)

	truncateData := e.ConfigOptions.Options.TruncateData

	redactData := e.ConfigOptions.Options.RedactSensitiveData["Enable"]

	pmutils.Log.Debug("Spans to be exported are")

	for idx, span := range ss {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Issue faced while running plugins:", r)
			}
		}()

		if truncateData {
			plugins.Truncation(span)
		}

		if (redactData == true || redactData == nil) && e.ConfigOptions.Options.RedactSensitiveData["available"] == true {
			rules := e.ConfigOptions.Options.RedactSensitiveData["Rules"]
			if rules == nil {
				rules = make(map[string]interface{})
			}

			plugins.Redaction(span, rules)
			log.Printf("Rules %+v", rules)
		}
		log.Printf("Debug: span number:%d span:%+v", idx, span)
		pmutils.Log.Debug("Span number:%d span:%+v", idx, span)
	}
	return e.Exporter.ExportSpans(ctx, ss)
}
