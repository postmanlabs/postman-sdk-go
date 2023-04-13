package receiver

const BOOTSTRAP_PATH = "/sdk/bootstrap"
const TRACE_RECEIVER_PATH = "/traces"
const HEALTHCHECK_PATH = "/sdk/health"
const X_API_KEY = "x-api-key"

const BOOTSTRAP_RETRY_COUNT = 2
const EXPONENTIAL_BACKOFF_BASE = 2

const DEFAULT_HEALTH_PING_INTERVAL_SECONDS = 60
const HEALTH_CHECK_ERROR_COUNT_THRESHOLD = 5
