FROM openjdk:latest
RUN mkdir /zu
COPY zu /zu
WORKDIR /zu
RUN ./gradlew assemble

FROM zookeeper:3.5.4-beta
COPY bin /usr/local/bin
RUN chmod +x /usr/local/bin/*
COPY --from=0 /zu/build/libs/zu.jar /root/
