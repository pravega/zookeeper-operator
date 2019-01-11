#
# Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#

FROM openjdk:latest
RUN mkdir /zu
COPY zu /zu
WORKDIR /zu
RUN ./gradlew assemble

FROM zookeeper:3.5.4-beta
COPY bin /usr/local/bin
RUN chmod +x /usr/local/bin/*
COPY --from=0 /zu/build/libs/zu.jar /root/
