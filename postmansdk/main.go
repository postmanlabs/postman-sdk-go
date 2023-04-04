package postmansdk

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	instrumentations_gin "github.com/postmanlabs/postman-go-sdk/postmansdk/instrumentations/gin"
	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
	"github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var globalCollectionId string

func Initialize(
	collectionId string,
	apiKey string,
	options ...pminterfaces.PostmanSDKConfigOption,
) func(context.Context) error {

	sdkconfig := pminterfaces.Init(collectionId, apiKey, options...)
	log.Printf("SdkConfig is intialized as %v", sdkconfig)

	// Adding collectionId to global var
	globalCollectionId = collectionId

	PrintVersion()

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
	router.Use(otelgin.Middleware(globalCollectionId))
	router.Use(instrumentations_gin.Middleware())
}
