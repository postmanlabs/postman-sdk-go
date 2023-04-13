package receiver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
)

type SdkPayload struct {
	CollectionId string `json:"collectionId"`
	Enabled      bool   `json:"enabled"`
}

type apiResponse struct {
	StatusCode  int
	Body        io.ReadCloser
	Error       error
	DecodeError error
}

func callApi(urlPath string, payload interface{}, sdkconfig *pminterfaces.PostmanSDKConfig) apiResponse {

	var ar apiResponse
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(payload)

	if err != nil {
		ar.Error = fmt.Errorf("error in json encoding %v", err)
		return ar
	}

	url := sdkconfig.Options.ReceiverBaseUrl + urlPath
	client := &http.Client{}
	req, reqErr := http.NewRequest("POST", url, b)

	if reqErr != nil {
		ar.Error = fmt.Errorf("error:%v while creating request", reqErr)
		return ar
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(X_API_KEY, sdkconfig.ApiKey)

	resp, err := client.Do(req)

	if err != nil {
		ar.Error = fmt.Errorf("HTTP call failed:%v", err)
		return ar
	}

	ar.Body = resp.Body
	ar.StatusCode = resp.StatusCode

	return ar
}

func exponentialDelay(factor int) {
	delay := time.Duration(math.Pow(EXPONENTIAL_BACKOFF_BASE, float64(factor)))
	time.Sleep(delay * time.Second)
}
