package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	HealthGroup       = "/health"
	GetHealthEndpoint = "/"
)

func AddHealthRoutes(rg *gin.RouterGroup) {
	healthRouter := rg.Group("health")
	healthRouter.POST(GetHealthEndpoint, getHealthHandler)
}

func getHealthHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "")
}
