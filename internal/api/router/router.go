package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jonioliveira/quick-preview-api/internal/api/preview"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

const (
	v1 = "/v1"
)

func BuildRouter() *gin.Engine {
	router := gin.Default()
	p := ginprometheus.NewPrometheus("gin")
	p.Use(router)

	apiV1 := router.Group(v1)
	preview.AddPreviewRoutes(apiV1)

	return router
}
