package receiver

import (
	"encoding/json"
	"fmt"
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

type healthcheckApiResponse struct {
	ar   apiResponse
	Body hResponseBody
}

func callHealthApi(sdkconfig *pminterfaces.PostmanSDKConfig) healthcheckApiResponse {
	payload := hRequestBody{
		SDK: SdkPayload{
			CollectionId: sdkconfig.CollectionId,
			Enabled:      sdkconfig.Options.Enable,
		},
	}

	resp := callApi(HEALTHCHECK_PATH, payload, sdkconfig)

	defer resp.Body.Close()

	var hr healthcheckApiResponse
	var hbody hResponseBody
	hr.ar = resp

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
			sdkconfig.Suppress()
			pmutils.Log.Debug("Max retries exceeded disabling Healthcheck")

			return
		}

		resp := callHealthApi(sdkconfig)

		if resp.ar.Error != nil {
			pmutils.Log.Debug("Healthcheck API failed: %v", resp.ar.Error)

			sdkconfig.Suppress()
			retry += 1
			exponentialDelay(retry)

		} else if resp.ar.StatusCode == http.StatusOK {

			if resp.ar.DecodeError == nil && resp.Body.Healthy {
				retry = 0
				sdkconfig.Unsuppress()
			}

			time.Sleep(DEFAULT_HEALTH_PING_INTERVAL_SECONDS * time.Second)

		} else if resp.ar.StatusCode == http.StatusConflict {

			if err := UpdateConfig(sdkconfig); err != nil {
				pmutils.Log.Debug("Shutting down healthcheck")

				return
			}

			time.Sleep(DEFAULT_HEALTH_PING_INTERVAL_SECONDS * time.Second)

		} else if resp.ar.StatusCode == http.StatusNotFound {
			// Healthcheck received without bootstrapping
			if resp.ar.DecodeError == nil && !resp.Body.Healthy {

				if err := UpdateConfig(sdkconfig); err != nil {
					pmutils.Log.Debug("Shutting down healthcheck")

					return
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
			pmutils.Log.Debug(fmt.Printf("Retrying healthcheck %d", retry))
			exponentialDelay(retry)
		}
	}

}

func UpdateConfig(pc *pminterfaces.PostmanSDKConfig) error {
	enable, err := Bootstrap(pc)

	if err != nil {
		pc.Suppress()
		pmutils.Log.WithField("error", err).Error("SDK disabled due to bootstrap failure")
		return err
	}

	if !enable {
		pc.Suppress()
	} else {
		pc.Unsuppress()
	}

	return nil
}
