package preview

import (
	"fmt"
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

	if err == nil {
		ctx.JSON(http.StatusOK, "")
	}

	// TODO: STEP_1 clone the git repository

	// TODO: STEP_2 pass the path to the Dockerfile on the root of the git cloned project
	port, err := getPortFromDockerfile("./Dockerfile")
	if err != nil {
		panic(err)
	}

	// TODO: STEP_3 build docker image and push to repository

	// TODO: STEP_4 deploy helm with the image and pass the port for the service from EXPOSE

}

func getPortFromDockerfile(pathToDockerfile string) (string, error) {
	foundExpose := false
	file, err := os.Open(pathToDockerfile)
	if err != nil {
		return "", fmt.Errorf("Error: ", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		if scanner.Text() == "EXPOSE" {
			foundExpose = true
		}
		if foundExpose == true && scanner.Text() != "EXPOSE" {
			return scanner.Text(), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("Error: ", err)
	}
	return "", nil
}
