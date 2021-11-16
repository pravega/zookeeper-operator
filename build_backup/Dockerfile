FROM ubuntu:20.04

WORKDIR /zookeeper
COPY backup.sh /zookeeper/backup.sh
RUN chmod +x /zookeeper/backup.sh
RUN apt-get update && apt-get install -y netcat
