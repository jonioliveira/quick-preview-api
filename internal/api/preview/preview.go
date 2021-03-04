package preview

import (
	"bufio"
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/jonioliveira/quick-preview-api/pkg/logger"

	validator "github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
)

const (
	GithubRepoPath = "/tmp/gitrepo"

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

	// decoded kubeconfig
	fmt.Println(string(kubeconfig))

	// TODO: STEP_1 clone the git repository
	_ = removeDirRecursively(GithubRepoPath)

	_, err = git.PlainClone(GithubRepoPath, false, &git.CloneOptions{
		URL:      deploy.Repository,
		Progress: os.Stdout,
	})
	if err != nil {
		ctxErr := ctx.Error(err)
		if ctxErr != nil {
			logger.Error("Failed to attach error to context")
		}
		return
	}

	// TODO: STEP_2 pass the path to the Dockerfile on the root of the git cloned project
	port, err := getPortFromDockerfile(GithubRepoPath + "/Dockerfile")
	if err != nil {
		ctxErr := ctx.Error(err)
		if ctxErr != nil {
			logger.Error("Failed to attach error to context")
		}
		return
	}
	println("port: ", port)

	// TODO: STEP_3 build docker image and push to repository

	// TODO: STEP_4 deploy helm with the image and pass the port for the service from EXPOSE
	args := []string{
		"--atomic",
		"--debug",
		"--timeout 15m",
		"--install",
		"--namespace " + deploy.Namespace,
		"--set image.repository=" + "some-image-url-from-docker-io",
		"--set image.tag=latest",
	}

	output, err := RunCMD("helm upgrade", args)
	if err != nil {
		fmt.Println("Error:", output)
		return
	}

	fmt.Println("Result:", output)

	ctx.JSON(http.StatusOK, "Deployed at: ${FILL WITH DNS FROM CMY}")
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

func getPortFromDockerfile(pathToDockerfile string) (string, error) {
	foundExpose := false
	file, err := os.Open(pathToDockerfile)
	if err != nil {
		return "", fmt.Errorf("Error: %s", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		if scanner.Text() == "EXPOSE" {
			foundExpose = true
		}
		if foundExpose && scanner.Text() != "EXPOSE" {
			return scanner.Text(), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("Error: %s", err)
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

func RunCMD(path string, args []string) (out string, err error) {

	cmd := exec.Command(path, args...)

	var b []byte
	b, err = cmd.CombinedOutput()
	out = string(b)

	fmt.Println(strings.Join(cmd.Args[:], " "))

	if err != nil {
		fmt.Println("RunCMD ERROR")
		fmt.Println(out)
	}

	return
}