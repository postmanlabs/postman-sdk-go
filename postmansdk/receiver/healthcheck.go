package receiver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
	pmutils "github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
)

type healthCheckAPIPayload struct {
	SDK SdkPayload `json:"sdk"`
}

type healthCheckApiResponse struct {
	Healthy bool                     `json:"healthy"`
	Message string                   `json:"message"`
	Data    []map[string]interface{} `json:"data"`
}

type hcResp struct {
	StatusCode int
	Body       healthCheckApiResponse
	Error      error
}

func callHealthApi(sdkconfig *pminterfaces.PostmanSDKConfig) hcResp {
	payload := &healthCheckAPIPayload{
		SDK: SdkPayload{
			CollectionId: sdkconfig.CollectionId,
			Enabled:      sdkconfig.Options.Enable,
		},
	}
	url := sdkconfig.Options.ReceiverBaseUrl + HEALTHCHECK_PATH

	var hr hcResp

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(payload)

	if err != nil {
		hr.Error = fmt.Errorf("error:%v in encoding payload", err)
		return hr
	}

	client := &http.Client{}
	req, reqErr := http.NewRequest("POST", url, b)

	if reqErr != nil {
		hr.Error = fmt.Errorf("error:%v while creating request", reqErr)
		return hr
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", sdkconfig.ApiKey)

	resp, err := client.Do(req)

	if err != nil {
		hr.Error = fmt.Errorf("HTTP call failed: %v", err)
		return hr
	}
	body, berr := decodeBody(resp)

	hr.Error = berr
	hr.Body = body
	pmutils.Log.Debug(fmt.Printf("Healtcheck API %+v", hr))

	return hr
}

func decodeBody(resp *http.Response) (healthCheckApiResponse, error) {
	defer resp.Body.Close()
	var healthResp healthCheckApiResponse
	err := json.NewDecoder(resp.Body).Decode(&healthResp)

	if err != nil {
		return healthResp, fmt.Errorf("parsing resp.Body:%+v failed:%v", resp.Body, err)
	}

	return healthResp, nil

}

func HealthCheck(sdkconfig *pminterfaces.PostmanSDKConfig) {

	for {
		retry := 0

		if retry > HEALTH_CHECK_ERROR_COUNT_THRESHOLD {
			pmutils.Log.Debug("Max retries exceeded disabling Healthcheck")
			return
		}

		resp := callHealthApi(sdkconfig)

		if resp.Error != nil {
			pmutils.Log.Debug("Healthcheck API failed: %v", resp.Error)
			// this will also retry on JSON parsing failed
			// should this be done ?
			continue
		}

		if resp.StatusCode == http.StatusOK {

			if resp.Body.Healthy {
				retry = 0
				sdkconfig.Unsuppress()
			}

			time.Sleep(DEFAULT_HEALTH_PING_INTERVAL_SECONDS * time.Second)
		} else if resp.StatusCode == http.StatusConflict {
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

		} else if resp.StatusCode == http.StatusNotFound {

			if resp.Error == nil && !resp.Body.Healthy {
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

			// pass to channel self.postman_tracer.suppress()
			sdkconfig.Suppress()

			retry += 1
			delay := time.Duration(math.Pow(EXPONENTIAL_BACKOFF_BASE, float64(retry)))
			time.Sleep(delay * time.Second)
		}
	}

}
