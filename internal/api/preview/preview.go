package preview

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Preview struct {
	Namespace  string `json:"namespace" binding:"required"`
	Kubeconfig string `json:"kubeconfig" binding:"required"`
	Repository string `json:"repository" binding:"required"`
}

func AddPreviewRoutes(rg *gin.RouterGroup) {
	previewRouter := rg.Group("/preview")
	previewRouter.POST("/", postPreviewHandler)
}

func postPreviewHandler(ctx *gin.Context) {
	var newPreview Preview
	err := ctx.Bind(&newPreview)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "")
		return
	}

	ctx.JSON(http.StatusOK, "")
}
