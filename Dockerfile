# build builder
FROM golang:1.14 as builder

ARG VERSION
ARG GIT_SHA
ARG PROJECT_NAME
ARG CODE_PATH

WORKDIR $GOPATH/src/$CODE_PATH

COPY . .

RUN GO111MODULE=on GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOPROXY=https://goproxy.cn go build -o /tmp/${PROJECT_NAME} -ldflags "-X ${CODE_PATH}/pkg/version.Version=${VERSION} -X ${CODE_PATH}/pkg/version.GitSHA=${GIT_SHA}" ${CODE_PATH}/cmd/manager

# build server
FROM alpine:3.8

ARG PROJECT_NAME

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /

COPY --from=builder /tmp/${PROJECT_NAME} /usr/local/bin/${PROJECT_NAME}

RUN adduser -D zookeeper-operator
USER ${PROJECT_NAME}

ENTRYPOINT ["/usr/local/bin/zookeeper-operator"]