package instrumentations_gin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func jsonStringify(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		fmt.Println(err)
	}

	return string(b)
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func getRequestHeaders(c *gin.Context) string {
	return jsonStringify(c.Request.Header)
}

func getRequestParams(c *gin.Context) string {
	return jsonStringify(c.Params)
}

func getRequestQuery(c *gin.Context) string {
	return jsonStringify(c.Request.URL.Query())
}

func getRequestBody(c *gin.Context) string {
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		log.Println(err)
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(data))

	return string(data)
}

func getResponseHeaders(c *gin.Context) string {
	return jsonStringify(c.Writer.Header())
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
