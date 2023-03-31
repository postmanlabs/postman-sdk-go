package instrumentations_gin

import (
	"bytes"
	"io"
	"log"

	"github.com/gin-gonic/gin"
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
		log.Println(err)
	}

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

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentSpan := trace.SpanFromContext(c.Request.Context())

		if !currentSpan.IsRecording() {
			// Call the next middleware, and bailout
			c.Next()

			return
		}

		reqBody := getRequestBody(c)
		blw := getResponseBody(c)

		// Call the next middleware
		c.Next()

		currentSpan.SetAttributes(
			attribute.String("http.request.headers", getRequestHeaders(c)),
			attribute.String("http.request.params", getRequestParams(c)),
			attribute.String("http.request.query", getRequestQuery(c)),
			attribute.String("http.request.body", reqBody),
			attribute.String("http.response.headers", getResponseHeaders(c)),
			attribute.String("http.response.body", blw.body.String()),
		)
	}
}
