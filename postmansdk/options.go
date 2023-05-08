package postmansdk

import (
	"time"

	pminterfaces "github.com/postmanlabs/postman-sdk-go/postmansdk/interfaces"
)

func WithBufferIntervalInMilliseconds(bufferMillis int) pminterfaces.PostmanSDKConfigOption {
	return func(option *pminterfaces.PostmanSDKConfigOptions) {
		option.BufferIntervalInMilliseconds = time.Duration(bufferMillis) * time.Millisecond
	}
}

func WithDebug(debug bool) pminterfaces.PostmanSDKConfigOption {
	return func(option *pminterfaces.PostmanSDKConfigOptions) {
		option.Debug = debug
	}
}

func WithEnable(enable bool) pminterfaces.PostmanSDKConfigOption {
	return func(option *pminterfaces.PostmanSDKConfigOptions) {
		option.Enable = enable
	}
}
func WithReceiverBaseUrl(receiverBaseUrl string) pminterfaces.PostmanSDKConfigOption {
	return func(option *pminterfaces.PostmanSDKConfigOptions) {
		option.ReceiverBaseUrl = receiverBaseUrl
	}
}
func WithTruncateData(truncateData bool) pminterfaces.PostmanSDKConfigOption {
	return func(option *pminterfaces.PostmanSDKConfigOptions) {
		option.TruncateData = truncateData

	}
}
func WithRedactSensitiveData(enable bool, rules map[string]string) pminterfaces.PostmanSDKConfigOption {
	return func(option *pminterfaces.PostmanSDKConfigOptions) {
		option.RedactSensitiveData = pminterfaces.RedactSensitiveDataConfig{
			Enable: enable,
			Rules:  rules,
		}
	}
}
func WithIgnoreOutgoingRequests(ignoreOutgoingRequests []string) pminterfaces.PostmanSDKConfigOption {
	return func(option *pminterfaces.PostmanSDKConfigOptions) {
		option.IgnoreOutgoingRequests = ignoreOutgoingRequests
	}
}
func WithIgnoreIncomingRequests(ignoreIncomingRequests []string) pminterfaces.PostmanSDKConfigOption {
	return func(option *pminterfaces.PostmanSDKConfigOptions) {
		option.IgnoreIncomingRequests = ignoreIncomingRequests
	}
}
