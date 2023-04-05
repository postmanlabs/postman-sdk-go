package receiver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	pminterfaces "github.com/postmanlabs/postmansdk/interfaces"
)

type SdkPayload struct {
	CollectionId string `json:"collectionId"`
	Enabled      bool   `json:"enabled"`
}

type bootStrapAPIPaylod struct {
	SDK SdkPayload `json:"sdk"`
}

type bootStrapApIResponse struct {
	OK            bool   `json:"ok"`
	Message       string `json:"message"`
	CurrentConfig struct {
		Enabled bool `json:"enabled"`
	}
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

func CallBootStrapAPI(sdkconfig pminterfaces.PostmanSDKConfig) (bool, error) {
	payload := &bootStrapAPIPaylod{
		SDK: SdkPayload{
			CollectionId: sdkconfig.CollectionId,
			Enabled:      sdkconfig.ConfigOptions.Enable,
		},
	}
	url := sdkconfig.ConfigOptions.ReceiverBaseUrl + BOOTSTRAP_PATH

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(payload)

	if err != nil {
		log.Printf("Error in json encoding %v", err)
	}
	log.Printf("Bootstrap API payload:%v", b)

	client := &http.Client{}
	req, reqErr := http.NewRequest("POST", url, b)

	if reqErr != nil {
		return false, fmt.Errorf("error:%v while creating request", reqErr)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", sdkconfig.ApiKey)

	for retries := 0; retries < 2; retries++ {
		resp, err := client.Do(req)

		if err != nil {

			return false, fmt.Errorf("error making request:%v", err)
		}

		defer resp.Body.Close()

		if isRetryable(resp.StatusCode) {
			log.Printf(
				"Retry:%d bootstrap API received resp.status:%d",
				retries,
				resp.StatusCode,
			)
			time.Sleep(1 * time.Second)
			continue
		} else if resp.StatusCode == http.StatusOK {

			var bootResp bootStrapApIResponse
			decodeErr := json.NewDecoder(resp.Body).Decode(&bootResp)

			if decodeErr != nil {

				return false, fmt.Errorf(
					"failed to parse bootstrap api resp.status:%v resp.Body:%v into json with error:%v",
					resp.StatusCode,
					resp.Body,
					decodeErr,
				)
			}

			if !bootResp.OK {

				return false, fmt.Errorf(
					"non OK bootstrap API resp.status:%v resp.body: %v",
					resp.StatusCode,
					bootResp,
				)
			}

			log.Printf(
				"Bootstrap API called resp.status:%v, resp.body:%+v",
				resp.StatusCode, bootResp,
			)

			// If CurrentConfig wasn't returned by bootstrap, it'll be set
			// as false by default due to golang.
			// TODO: Throw error if currentConfig not present ?
			return bootResp.CurrentConfig.Enabled, nil

		} else {

			return false,
				fmt.Errorf("bootstrap resp.status:%d", resp.StatusCode)
		}

	}

	return false, fmt.Errorf("bootstrap API max retries exceeded")

}
