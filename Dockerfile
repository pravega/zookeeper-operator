FROM golang:1.13.4-alpine3.10 as go-builder

ARG PROJECT_NAME=zookeeper-operator

# Build version and commit should be passed in when performing docker build
ARG VERSION=0.0.0-localdev
ARG GIT_SHA=0000000

WORKDIR /src

COPY pkg ./pkg
COPY cmd ./cmd
COPY go.mod* /src

RUN ls /src/ && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /src/${PROJECT_NAME} \
    -ldflags "-X ${REPO_PATH}/pkg/version.Version=${VERSION} -X ${REPO_PATH}/pkg/version.GitSHA=${GIT_SHA}" \
    /src/cmd/manager

# =============================================================================
FROM alpine:3.10 AS final

ARG PROJECT_NAME=zookeeper-operator

COPY --from=go-builder /src/${PROJECT_NAME} /usr/local/bin/${PROJECT_NAME}

RUN adduser -D ${PROJECT_NAME}
USER ${PROJECT_NAME}

ENTRYPOINT ["/usr/local/bin/zookeeper-operator"]
