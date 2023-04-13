package postmansdk

import (
	"context"
	"fmt"
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

func errorCleanup(context.Context) error {
	return nil
}

func Initialize(
	collectionId string,
	apiKey string,
	options ...pminterfaces.PostmanSDKConfigOption,
) (func(context.Context) error, error) {

	sdkconfig := pminterfaces.InitializeSDKConfig(collectionId, apiKey, options...)

	if !sdkconfig.Options.Enable {
		return errorCleanup, fmt.Errorf("postman SDK is not enabled")
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

	if _, err := psdk.updateConfig(); err != nil {
		return errorCleanup, err
	}

	ctx := context.Background()

	shutdown, err := psdk.installExportPipeline(ctx)

	if err != nil {
		pmutils.Log.WithError(err).Error("Failed to create a new exporter")

		return errorCleanup, err
	}

	go pmreceiver.HealthCheck(psdk.Config)

	return shutdown, nil
}

func (psdk *postmanSDK) updateConfig() (bool, error) {
	enable, err := pmreceiver.Bootstrap(psdk.Config)

	if err != nil {
		psdk.Config.Suppress()
		pmutils.Log.WithField("error", err).Error("SDK disabled due to bootstrap failure")
		return false, err
	}

	if !enable {
		psdk.Config.Suppress()
	} else {
		psdk.Config.Unsuppress()
	}

	return true, nil
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
		Exporter: *exporter,
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("library.language", "go"),
			attribute.String(
				pmutils.POSTMAN_COLLECTION_ID_ATTRIBUTE_NAME, psdk.Config.CollectionId,
			),
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
