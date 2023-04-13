package receiver

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
	pmutils "github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
)

type hRequestBody struct {
	SDK SdkPayload `json:"sdk"`
}

type hResponseBody struct {
	Healthy bool                     `json:"healthy"`
	Message string                   `json:"message"`
	Data    []map[string]interface{} `json:"data"`
}

type healthcheckAPIResponse struct {
	ar   apiResponse
	Body hResponseBody
}

func callHealthApi(sdkconfig *pminterfaces.PostmanSDKConfig) healthcheckAPIResponse {
	payload := hRequestBody{
		SDK: SdkPayload{
			CollectionId: sdkconfig.CollectionId,
			Enabled:      sdkconfig.Options.Enable,
		},
	}

	resp := callApi(HEALTHCHECK_PATH, payload, sdkconfig)

	defer resp.Body.Close()

	var hr healthcheckAPIResponse
	var hbody hResponseBody

	err := json.NewDecoder(resp.Body).Decode(&hbody)

	if err != nil {
		hr.ar.DecodeError = fmt.Errorf("parsing resp.Body:%+v failed:%v", hr.ar.Body, err)
	}

	hr.Body = hbody
	pmutils.Log.Debug(fmt.Printf("Healtcheck API %+v", hr))

	return hr
}

func HealthCheck(sdkconfig *pminterfaces.PostmanSDKConfig) {

	for {
		retry := 0

		if retry > HEALTH_CHECK_ERROR_COUNT_THRESHOLD {
			pmutils.Log.Debug("Max retries exceeded disabling Healthcheck")
			return
		}

		resp := callHealthApi(sdkconfig)

		if resp.ar.Error != nil {
			pmutils.Log.Debug("Healthcheck API failed: %v", resp.ar.Error)

			sdkconfig.Suppress()
			retry += 1
			delay := time.Duration(math.Pow(EXPONENTIAL_BACKOFF_BASE, float64(retry)))
			time.Sleep(delay * time.Second)

			continue
		}

		if resp.ar.StatusCode == http.StatusOK {

			if resp.ar.DecodeError == nil && resp.Body.Healthy {
				retry = 0
				sdkconfig.Unsuppress()
			}

			time.Sleep(DEFAULT_HEALTH_PING_INTERVAL_SECONDS * time.Second)
		} else if resp.ar.StatusCode == http.StatusConflict {
			br, err := Bootstrap(sdkconfig)

			if err != nil {
				sdkconfig.Suppress()
				pmutils.Log.Debug("Shutting down healthcheck")
				return
			}
			if !br {
				sdkconfig.Suppress()
			} else {
				sdkconfig.Unsuppress()
			}

			time.Sleep(DEFAULT_HEALTH_PING_INTERVAL_SECONDS * time.Second)

		} else if resp.ar.StatusCode == http.StatusNotFound {

			if resp.ar.DecodeError == nil && !resp.Body.Healthy {
				br, err := Bootstrap(sdkconfig)
				if err != nil {
					sdkconfig.Suppress()
					pmutils.Log.Debug("Shutting down healthcheck")
					return
				}
				if !br {
					sdkconfig.Suppress()
				} else {
					sdkconfig.Unsuppress()
				}
				time.Sleep(DEFAULT_HEALTH_PING_INTERVAL_SECONDS * time.Second)

			} else {
				//Case when url itself is wrong
				sdkconfig.Suppress()
				pmutils.Log.Debug("Shutting down healthcheck")

				return
			}

		} else {
			sdkconfig.Suppress()
			retry += 1
			exponentialDelay(retry)
		}
	}

}
