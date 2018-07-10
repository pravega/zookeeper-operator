FROM openjdk:latest
RUN mkdir /add-zk-member
COPY ./ /add-zk-member
WORKDIR /add-zk-member
RUN ./gradlew assemble

FROM openjdk:latest
WORKDIR /root/
COPY --from=0 /add-zk-member/build/libs/add-zk-member.jar /root/
ENTRYPOINT ["java" "-jar" "/root/add-zk-member.jar"]
