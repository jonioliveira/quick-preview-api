package preview

import (
	b64 "encoding/base64"

	"fmt"
	"net/http"
	"strings"

	validator "github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"

	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

const (
	DeployGroup        = "/"
	PostDeployEndpoint = "/deploy"
)

type Deploy struct {
	Namespace  string `json:"namespace" validate:"required"`
	Kubeconfig string `json:"kubeconfig"  validate:"required"`
	Repository string `json:"repository" validate:"required,url"`
}

func AddPreviewRoutes(rg *gin.RouterGroup) {
	previewRouter := rg.Group(DeployGroup)
	previewRouter.POST(PostDeployEndpoint, postPreviewHandler)
}

func postPreviewHandler(ctx *gin.Context) {
	var deploy Deploy
	err := ctx.Bind(&deploy)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"errors": "Route not found"})
	}

	errors := ValidateDeployRequest(&deploy)
	if errors != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	kubeconfig, err := b64.StdEncoding.DecodeString(deploy.Kubeconfig)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"kubeconfig": "Invalid kubeconfig"})
		return
	}

	//decoded kubeconfig
	fmt.Println(string(kubeconfig))

	fs := memfs.New()

	repo, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL: deploy.Repository,
	})
	fmt.Println(repo)

	file, err := fs.Open("Dockerfile")
	fmt.Println(file)

}

func ValidateDeployRequest(i interface{}) map[string]string {
	errors := make(map[string]string)
	validate := validator.New()
	if err := validate.Struct(i); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errors[strings.ToLower(err.Field())] = fmt.Sprintf("%s is %s %s", err.Field(), err.Tag(), err.Param())
		}

		return errors
	}

	return nil
}
