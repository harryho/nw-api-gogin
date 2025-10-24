package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"

	"github.com/glb/nw-api-gogin/internal/api"
)

func main() {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatal(err)
	}
	swagger.Servers = nil

	r.Use(ginmiddleware.OapiRequestValidator(swagger))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	h := api.NewHandler()
	api.RegisterHandlersWithOptions(r, h, api.GinServerOptions{
		ErrorHandler: func(c *gin.Context, handlerErr error, statusCode int) {
			c.JSON(statusCode, api.ErrorResponse{Code: "invalid_request", Message: handlerErr.Error()})
		},
	})

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
