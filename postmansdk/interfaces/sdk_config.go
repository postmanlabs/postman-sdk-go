package interfaces

import (
	"time"
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
	RedactSensitiveData          map[string]interface{}
	IgnoreOutgoingRequests       []string
	IgnoreIncomingRequests       []string
}

type PostmanSDKConfig struct {
	ApiKey       string
	CollectionId string
	Options      PostmanSDKConfigOptions
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
func WithRedactSensitiveData(redactSensitiveData map[string]interface{}) PostmanSDKConfigOption {
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
