
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
DOCKER_RUN_BASE=docker run --rm -v $$PROJECT_ROOT:/opt/app/ -v /opt/app/bin -v $$PROJECT_ROOT/.cache/:/.cache/ -p $(PORT):8080 -e GOCACHE=/.cache/go-build -e GOLANGCI_LINT_CACHE=/.cache/golangci-lint
DOCKER_DEV_RUN=$(DOCKER_RUN_BASE) $(DOCKER_LOCAL_IMAGE)
DOCKER_DEV_RUN_IT=$(DOCKER_RUN_BASE) -it $(DOCKER_LOCAL_IMAGE)
DOCKER_COMPOSE=docker-compose -f deployments/docker/compose.yaml -p quick-preview-api

## General

# target: help - Display available recipes.
.PHONY: help
help:
	@egrep "^# target:" [Mm]akefile


## GIT

# Get the latest tag of the commit history.
.PHONY: git-latest-tag
git-latest-tag:
	@git describe --abbrev=0 --tags

# Get the tag of the current commit.
.PHONY: git-commit-tag
git-commit-tag:
	@git describe --exact-match --tags HEAD

# Get the version tag of the current commit or default to 'dev-latest'.
.PHONY: git-version-tag
git-version-tag:
	@(echo ${VERSION})

## Shell

# Run all go generate directives (code generation).
.PHONY: shell-go-generate
shell-go-generate:
	go generate ./...

# Generate mocks (for testing) from all interfaces.
.PHONY: shell-go-generate-mocks
shell-go-generate-mocks:
	grep -rl cmd/ pkg/ internal/ -e 'type \s*[a-zA-Z0-9]\+\s* interface' | xargs -n 1 dirname | uniq | xargs -P 4 -I % sh -c 'mockery -name "[a-zA-Z0-9]*" -note "Run \`make generate-mocks\` to regenerate this file." -outpkg mocks -dir "%" -output "%/mocks"'

# Lint all go files.
# No need to call `gofmt` since `golint` will be called.
.PHONY: shell-go-lint
shell-go-lint:
	golangci-lint run

# Lint and fix (if possible) all go files.
.PHONY: shell-go-fix
shell-go-fix:
	go fmt ./...
	golangci-lint run --fix
	go mod tidy

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


## Shell aggregators

.PHONY: shell-go-generate-lint
shell-go-generate-lint: shell-go-generate shell-go-lint

.PHONY: shell-go-fix-generate
shell-go-fix-generate: shell-go-fix shell-go-generate

.PHONY: shell-go-generate-build
shell-go-generate-build: shell-go-generate shell-go-build

.PHONY: shell-go-generate-test
shell-go-generate-test: shell-go-generate shell-go-test

.PHONY: shell-go-generate-run
shell-go-generate-run: shell-go-generate shell-go-run


## Docker

# target: docker-sh - Run a sh shell inside the container.
.PHONY: docker-sh
docker-sh:
	$(DOCKER_DEV_BUILD)
	$(DOCKER_DEV_RUN_IT) sh

# Run the linter inside the container.
.PHONY: docker-lint-app
docker-lint-app:
	$(DOCKER_DEV_BUILD)
	$(DOCKER_DEV_RUN) make shell-go-generate-lint

# Generate mocks (for testing) from all interfaces inside the container.
.PHONY: docker-generate-mocks
docker-generate-mocks:
	$(DOCKER_DEV_BUILD)
	$(DOCKER_DEV_RUN) make shell-go-generate-mocks

# Run the linter and fix (if possible) inside the container.
.PHONY: docker-fix-app
docker-fix-app:
	$(DOCKER_DEV_BUILD)
	$(DOCKER_DEV_RUN) make shell-go-fix-generate

# Build the app inside the container.
.PHONY: docker-build-app
docker-build-app:
	$(DOCKER_DEV_BUILD)
	$(DOCKER_DEV_RUN) make shell-go-generate-build

# Run app tests inside the container.
.PHONY: docker-test-app
docker-test-app:
	$(DOCKER_DEV_BUILD)
	$(DOCKER_DEV_RUN) make GO_TEST_FLAGS="$(GO_TEST_FLAGS)" shell-go-generate-test

# Run the app inside the container.
.PHONY: docker-run-app-only
docker-run-app-only:
	$(DOCKER_DEV_BUILD)
	$(DOCKER_DEV_RUN_IT) make shell-go-generate-run

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

# target: app-version - Get the current app version.
.PHONY: app-version
app-version: git-version-tag

# target: generate-mocks - Generate mocks (inside the container).
.PHONY: generate-mocks
generate-mocks: docker-generate-mocks

# target: lint - Run the linter (inside the container).
.PHONY: lint
lint: docker-lint-app

# target: fix - Run the linter and fix issues if possible (runs inside the container). Good to be called on file save in an IDE.
.PHONY: fix
fix: docker-fix-app

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

# target: ci-lint - Lint the app (inside the CI environment).
.PHONY: ci-lint
ci-lint: shell-go-generate-lint

# target: ci-test - Run app tests (inside the CI environment).
.PHONY: ci-test
ci-test: shell-go-generate-test

# target: ci-build - Build the go binary (inside the CI environment).
.PHONY: ci-build
ci-build: shell-go-generate-build

# target: ci-build-image - Build the app image (inside the CI environment).
.PHONY: ci-build-image-image
ci-build-image: docker-build docker-push

# target: ci-deploy - Deploy using helm (inside the CI environment).
.PHONY: ci-deploy
ci-deploy:
	./scripts/deploy.sh

# target: check-swagger - Check swagger documentation if is up-to-date
.PHONY: check-swagger
check-swagger: shell-go-generate
	git diff --exit-code api/openapi-spec/swagger.json api/openapi-spec/swagger.yaml
