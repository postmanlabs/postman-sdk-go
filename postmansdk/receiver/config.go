package receiver

const (
	BOOTSTRAP_PATH   = "/sdk/bootstrap"
	ContentType      = "application/json"
	DEFAULT_BASE_URL = "https://trace-receiver.postman.com"
)

type ReceiverResponse struct {
	Body   map[string]interface{}
	Status int
}
