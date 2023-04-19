package main

import (
	"github.com/gin-gonic/gin"

	pm "github.com/postmanlabs/postman-go-sdk/postmansdk"
)

func main() {
	apiKey := "REPLACE-THIS"
	collectionId := "REPLACE-THIS"
	Rules := map[string]string{
		"amazonAccessKeyId": "AKIA[0-9A-Z]{16}",
	}

	router := gin.Default()
	psdk, err := pm.Initialize(
		collectionId,
		apiKey, pm.WithReceiverBaseUrl("REPLACE-THIS"),
		pm.WithRedactSensitiveData(true, Rules),
	)
	if err == nil {
		psdk.Integrations.Gin(router)
	}

	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	router.Run("localhost:8080")

}
