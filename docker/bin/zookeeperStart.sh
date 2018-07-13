#!/usr/bin/env bash
set -e

# INITIALIZE CONTANTS

HOST=`hostname -s`
DATA_DIR="/data"
MYID_FILE="$DATA_DIR/myid"
DYNCONFIG="$DATA_DIR/zoo.cfg.dynamic"
DOMAIN=$1
CLIENT_PORT=$2
QUORUM_PORT=$3
LEADER_PORT=$4

source /usr/local/bin/zookeeperFunctions.sh

# Extract cluster name and this members ordinal value from pod hostname
if [[ $HOST =~ (.*)-([0-9]+)$ ]]; then
    NAME=${BASH_REMATCH[1]}
    ORD=${BASH_REMATCH[2]}
else
    echo "Failed to parse name and ordinal of Pod"
    exit 1
fi

MYID=$((ORD+1))

# Check existing myid file

WRITE_CONFIGURATION=true

if [ -e $MYID_FILE ]; then
  EXISTING_ID="`cat $DATA_DIR/myid`"
  if [ "$EXISTING_ID" == "$MYID" ]; then
      echo $MYID_FILE found, Existing myid matches $MYID, using existing configuration
      WRITE_CONFIGURATION=false
  fi
fi

# Write myid file and overwrite dynamic configuration

if [ "$WRITE_CONFIGURATION" == true ]; then
  echo "Writing myid: $MYID to: $MYID_FILE."
  echo $MYID > $MYID_FILE

  if [[ $MYID -eq 1 ]]; then
    ROLE=participant
    echo Initial initialization of ordinal 0 pod, creating new config.
    ZKCONFIG=$(zkConfig)

    echo Writing bootstrap configuration with the following config:
    echo $ZKCONFIG
    echo "server.${MYID}=${ZKCONFIG}" > $DYNCONFIG

  else
    echo "On incoming observer. Pulling config via Zookeeper client."
    ROLE=observer
    ZKURL=$(zkConnectionString)
    ZKCONFIG=$(zkConfig)

    echo Using ZKURL:
    echo $ZKURL

    echo Adding server to ensemble with config:
    echo $ZKCONFIG

    echo "Updating and writing local configuration to disk."
    java -jar /root/zu.jar add $ZKURL $MYID  $ZKCONFIG $DYNCONFIG
  fi
fi

echo "Starting zookeeper service"
zkServer.sh start-foreground
