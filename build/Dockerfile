FROM alpine:3.8
RUN apk upgrade --update --no-cache
USER nobody
COPY build/_output/bin/zookeeper-operator /usr/local/bin/zookeeper-operator
