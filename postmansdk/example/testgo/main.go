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
	pm.Initialize(
		collectionId,
		apiKey, pminterfaces.WithReceiverBaseUrl("REPLACE-THIS"),
		pminterfaces.WithRedactSensitiveData(true, Rules),
		pminterfaces.WithGinInstrumentation(router),
	)

	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	router.Run("localhost:8080")

}
