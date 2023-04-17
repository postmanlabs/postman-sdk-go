package instrumentations_gin

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
	pmutils "github.com/postmanlabs/postman-go-sdk/postmansdk/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func getRequestHeaders(c *gin.Context) string {
	return jsonStringify(arrayToValue(c.Request.Header))
}

func getRequestParams(c *gin.Context) string {
	if c.Params == nil {
		return jsonStringify(map[string]string{})
	}

	// Convert from [{Key: "<key>", Value: "<value>"}] -> {<key>: <value>}
	paramObject := make(map[string]string, len(c.Params))

	for _, p := range c.Params {
		paramObject[p.Key] = p.Value
	}

	return jsonStringify(paramObject)
}

func getRequestQuery(c *gin.Context) string {
	return jsonStringify(arrayToValue(c.Request.URL.Query()))
}

func getRequestBody(c *gin.Context) string {
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		pmutils.Log.WithError(err).Error("Failed to read request body data")
	}

	// Request Body is an io.ReadCloser stream, which can be read only once, since
	// we are reading the same here, the actual gin.Handler will not have the request
	// body available, and so we are adding the same back.
	//
	// io.NopCloser is used for functions that require io.ReadCloser but our current object
	// (bytes.Buffer) doesn't provide a Close function.
	c.Request.Body = io.NopCloser(bytes.NewBuffer(data))

	if string(data) == "" {
		return jsonStringify(map[string]string{})
	}

	return string(data)
}

func getResponseHeaders(c *gin.Context) string {
	return jsonStringify(arrayToValue(c.Writer.Header()))
}

func getResponseBody(c *gin.Context) *bodyLogWriter {
	// For response body
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw

	return blw
}

// Fetches the current span from the request context and adds the missing attributes for request and response data.
func Middleware(sdkconfig *pminterfaces.PostmanSDKConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentSpan := trace.SpanFromContext(c.Request.Context())

		// isRecording will handle the case of:
		//	- currentSpan being a noopSpan -> As a noopSpan's isRecording will always return false.
		//	- span is closed for writing.
		if !currentSpan.IsRecording() {
			// Call the next middleware, and bailout
			c.Next()

			return
		}

		reqBody := getRequestBody(c)
		// TLDR: We need to call this before `c.Next()` as we are overriding the `c.Writer` method in this call.
		//
		// Response body is written as a stream and is not stored anywhere, making it hard to capture.
		// We are overrding the c.Writer method with our bodyLogWriter. This keeps the ResponseWriter as it is, but
		// adds a new bytes.Buffer reference.
		// When a call is made to c.ResponseWriter.Write() method, we perform the following:
		// 			w.body.Write(b)
		// 			return w.ResponseWriter.Write(b)
		//
		// This allows us to capture the response body while it is being written to the stream.
		blw := getResponseBody(c)

		// Call the next middleware.
		// This has to be called before accessing any response attributes, as they are only available after this call
		// completes.
		c.Next()

		currentSpan.SetAttributes(
			attribute.String("http.request.headers", getRequestHeaders(c)),
			attribute.String("http.request.params", getRequestParams(c)),
			attribute.String("http.request.query", getRequestQuery(c)),
			attribute.String("http.request.body", reqBody),
			attribute.String("http.response.headers", getResponseHeaders(c)),
			attribute.String("http.response.body", blw.body.String()),
			attribute.Bool(pmutils.POSTMAN_DATA_TRUNCATION_ATTRIBUTE_NAME, sdkconfig.Options.TruncateData),
		)
	}
}
