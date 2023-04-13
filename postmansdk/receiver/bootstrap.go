package receiver

import (
	"encoding/json"
	"fmt"
	"net/http"

	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
	pmutils "github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
)

type bRequestBody struct {
	SDK SdkPayload `json:"sdk"`
}

type bResponseBody struct {
	OK            bool   `json:"ok"`
	Message       string `json:"message"`
	CurrentConfig struct {
		Enabled bool `json:"enabled"`
	}
}

type bootstrapAPIResponse struct {
	ar   apiResponse
	Body bResponseBody
}

func isRetryable(statusCode int) bool {
	retryCodes := []int{
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	}
	for _, value := range retryCodes {
		if statusCode == value {
			return true
		}
	}
	return false
}

func callBootstrapApi(sdkconfig *pminterfaces.PostmanSDKConfig) bootstrapAPIResponse {
	payload := bRequestBody{
		SDK: SdkPayload{
			CollectionId: sdkconfig.CollectionId,
			Enabled:      sdkconfig.Options.Enable,
		},
	}

	resp := callApi(BOOTSTRAP_PATH, payload, sdkconfig)
	defer resp.Body.Close()

	var br bootstrapAPIResponse
	var body bResponseBody
	br.ar = resp

	err := json.NewDecoder(resp.Body).Decode(&body)

	if err != nil {
		br.ar.DecodeError = fmt.Errorf("parsing resp.Body:%+v failed:%v", resp.Body, err)
	}

	br.Body = body
	pmutils.Log.Debug(fmt.Printf("Bootstrap API %+v", br))

	return br
}

func Bootstrap(sdkconfig *pminterfaces.PostmanSDKConfig) (bool, error) {

	for retries := 0; retries < BOOTSTRAP_RETRY_COUNT; retries++ {

		resp := callBootstrapApi(sdkconfig)
		if resp.ar.Error != nil {
			pmutils.Log.Debug("Bootstrap API Failed resp: %+v", resp)
			exponentialDelay(retries)
		}

		if resp.ar.StatusCode == http.StatusOK {

			if resp.ar.DecodeError != nil && !resp.Body.OK {
				return false, fmt.Errorf(
					"non OK bootstrap API resp.status:%v resp.body: %+v",
					resp.ar.StatusCode,
					resp.Body,
				)
			}

			return resp.Body.CurrentConfig.Enabled, nil

		} else if isRetryable(resp.ar.StatusCode) {
			pmutils.Log.Debug(
				"Retry:%d bootstrap API received resp.status:%d",
				retries,
				resp.ar.StatusCode,
			)
			exponentialDelay(retries)

		} else {
			return false, fmt.Errorf("unhandled status code bootstrap API resp:%+v", resp)
		}

	}

	return false, fmt.Errorf("bootstrap API max retries exceeded")

}
