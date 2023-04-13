package receiver

const (
	BOOTSTRAP_PATH                       = "/sdk/bootstrap"
	TRACE_RECEIVER_PATH                  = "/traces"
	HEALTHCHECK_PATH                     = "/sdk/health"
	X_API_KEY                            = "x-api-key"
	BOOTSTRAP_RETRY_COUNT                = 2
	BOOTSTRAP_RETRY_DELAY_SECONDS        = 1
	EXPONENTIAL_BACKOFF_BASE             = 2
	DEFAULT_HEALTH_PING_INTERVAL_SECONDS = 60
	HEALTH_CHECK_ERROR_COUNT_THRESHOLD   = 5
)
