ARG DOCKER_REGISTRY
ARG ALPINE_VERSION=3.15
ARG BUILDPLATFORM=linux/amd64
FROM --platform=$BUILDPLATFORM ${DOCKER_REGISTRY:+$DOCKER_REGISTRY/}golang:1.18-alpine${ALPINE_VERSION} as go-builder

ARG PROJECT_NAME=zookeeper-operator
ARG REPO_PATH=github.com/pravega/$PROJECT_NAME

# Build version and commit should be passed in when performing docker build
ARG VERSION=0.0.0-localdev
ARG GIT_SHA=0000000

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /src
COPY pkg ./pkg
COPY cmd ./cmd
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Download all dependencies.
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o /src/${PROJECT_NAME} \
    -ldflags "-X ${REPO_PATH}/pkg/version.Version=${VERSION} -X ${REPO_PATH}/pkg/version.GitSHA=${GIT_SHA}" main.go

FROM ${DOCKER_REGISTRY:+$DOCKER_REGISTRY/}alpine:${ALPINE_VERSION} AS final

ARG PROJECT_NAME=zookeeper-operator

COPY --from=go-builder /src/${PROJECT_NAME} /usr/local/bin/${PROJECT_NAME}

RUN adduser -D ${PROJECT_NAME}
USER ${PROJECT_NAME}

ENTRYPOINT ["/usr/local/bin/zookeeper-operator"]
