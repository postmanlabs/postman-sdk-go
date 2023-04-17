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
	Rules := map[string]string{
		"amazonAccessKeyId": "AKIA[0-9A-Z]{16}",
	}

	router := gin.Default()
	cleanup, err := pm.Initialize(collectionId, apiKey, pminterfaces.WithReceiverBaseUrl("https://trace-receiver.postman-preview.com"),
		pminterfaces.WithRedactSensitiveData(true, Rules))

	if err == nil {
		defer cleanup(context.Background())
		pm.InstrumentGin(router)
	}

	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	router.Run("localhost:8080")

}
