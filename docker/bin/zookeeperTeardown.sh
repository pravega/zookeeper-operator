#!/usr/bin/env bash
set -ex

source /conf/env.sh
source /usr/local/bin/zookeeperFunctions.sh

# Check to see if zookeeper service for this node is a participant
ZKURL=$(zkConnectionString)
MYID=`cat $MYID_FILE`

java -jar -Dlog4j.configuration=file:"$LOG4J_CONF" /root/zu.jar remove $ZKURL $MYID
