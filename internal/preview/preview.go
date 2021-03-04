package preview

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"

	"github.com/gin-gonic/gin"
)

const GIT_REPO_PATH = "/tmp/gitrepo"

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
	removeDirRecursively(GIT_REPO_PATH)

	_, err = git.PlainClone(GIT_REPO_PATH, false, &git.CloneOptions{
		URL:      newPreview.Repository,
		Progress: os.Stdout,
	})
	if err != nil {
		ctx.Error(err)
		return
	}

	// TODO: STEP_2 pass the path to the Dockerfile on the root of the git cloned project
	port, err := getPortFromDockerfile(GIT_REPO_PATH + "/Dockerfile")
	if err != nil {
		ctx.Error(err)
		return
	}
	println("port: ", port)

	// TODO: STEP_3 build docker image and push to repository

	// TODO: STEP_4 deploy helm with the image and pass the port for the service from EXPOSE

	ctx.JSON(http.StatusOK, "Deployed at: ${FILL WITH DNS FROM CMY}")
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

func removeDirRecursively(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
