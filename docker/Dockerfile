#
# Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#

FROM gradle:jdk17

RUN mkdir /zu
COPY zu /zu
WORKDIR /zu
RUN ./gradlew shadowJar

# use forked base zookeeper 3.7.1 docker image
# that runs on openjdk-17
FROM adobe/zookeeper:3.7.1-apache-zk-20220513
COPY bin /usr/local/bin
RUN chmod +x /usr/local/bin/*
COPY --from=0 /zu/build/libs/zu.jar /opt/libs/

RUN apt-get -q update && \
    apt-get install --no-install-recommends -y dnsutils curl procps socat
