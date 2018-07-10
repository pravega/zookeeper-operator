FROM openjdk:latest
RUN mkdir /add-zk-member
COPY add-zk-member-src /add-zk-member
WORKDIR /add-zk-member
RUN ./gradlew assemble

FROM openjdk:latest
COPY bin /usr/local/bin
WORKDIR /root/
COPY --from=0 /add-zk-member/build/libs/add-zk-member.jar /root/
ENTRYPOINT ["initializeConfig.sh"]
CMD []
