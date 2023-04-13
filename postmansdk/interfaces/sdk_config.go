package interfaces

import (
	"time"

	"github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
)

const (
	DefaultBufferIntervalInMilliseconds = 5000
	DefaultDebug                        = false
	DefaultEnable                       = true
	DefaultReceiverBaseUrl              = "https://trace-receiver.postman.com"
	DefaultTruncateData                 = true
)

type PostmanSDKConfigOption func(option *PostmanSDKConfigOptions)

type PostmanSDKConfigOptions struct {
	BufferIntervalInMilliseconds time.Duration
	Debug                        bool
	Enable                       bool
	ReceiverBaseUrl              string
	TruncateData                 bool
	RedactSensitiveData          RedactSensitiveDataConfig
	IgnoreOutgoingRequests       []string
	IgnoreIncomingRequests       []string
}

type PostmanSDKConfig struct {
	ApiKey       string
	CollectionId string
	Options      PostmanSDKConfigOptions
}

type RedactSensitiveDataConfig struct {
	RedactionEnable bool
	Rules           map[string]string
}

func InitializeSDKConfig(collectionId string, apiKey string, options ...PostmanSDKConfigOption) PostmanSDKConfig {

	o := PostmanSDKConfigOptions{
		BufferIntervalInMilliseconds: DefaultBufferIntervalInMilliseconds * time.Millisecond,
		Debug:                        DefaultDebug,
		Enable:                       DefaultEnable,
		ReceiverBaseUrl:              DefaultReceiverBaseUrl,
		TruncateData:                 DefaultTruncateData,
	}
	for _, opt := range options {
		opt(&o)
	}
	sdkconfig := &PostmanSDKConfig{
		ApiKey:       apiKey,
		CollectionId: collectionId,
		Options:      o,
	}

	// Add a check here for the env config to start/stop the SDK.
	v, err := utils.GetenvBool(utils.POSTMAN_SDK_ENABLE_ENV_VAR_NAME)
	if err == nil {
		sdkconfig.Options.Enable = v
	}

	return *sdkconfig
}

func WithBufferIntervalInMilliseconds(bufferMillis int) PostmanSDKConfigOption {
	return func(option *PostmanSDKConfigOptions) {
		option.BufferIntervalInMilliseconds = time.Duration(bufferMillis) * time.Millisecond
	}
}
func WithDebug(debug bool) PostmanSDKConfigOption {
	return func(option *PostmanSDKConfigOptions) {
		option.Debug = debug
	}
}

func WithEnable(enable bool) PostmanSDKConfigOption {
	return func(option *PostmanSDKConfigOptions) {
		option.Enable = enable
	}
}
func WithReceiverBaseUrl(receiverBaseUrl string) PostmanSDKConfigOption {
	return func(option *PostmanSDKConfigOptions) {
		option.ReceiverBaseUrl = receiverBaseUrl
	}
}
func WithTruncateData(truncateData bool) PostmanSDKConfigOption {
	return func(option *PostmanSDKConfigOptions) {
		option.TruncateData = truncateData

	}
}
func WithRedactSensitiveData(redactSensitiveData RedactSensitiveDataConfig) PostmanSDKConfigOption {
	return func(option *PostmanSDKConfigOptions) {
		option.RedactSensitiveData = redactSensitiveData
	}
}
func WithIgnoreOutgoingRequests(ignoreOutgoingRequests []string) PostmanSDKConfigOption {
	return func(option *PostmanSDKConfigOptions) {
		option.IgnoreOutgoingRequests = ignoreOutgoingRequests
	}
}
func WithIgnoreIncomingRequests(ignoreIncomingRequests []string) PostmanSDKConfigOption {
	return func(option *PostmanSDKConfigOptions) {
		option.IgnoreIncomingRequests = ignoreIncomingRequests
	}
}
