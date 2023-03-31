package instrumentations_gin

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
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

func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)

	return w.ResponseWriter.WriteString(s)
}

func arrayToValue(h map[string][]string) map[string]string {
	newMap := make(map[string]string, len(h))

	for k, v := range h {
		newMap[k] = v[0]
	}

	return newMap
}
