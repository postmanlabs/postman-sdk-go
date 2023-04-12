package exporter

import (
	"context"
	"fmt"
	"log"

	plugins "github.com/postmanlabs/postman-go-sdk/postmansdk/exporter/plugins"
	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type PostmanExporter struct {
	otlptrace.Exporter
	ConfigOptions pminterfaces.PostmanSDKConfig
}

func (e *PostmanExporter) ExportSpans(ctx context.Context, ss []tracesdk.ReadOnlySpan) error {
	log.Printf("Configuration %+v", e.ConfigOptions)

	truncateData := e.ConfigOptions.ConfigOptions.TruncateData

	redactData := e.ConfigOptions.ConfigOptions.RedactSensitiveData["Enable"]

	for idx, span := range ss {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Issue faced while running plugins:", r)
			}
		}()

		if truncateData {
			plugins.Truncation(span)
		}

		if redactData == true {
			rules := e.ConfigOptions.ConfigOptions.RedactSensitiveData["Rules"]
			plugins.Redaction(span, rules)
			log.Printf("Rules %+v", rules)
		}

		log.Printf("Debug: span number:%d span:%+v", idx, span)
	}
	return e.Exporter.ExportSpans(ctx, ss)
}
