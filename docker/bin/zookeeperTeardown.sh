#!/usr/bin/env bash
set -ex

source /conf/env.sh
source /usr/local/bin/zookeeperFunctions.sh

DATA_DIR=/data
MYID_FILE=$DATA_DIR/myid
LOG4J_CONF=/conf/log4j-quiet.properties

# Check to see if zookeeper service for this node is a participant
ZKURL=$(zkConnectionString)
MYID=`cat $MYID_FILE`

# Remove server from zk configuration
java -Dlog4j.configuration=file:"$LOG4J_CONF" -jar /root/zu.jar remove $ZKURL $MYID

# Wait for client connections to drain. Kubernetes will wait until the confiugred
# "terminationGracePeriodSeconds" before focibly killing the container
CONN_COUNT=`echo cons | nc localhost 2181 | grep -v "^$" |grep -v "/127.0.0.1:" | wc -l`
for (( i = 0; i < 36; i++ )); do
  echo "$CONN_COUNT non-local connections still connected."
  sleep 5
  CONN_COUNT=`echo cons | nc localhost 2181 | grep -v "^$" |grep -v "/127.0.0.1:" | wc -l`
done

# Kill the primary process ourselves to circumvent the terminationGracePeriodSeconds
PID=`ps -ef | grep zoo.cfg | grep -v grep | awk '{print $1}'`
kill $PID
