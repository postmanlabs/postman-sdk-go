// Deprecated: This package is DEPRECATED! It is no longer maintained.
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

	pmexporter "github.com/postmanlabs/postman-sdk-go/postmansdk/exporter"
	pminterfaces "github.com/postmanlabs/postman-sdk-go/postmansdk/interfaces"
	pmreceiver "github.com/postmanlabs/postman-sdk-go/postmansdk/receiver"
	pmutils "github.com/postmanlabs/postman-sdk-go/postmansdk/utils"
)

type postmanSDK struct {
	Config       *pminterfaces.PostmanSDKConfig
	Integrations integrations
}

var psdk postmanSDK

func Initialize(
	collectionId string,
	apiKey string,
	options ...pminterfaces.PostmanSDKConfigOption,
) (postmanSDK, error) {

	sdkconfig := pminterfaces.InitializeSDKConfig(collectionId, apiKey, options...)
	psdk = postmanSDK{
		Config: sdkconfig,
	}

	if sdkconfig.Options.Debug {
		pmutils.CreateNewLogger(logrus.DebugLevel)
	} else {
		pmutils.CreateNewLogger(logrus.ErrorLevel)
	}

	if !sdkconfig.Options.Enable {
		pmutils.Log.Error("Postman SDK is not enabled")
		return psdk, fmt.Errorf("postman SDK is not enabled")
	}

	// Register live collection
	if err := pmreceiver.UpdateConfig(sdkconfig); err != nil {
		pmutils.Log.WithError(err).Error("Postman SDK disabled due to bootstrap failure")
		return psdk, err
	}

	ctx := context.Background()

	shutdown, err := psdk.installExportPipeline(ctx)

	if err != nil {
		pmutils.Log.WithError(err).Error("Failed to create a new exporter")
		defer shutdown(ctx)
		return psdk, err
	}

	go pmreceiver.HealthCheck(psdk.Config)

	pmutils.Log.Info("Postman SDK initialized")
	return psdk, nil
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

func (psdk *postmanSDK) getExporter(ctx context.Context) (*pmexporter.PostmanExporter, error) {
	exporter, err := psdk.getOTLPExporter(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	pexporter := &pmexporter.PostmanExporter{
		Exporter:  *exporter,
		Sdkconfig: psdk.Config,
	}
	return pexporter, nil
}

func getResource(ctx context.Context) (*resource.Resource, error) {
	return resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String("telemetry.sdk.language", "go"),
			attribute.String(
				pmutils.POSTMAN_COLLECTION_ID_ATTRIBUTE_NAME, psdk.Config.CollectionId,
			),
			attribute.String(pmutils.POSTMAN_SDK_VERSION_ATTRIBUTE_NAME, POSTMAN_SDK_VERSION),
		),
	)
}

func (psdk *postmanSDK) installExportPipeline(
	ctx context.Context,
) (func(context.Context) error, error) {

	exporter, err := psdk.getExporter(ctx)
	if err != nil {
		return nil, err
	}

	resources, err := getResource(ctx)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithBatchTimeout(
				psdk.Config.Options.BufferIntervalInMilliseconds,
			),
		),
		sdktrace.WithResource(resources),
	)
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider.Shutdown, nil
}
