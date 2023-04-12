package receiver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
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
			Enabled:      sdkconfig.Options.Enable,
		},
	}
	url := sdkconfig.Options.ReceiverBaseUrl + BOOTSTRAP_PATH

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(payload)

	if err != nil {
		return false, fmt.Errorf("error in json encoding %v", err)
	}
	log.Printf("bootstrap API payload:%v", b)

	client := &http.Client{}
	req, reqErr := http.NewRequest("POST", url, b)

	if reqErr != nil {
		return false, fmt.Errorf("error:%v while creating request", reqErr)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", sdkconfig.ApiKey)

	for retries := 0; retries < BOOTSTRAP_RETRY_COUNT; retries++ {
		resp, err := client.Do(req)

		if err != nil {

			return false, fmt.Errorf("error making request:%v", err)
		}

		defer resp.Body.Close()

		log.Printf("bootstrap API resp.status:%d", resp.StatusCode)

		var bootResp bootStrapApIResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&bootResp)

		if decodeErr != nil {
			decodeErr = fmt.Errorf(
				"bootstrap API resp.status:%d resp.Body:%+v parsing failed with error:%v",
				resp.StatusCode,
				resp.Body,
				decodeErr,
			)
		}else {
			log.Printf("bootstrap API resp.Body: %+v", bootResp)
		}
		if resp.StatusCode == http.StatusOK {

			if decodeErr != nil {
				return false, decodeErr
			}

			if !bootResp.OK {

				return false, fmt.Errorf(
					"non OK bootstrap API resp.status:%v resp.body: %+v",
					resp.StatusCode,
					bootResp,
				)
			}

			return bootResp.CurrentConfig.Enabled, nil

		} else if isRetryable(resp.StatusCode) {
			log.Printf(
				"Retry:%d bootstrap API received resp.status:%d",
				retries,
				resp.StatusCode,
			)

			delay := time.Duration(math.Pow(EXPONENTIAL_BACKOFF_BASE, float64(retries)))

			time.Sleep(delay * time.Second)
			continue
		} else {
			if decodeErr != nil {
				return false,
					fmt.Errorf("bootstrap failed resp.status:%d", resp.StatusCode)
			}
			return false,
				fmt.Errorf(
					"bootstrap failed resp.status:%d, resp.Body:%+v",
					resp.StatusCode,
					bootResp,
				)
		}

	}

	return false, fmt.Errorf("bootstrap API max retries exceeded")

}
