FROM alpine:3.8

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/bin/zookeeper-operator-linux-amd64 /usr/local/bin/zookeeper-operator
