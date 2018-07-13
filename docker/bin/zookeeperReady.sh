#!/usr/bin/env bash
set -ex

DATA_DIR="/data"
MYID_FILE="$DATA_DIR/myid"
DOMAIN=$1
CLIENT_PORT=$2

source /usr/local/bin/zookeeperFunctions.sh

OK=$(echo ruok | nc 127.0.0.1 $CLIENT_PORT)

# Check to see if zookeeper service answers
if [[ "$OK" == "imok" ]]; then

  # Check to see if zookeeper service for this node is a participant
  ZKURL=$(zkConnectionString)
  MYID=`cat $MYID_FILE`
  ROLE=`java -jar /root/zu.jar get-role $ZKURL $MYID`

  if [[ "$ROLE" == "participant" ]]; then
  echo "Zookeeper service is available and an active participant"
  exit 0

  elif [[ "$ROLE" == "observer" ]]; then
    echo "Zookeeper service is ready to be upgraded from observer to participant"
    ROLE=participant
    ZKCONFIG=$(zkConfig)
    java -jar /root/zu.jar add $ZKURL $MYID  $ZKCONFIG
  else
  exit 1

  fi

else
  echo "Zookeeper service is not available for requests"
  exit 1

fi
