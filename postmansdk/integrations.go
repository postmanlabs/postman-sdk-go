package postmansdk

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	instrumentations_gin "github.com/postmanlabs/postman-go-sdk/postmansdk/instrumentations/gin"
	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
)

type integrations struct{}

func (i *integrations) Gin(router *gin.Engine) {
	if !psdk.Config.Options.Enable {
		return
	}
	InstrumentGin(router, psdk.Config)
}

// getFilters reads the ignoreIncomingRequests array and produces a otelgin.Option
// to fiter out requests that match the regex.
func getFilters() otelgin.Option {
	if len(psdk.Config.Options.IgnoreIncomingRequests) == 0 {
		return nil
	}

	return otelgin.WithFilter(func(r *http.Request) bool {
		for _, f := range psdk.Config.Options.IgnoreIncomingRequests {
			matches, _ := regexp.MatchString(f, r.URL.Path)

			if matches {
				return false
			}
		}

		return true
	})
}

// getMiddlewareOptions returns an array of otelgin.Option(s), these can include
// filters, propagators or formatters.
//
// If no option has been selected by the user, an empty array is returned.
func getMiddlewareOptions() []otelgin.Option {
	var middlewareOptions []otelgin.Option

	// Add filters
	f := getFilters()
	if f != nil {
		middlewareOptions = append(middlewareOptions, f)
	}

	return middlewareOptions
}

func InstrumentGin(router *gin.Engine, sdkconfig *pminterfaces.PostmanSDKConfig) {
	router.Use(otelgin.Middleware("", getMiddlewareOptions()...))
	router.Use(instrumentations_gin.Middleware(sdkconfig))
}
