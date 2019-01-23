FROM golang:1.10.1-alpine3.7 as go-builder

ARG PROJECT_NAME=zookeeper-operator
ARG REPO_PATH=github.com/pravega/$PROJECT_NAME
ARG BUILD_PATH=${REPO_PATH}/cmd/manager

# Build version and commit should be passed in when performing docker build
ARG VERSION=0.0.0-localdev
ARG GIT_SHA=0000000

RUN mkdir -p /go/src/${REPO_PATH}/vendor

COPY pkg /go/src/${REPO_PATH}/pkg
COPY cmd /go/src/${REPO_PATH}/cmd
COPY vendor /go/src/${REPO_PATH}/vendor

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${GOBIN}/${PROJECT_NAME} \
    -ldflags "-X ${REPO_PATH}/pkg/version.Version=${VERSION} -X ${REPO_PATH}/pkg/version.GitSHA=${GIT_SHA}" \
    $BUILD_PATH

# =============================================================================
FROM alpine:3.7 AS final

ARG PROJECT_NAME=zookeeper-operator
ARG REPO_PATH=github.com/pravega/$PROJECT_NAME

COPY --from=go-builder ${GOBIN}/${PROJECT_NAME} /usr/local/bin/${PROJECT_NAME}

RUN adduser -D ${PROJECT_NAME}
USER ${PROJECT_NAME}

ENTRYPOINT ["/usr/local/bin/zookeeper-operator"]
