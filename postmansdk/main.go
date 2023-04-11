package postmansdk

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	pmexporter "github.com/postmanlabs/postman-go-sdk/postmansdk/exporter"
	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
	pmutils "github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
)

type postmanSDK struct {
	Config pminterfaces.PostmanSDKConfig
}

// This implementation will be replaced by something else that @gmann42 will add to save state.
var ignoreIncomingRequests []string

func Initialize(
	collectionId string,
	apiKey string,
	options ...pminterfaces.PostmanSDKConfigOption,
) func(context.Context) error {

	sdkconfig := pminterfaces.InitializeSDKConfig(collectionId, apiKey, options...)
	log.Printf("SdkConfig is intialized as %+v", sdkconfig)

	psdk := &postmanSDK{
		Config: sdkconfig,
	}
	// Remove this from here.
	ignoreIncomingRequests = sdkconfig.ConfigOptions.IgnoreIncomingRequests

	ctx := context.Background()

	shutdown, err := psdk.installExportPipeline(ctx)

	if err != nil {
		log.Fatal(err)
	}
	return shutdown

}

func (psdk *postmanSDK) getOTLPExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	clientHeaders := map[string]string{
		"x-api-key": psdk.Config.ApiKey,
	}
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(
			strings.Replace(
				psdk.Config.ConfigOptions.ReceiverBaseUrl,
				"https://",
				"",
				1,
			),
		),
		otlptracehttp.WithURLPath(pmutils.TRACE_RECEIVER_PATH),
		otlptracehttp.WithHeaders(clientHeaders),
	)
	exporter, err := otlptrace.New(ctx, client)

	return exporter, err
}

func (psdk *postmanSDK) installExportPipeline(
	ctx context.Context,
) (func(context.Context) error, error) {

	exporter, err := psdk.getOTLPExporter(ctx)

	if err != nil {

		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	pexporter := &pmexporter.PostmanExporter{
		Exporter: *exporter,
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("library.language", "go"),
			attribute.String(
				pmutils.POSTMAN_COLLECTION_ID_ATTRIBUTE_NAME,
				psdk.Config.CollectionId,
			),
		),
	)
	if err != nil {
		log.Println("Could not set resources: ", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			pexporter,
			sdktrace.WithBatchTimeout(
				psdk.Config.ConfigOptions.BufferIntervalInMilliseconds,
			),
		),
		sdktrace.WithResource(resources),
	)
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider.Shutdown, nil
}
