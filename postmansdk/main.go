package postmansdk

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/sirupsen/logrus"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	pmexporter "github.com/postmanlabs/postman-go-sdk/postmansdk/exporter"
	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
	pmreceiver "github.com/postmanlabs/postman-go-sdk/postmansdk/receiver"
	pmutils "github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
)

type postmanSDK struct {
	Config *pminterfaces.PostmanSDKConfig
}

var psdk *postmanSDK

func Initialize(
	collectionId string,
	apiKey string,
	options ...pminterfaces.PostmanSDKConfigOption,
) {

	sdkconfig := pminterfaces.InitializeSDKConfig(collectionId, apiKey, options...)

	if !sdkconfig.Options.Enable {
		log.Println(fmt.Errorf("postman SDK is not enabled"))
		return
	}

	if sdkconfig.Options.Debug {
		pmutils.CreateNewLogger(logrus.DebugLevel)
	} else {
		pmutils.CreateNewLogger(logrus.ErrorLevel)
	}

	pmutils.Log.WithField("sdkconfig", sdkconfig).Info("SdkConfig is intialized")

	psdk = &postmanSDK{
		Config: sdkconfig,
	}

	// Register live collection
	if err := pmreceiver.UpdateConfig(sdkconfig); err != nil {
		pmutils.Log.Error(err)
		return
	}

	ctx := context.Background()

	shutdown, err := psdk.installExportPipeline(ctx)

	if err != nil {
		pmutils.Log.WithError(err).Error("Failed to create a new exporter")
		defer shutdown(context.Background())
		return
	}

	go pmreceiver.HealthCheck(psdk.Config)

	psdk.registerInstrumentations()

}

func (psdk *postmanSDK) getOTLPExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	clientHeaders := map[string]string{
		"x-api-key": psdk.Config.ApiKey,
	}
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(
			strings.Replace(
				psdk.Config.Options.ReceiverBaseUrl,
				"https://",
				"",
				1,
			),
		),
		otlptracehttp.WithURLPath(pmreceiver.TRACE_RECEIVER_PATH),
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
		Exporter:  *exporter,
		Sdkconfig: psdk.Config,
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("telemetry.sdk.language", "go"),
			attribute.String(
				pmutils.POSTMAN_COLLECTION_ID_ATTRIBUTE_NAME, psdk.Config.CollectionId,
			),
			attribute.String(pmutils.POSTMAN_SDK_VERSION_ATTRIBUTE_NAME, POSTMAN_SDK_VERSION),
		),
	)
	if err != nil {
		pmutils.Log.WithError(err).Error("Could not set resources")
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			pexporter,
			sdktrace.WithBatchTimeout(
				psdk.Config.Options.BufferIntervalInMilliseconds,
			),
		),
		sdktrace.WithResource(resources),
	)
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider.Shutdown, nil
}

func (psdk *postmanSDK) registerInstrumentations() {
	if psdk.Config.Options.GinInstrumentation == nil {
		return
	}
	InstrumentGin(psdk.Config.Options.GinInstrumentation)

}
