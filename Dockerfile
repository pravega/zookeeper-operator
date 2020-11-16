# build builder
ARG GO_VERSION=1.13.8
ARG ALPINE_VERSION=3.11
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as go-builder

ARG CODEPATH


WORKDIR $GOPATH/src/$CODEPATH


COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOPROXY=https://mirrors.aliyun.com/goproxy/  go build -o /zookeeper-operator ./cmd/manager/main.go

# build server
FROM alpine:3.8

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /

COPY --from=builder /zookeeper-operator .
RUN chmod +x /zookeeper-operator

ENTRYPOINT ["zookeeper-operator"]

