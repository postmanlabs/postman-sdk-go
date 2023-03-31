package postmansdk

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	pminterfaces "github.com/postmanlabs/postmansdk/interfaces"
	pmreceiver "github.com/postmanlabs/postmansdk/receiver"
)

type postmanSDK struct {
	Config        pminterfaces.PostmanSDKConfig
	Exporter      sdktrace.SpanExporter
	SpanProcessor sdktrace.SpanProcessor
}

func Initialize(
	collectionId string,
	apiKey string,
	options ...pminterfaces.PostmanSDKConfigOption,
) {

	sdkconfig := pminterfaces.InitializeSDKConfig(collectionId, apiKey, options...)
	log.Printf("SdkConfig is intialized as %+v", sdkconfig)

	psdk := &postmanSDK{
		Config: sdkconfig,
	}

	resp, err := pmreceiver.CallBootStrapAPI(sdkconfig)
	if err != nil {
		log.Println(err)

	}

	if !resp {
		return
	}

	go pmreceiver.CallHealthCheckAPI(sdkconfig)

	// TODO: Should this be passed from request handler ?
	ctx := context.Background()

	// Registers a tracer Provider globally.
	shutdown, err := psdk.installExportPipeline(ctx)

	if err != nil {
		log.Fatal(err)
	}

	defer func() {

		if err := shutdown(ctx); err != nil {
			// TODO: How to handle this error ?
			log.Println(err)
		}
	}()

}

func (psdk *postmanSDK) installExportPipeline(
	ctx context.Context,
) (func(context.Context) error, error) {

	clientHeaders := map[string]string{
		"x-api-key": psdk.Config.ApiKey,
	}
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(
			psdk.Config.ConfigOptions.ReceiverBaseUrl+pmreceiver.BOOTSTRAP_PATH,
		),
		otlptracehttp.WithHeaders(clientHeaders),
	)
	exporter, err := otlptrace.New(ctx, client)

	if err != nil {

		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Println("Could not set resources: ", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithBatchTimeout(
				psdk.Config.ConfigOptions.BufferIntervalInMilliseconds,
			),
		),
		sdktrace.WithResource(resources),
	)
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider.Shutdown, nil
}

// newExporter returns a console exporter.
func newExporter() (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		// stdouttrace.WithWriter(w),
		// Use human-readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
}

// func newResource()
