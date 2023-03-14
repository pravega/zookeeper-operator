ARG DOCKER_REGISTRY
ARG DISTROLESS_DOCKER_REGISTRY
ARG ALPINE_VERSION=3.17
FROM ${DOCKER_REGISTRY:+$DOCKER_REGISTRY/}golang:1.19-alpine${ALPINE_VERSION} as go-builder

ARG PROJECT_NAME=zookeeper-operator
ARG REPO_PATH=github.com/pravega/$PROJECT_NAME

# Build version and commit should be passed in when performing docker build
ARG VERSION=0.0.0-localdev
ARG GIT_SHA=0000000

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
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /src/${PROJECT_NAME} \
    -ldflags "-X ${REPO_PATH}/pkg/version.Version=${VERSION} -X ${REPO_PATH}/pkg/version.GitSHA=${GIT_SHA}" main.go

FROM ${DISTROLESS_DOCKER_REGISTRY:-gcr.io/}distroless/static-debian11:nonroot AS final

ARG PROJECT_NAME=zookeeper-operator

COPY --from=go-builder /src/${PROJECT_NAME} /usr/local/bin/${PROJECT_NAME}

ENTRYPOINT ["/usr/local/bin/zookeeper-operator"]
