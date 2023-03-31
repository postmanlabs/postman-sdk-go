package receiver

import (
	pminterfaces "github.com/postmanlabs/postmansdk/interfaces"
)

type healthCheckAPIPaylod struct {
	SDK SdkPayload `json:"sdk"`
}

type healthCheckApIResponse struct {
	Healthy       bool   `json:"healthy"`
	Message       string `json:"message"`
	CurrentConfig struct {
		Enabled bool `json:"enabled"`
	}
}

func CallHealthCheckAPI(config pminterfaces.PostmanSDKConfig) (bool, error) {
	return false, nil
}
