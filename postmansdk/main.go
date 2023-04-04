package postmansdk

import (
	"context"

	"github.com/gin-gonic/gin"
	instrumentations_gin "github.com/postmanlabs/postmansdk/instrumentations/gin"
	pminterfaces "github.com/postmanlabs/postmansdk/interfaces"
	"github.com/postmanlabs/postmansdk/utils"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var (
	globalCollectionId string
	// This implementation will be replaced by something else that @gmann42 will add to save state.
	ignoreIncomingRequests []string
	log                    *logrus.Entry
)

func Initialize(
	collectionId string,
	apiKey string,
	options ...pminterfaces.PostmanSDKConfigOption,
) func(context.Context) error {

	sdkconfig := pminterfaces.Init(collectionId, apiKey, options...)

	// Setting log level
	if sdkconfig.ConfigOptions.Debug {
		log = utils.CreateNewLogger(logrus.DebugLevel)
	} else {
		log = utils.CreateNewLogger(logrus.ErrorLevel)
	}

	log.WithField("sdkconfig", sdkconfig).Info("SdkConfig is intialized")

	// Check if the sdk should be enabled or not
	if !sdkconfig.ConfigOptions.Enable {
		return func(ctx context.Context) error {
			return nil
		}
	}

	// Adding collectionId to global var
	globalCollectionId = collectionId
	// Remove this from here.
	ignoreIncomingRequests = sdkconfig.ConfigOptions.IgnoreIncomingRequests

	// Adding a stdout exporter
	exporter, err := newExporter()

	if err != nil {
		log.WithError(err).Error("Failed to create a new exporter")
	}
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("library.language", "go"),
			attribute.String(utils.POSTMAN_COLLECTION_ID_ATTRIBUTE_NAME, collectionId),
		),
	)
	if err != nil {
		log.WithError(err).Error("Could not set resources")
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
