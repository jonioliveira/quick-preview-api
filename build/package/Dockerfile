FROM quickpreview/base:latest as tools
ENV GO111MODULE on
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.27.0 && \
  go get -u github.com/swaggo/swag/cmd/swag@v1.6.5 && \
  go get github.com/vektra/mockery/.../@v1.0.0

FROM quickpreview/base:latest AS base
COPY --from=tools /root/go/bin/golangci-lint /root/go/bin/swag /root/go/bin/mockery /go/bin/
RUN apt install make
WORKDIR /opt/app


# FROM base AS dependency-cache
# COPY go.mod go.sum ./
# RUN go mod download
# #   go mod verify
# ARG VERSION
# ENV VERSION=$VERSION

# FROM dependency-cache AS development
# CMD ["sleep infinity"]


# FROM dependency-cache AS build
# COPY . .
# RUN make shell-go-generate-build

# FROM debian:10 as production
# ENV GIN_MODE release
# USER 2000:2000
# WORKDIR /opt/app
# COPY --from=build /opt/app/bin/quick-preview-api ./quick-preview-api
# CMD ["./quick-preview-api"]
