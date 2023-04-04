package postmansdk

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/postmanlabs/postmansdk/utils"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	pminterfaces "github.com/postmanlabs/postmansdk/interfaces"
	instrumentations_gin "github.com/postmanlabs/postmansdk/instrumentations/gin"
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


	// TODO: Should this be passed from request handler ?
	ctx := context.Background()

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


	exporter, err := newExporter() 

	if err != nil {

		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("library.language", "go"),
			attribute.String(utils.POSTMAN_COLLECTION_ID_ATTRIBUTE_NAME, psdk.Config.CollectionId),
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

func InstrumentGin(router *gin.Engine) {
	router.Use(otelgin.Middleware(""))
	router.Use(instrumentations_gin.Middleware())
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
