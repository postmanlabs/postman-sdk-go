package postmansdk

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	instrumentations_gin "github.com/postmanlabs/postmansdk/instrumentations/gin"
	pminterfaces "github.com/postmanlabs/postmansdk/interfaces"
	"github.com/postmanlabs/postmansdk/utils"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var globalCollectionId string

// This implementation will be replaced by something else that @gmann42 will add to save state.
var ignoreIncomingRequests []string

func Initialize(
	collectionId string,
	apiKey string,
	options ...pminterfaces.PostmanSDKConfigOption,
) func(context.Context) error {

	sdkconfig := pminterfaces.Init(collectionId, apiKey, options...)
	log.Printf("SdkConfig is intialized as %v", sdkconfig)

	// Adding collectionId to global var
	globalCollectionId = collectionId
	// Remove this from here.
	ignoreIncomingRequests = sdkconfig.ConfigOptions.IgnoreIncomingRequests

	// Adding a stdout exporter
	exporter, err := newExporter()

	if err != nil {
		log.Fatal(err)
	}
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("library.language", "go"),
			attribute.String(utils.POSTMAN_COLLECTION_ID_ATTRIBUTE_NAME, collectionId),
		),
	)
	if err != nil {
		log.Println("Could not set resources: ", err)
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(sdkconfig.ConfigOptions.BufferIntervalInMilliseconds)),
			sdktrace.WithResource(resources),
		),
	)

	return exporter.Shutdown
}

func InstrumentGin(router *gin.Engine) {
	router.Use(otelgin.Middleware(globalCollectionId, getMiddlewareOptions()...))
	router.Use(instrumentations_gin.Middleware())
}
