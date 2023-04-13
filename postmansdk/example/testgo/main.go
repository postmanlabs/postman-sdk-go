package main

import (
	"context"

	"github.com/gin-gonic/gin"

	pm "github.com/postmanlabs/postman-go-sdk/postmansdk"
	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
)

func main() {

	apiKey := "REPLACE-THIS"
	collectionId := "REPLACE-THIS"
	redactSensitiveData := pminterfaces.RedactSensitiveDataConfig{
		Enable: true,
		Rules: map[string]string{
			"rule1": "wonderful",
		},
	}

	cleanup, err := pm.Initialize(collectionId, apiKey, pminterfaces.WithReceiverBaseUrl("REPLACE THIS"),
		pminterfaces.WithRedactSensitiveData(redactSensitiveData))
	if err == nil {
		defer cleanup(context.Background())
	}

	router := gin.Default()

	// Otel patch
	pm.InstrumentGin(router)

	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	router.Run("localhost:8080")

}
