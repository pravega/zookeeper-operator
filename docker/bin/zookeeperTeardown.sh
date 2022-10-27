#!/usr/bin/env bash
#
# Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#

set -ex

source /conf/env.sh
source /usr/local/bin/zookeeperFunctions.sh

DATA_DIR=/data
MYID_FILE=$DATA_DIR/myid
LOG4J_CONF=/conf/log4j-quiet.properties

# Wait for client connections to drain. Kubernetes will wait until the confiugred
# "terminationGracePeriodSeconds" before focibly killing the container
for (( i = 0; i < 6; i++ )); do
  CONN_COUNT=`echo cons | socat stdio tcp:localhost:$CLIENT_PORT | grep -v "^$" |grep -v "/127.0.0.1:" | wc -l`
  if [[ "$CONN_COUNT" -gt 0 ]]; then
    echo "$CONN_COUNT non-local connections still connected."
    sleep 5
  else
    echo "$CONN_COUNT non-local connections"
    break
  fi
done

# Check to see if zookeeper service for this node is a participant
set +e
ZKURL=$(zkConnectionString)
set -e
MYID=`cat $MYID_FILE`

ZNODE_PATH="/zookeeper-operator/$CLUSTER_NAME"
CLUSTERSIZE=`java -Dlog4j.configuration=file:"$LOG4J_CONF" -jar /opt/libs/zu.jar sync $ZKURL $ZNODE_PATH`
echo "CLUSTER_SIZE=$CLUSTERSIZE, MyId=$MYID"
if [[ -n "$CLUSTERSIZE" && "$CLUSTERSIZE" -lt "$MYID" ]]; then
  # If ClusterSize < MyId, this server is being permanantly removed.
  java -Dlog4j.configuration=file:"$LOG4J_CONF" -jar /opt/libs/zu.jar remove $ZKURL $MYID
  echo $?
fi

# Kill the primary process ourselves to circumvent the terminationGracePeriodSeconds
ps -ef | grep zoo.cfg | grep -v grep | awk '{print $2}' | xargs kill
