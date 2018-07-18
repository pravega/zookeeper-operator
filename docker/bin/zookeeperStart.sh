#!/usr/bin/env bash
set -ex

source /conf/env.sh
source /usr/local/bin/zookeeperFunctions.sh

HOST=`hostname -s`
DATA_DIR=/data
MYID_FILE=$DATA_DIR/myid
LOG4J_CONF=/conf/log4j-quiet.properties
DYNCONFIG=$DATA_DIR/zoo.cfg.dynamic

# Extract cluster name and this members ordinal value from pod hostname
if [[ $HOST =~ (.*)-([0-9]+)$ ]]; then
    NAME=${BASH_REMATCH[1]}
    ORD=${BASH_REMATCH[2]}
else
    echo "Failed to parse name and ordinal of Pod"
    exit 1
fi

MYID=$((ORD+1))

# Check validity of on-disk configuration
WRITE_CONFIGURATION=true
REGISTER_NODE=true

if [ -e $MYID_FILE ]; then
  EXISTING_ID="`cat $DATA_DIR/myid`"
  if [ "$EXISTING_ID" == "$MYID" && -e $DYNCONFIG ]; then
      echo $MYID_FILE found, Existing myid matches $MYID, using existing configuration
      WRITE_CONFIGURATION=false
  fi
fi

# The first node is a simple case. There is no cluster to join, so simply create
# a local config
if [[ $MYID -eq 1 && "$WRITE_CONFIGURATION" == true ]]; then
  ROLE=participant
  echo Initial initialization of ordinal 0 pod, creating new config.
  ZKCONFIG=$(zkConfig)

  echo Writing bootstrap configuration with the following config:
  echo $ZKCONFIG
  echo $MYID > $MYID_FILE
  echo "server.${MYID}=${ZKCONFIG}" > $DYNCONFIG
else
  ROLE=observer
  set -e
  set +e
  ZKURL=$(zkConnectionString)

  # Check to see if server is already joined to the cluster
  set -e
  CONFIG=`java -Dlog4j.configuration=file:"$LOG4J_CONF" -jar /root/zu.jar get-all $ZKURL`
  REGISTERED=`echo "$CONFIG" | grep "server.${MYID}=" | wc -l`
  if [[ $REGISTERED -eq 1 ]]; then
    REGISTER_NODE=false
    WRITE_CONFIGURATION=true
  fi

  if [[ "$WRITE_CONFIGURATION" == true ]]; then
    echo "Writing myid: $MYID to: $MYID_FILE."
    echo $MYID > $MYID_FILE

    ZKCONFIG=$(zkConfig)

    if [[ "$REGISTER_NODE" == true ]]; then
      echo Registering node and writing local configuration to disk.
      java -Dlog4j.configuration=file:"$LOG4J_CONF" -jar /root/zu.jar add $ZKURL $MYID  $ZKCONFIG $DYNCONFIG

    else
      echo Writing configuration gleaned from zookeeper cluster
      echo "$CONFIG" | grep -v "^version="> $DYNCONFIG
    fi
  fi
fi

echo "Starting zookeeper service"
zkServer.sh start-foreground
