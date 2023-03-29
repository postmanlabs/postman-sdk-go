package postmansdk

import (
	"context"
	"log"

	pminterfaces "github.com/postmanlabs/postmansdk/interfaces"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func Initialize(
	apiKey string,
	collectionId string,
	options ...pminterfaces.PostmanSDKConfigOption) func(context.Context) error {

	sdkconfig := pminterfaces.Init(apiKey, collectionId, options...)
	log.Printf("SdkConfig is intialized as %v", sdkconfig)

	// Adding a stdout exporter
	exporter, err := newExporter()

	if err != nil {
		log.Fatal(err)
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

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(sdkconfig.ConfigOptions.BufferIntervalInMilliseconds)),
			sdktrace.WithResource(resources),
		),
	)
	return exporter.Shutdown
}
