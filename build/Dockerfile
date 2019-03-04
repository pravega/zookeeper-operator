FROM alpine:3.8

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/bin/linux_amd64/zookeeper-operator /usr/local/bin/zookeeper-operator
