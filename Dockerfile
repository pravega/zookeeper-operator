FROM alpine:3.8

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /

COPY ./zookeeper-operator .
RUN chmod +x /zookeeper-operator

ENTRYPOINT ["zookeeper-operator"]

