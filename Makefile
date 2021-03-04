
.DEFAULT_GOAL := help

export LC_ALL=en_US.UTF-8
export PROJECT_ROOT=$(shell pwd)
export PATH=$(shell (echo "$$(go env GOPATH 2> /dev/null)/bin:" || echo ""))$(shell echo $$PATH)
export REPOSITORY?=quickpreview/api
export VERSION?=dev-latest

GO_LDFLAGS=-ldflags ""
# By default -count=1 for no cache.
# -p number of paralel processes allowed
GO_TEST_FLAGS?=-count=1 -p=4
PORT?=8081
DOCKER_LOCAL_IMAGE=$(REPOSITORY):dev-local
DOCKER_DEV_BUILD=docker build -f build/package/Dockerfile --target development --tag $(DOCKER_LOCAL_IMAGE) --build-arg VERSION .
DOCKER_RUN_BASE=docker run --rm -v $$PROJECT_ROOT:/opt/app/ -v /opt/app/bin -v $$PROJECT_ROOT/.cache/:/.cache/ -v /var/run/docker.sock:/var/run/docker.sock -p $(PORT):8081 -e GOCACHE=/.cache/go-build -e GOLANGCI_LINT_CACHE=/.cache/golangci-lint
DOCKER_DEV_RUN=$(DOCKER_RUN_BASE) $(DOCKER_LOCAL_IMAGE)
DOCKER_DEV_RUN_IT=$(DOCKER_RUN_BASE) -it $(DOCKER_LOCAL_IMAGE)
DOCKER_COMPOSE=docker-compose -f deployments/docker/compose.yaml -p quick-preview-api

## General

# target: help - Display available recipes.
.PHONY: help
help:
	@egrep "^# target:" [Mm]akefile

## Shell

# Build the go binary.
.PHONY: shell-go-build
shell-go-build:
	CGO_ENABLED=0 go build ${GO_LDFLAGS} -o bin/quick-preview-api cmd/quick-preview-api/main.go

# Run tests.
.PHONY: shell-go-test
shell-go-test:
	go test ${GO_LDFLAGS} $(GO_TEST_FLAGS) ./...

# Run the app.
.PHONY: shell-go-run
shell-go-run:
	go run  ${GO_LDFLAGS} cmd/quick-preview-api/main.go $(filter-out $@,$(MAKECMDGOALS))

# Clean the cache.
.PHONY: shell-clean-cache
shell-clean-cache:
	rm -Rf $$PROJECT_ROOT/.cache


## Docker

# target: docker-sh - Run a sh shell inside the container.
.PHONY: docker-sh
docker-sh:
	$(DOCKER_DEV_BUILD)
	$(DOCKER_DEV_RUN_IT) sh

# Build the app inside the container.
.PHONY: docker-build-app
docker-build-app:
	$(DOCKER_DEV_BUILD)
	$(DOCKER_DEV_RUN) make shell-build

# Run the app inside the container.
.PHONY: docker-run-app-only
docker-run-app-only:
	$(DOCKER_DEV_BUILD)
	$(DOCKER_DEV_RUN_IT) make shell-run

# Run the app and its dependencies inside the container (docker-compose).
.PHONY: docker-run-app
docker-run-app:
	$(DOCKER_DEV_BUILD)
	${DOCKER_COMPOSE} up

# Build the app container image.
.PHONY: docker-build
docker-build:
	docker build -f build/package/Dockerfile --target production --tag $(REPOSITORY):${VERSION} .

# Push the app container image.
.PHONY: docker-push
docker-push:
	docker push ${REPOSITORY}:${VERSION}

# Delete the container image and its assets.
.PHONY: docker-clean
docker-clean:
	docker rmi -f $(DOCKER_LOCAL_IMAGE)


## Alias

# target: build - Build the app (inside the container).
.PHONY: build
build: docker-build-app

# target: test - Run app tests (inside the container).
.PHONY: test
test: docker-test-app

# target: run - Run the app {inside the container}.
.PHONY: run
run: docker-run-app

# target: clean - Clean cache and local docker image.
.PHONY: clean
clean: shell-clean-cache docker-clean
