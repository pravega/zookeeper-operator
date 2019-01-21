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

HOST=`hostname -s`
DATA_DIR=/data
MYID_FILE=$DATA_DIR/myid
LOG4J_CONF=/conf/log4j-quiet.properties

OK=$(echo ruok | nc 127.0.0.1 $CLIENT_PORT)

# Check to see if zookeeper service answers
if [[ "$OK" == "imok" ]]; then

  # Extract resource name and this members ordinal value from pod hostname
  if [[ $HOST =~ (.*)-([0-9]+)$ ]]; then
      NAME=${BASH_REMATCH[1]}
      ORD=${BASH_REMATCH[2]}
  else
      echo "Failed to parse name and ordinal of Pod"
      exit 1
  fi

  if [[ $ORD -eq 0 ]]; then
    echo "Zookeeper service is available and an active participant"
    exit 0
  fi

  # Check to see if zookeeper service for this node is a participant
  set +e
  ZKURL=$(zkConnectionString)
  set -e
  MYID=`cat $MYID_FILE`
  ROLE=`java -Dlog4j.configuration=file:"$LOG4J_CONF" -jar /root/zu.jar get-role $ZKURL $MYID`

  if [[ "$ROLE" == "participant" ]]; then
    echo "Zookeeper service is available and an active participant"
    exit 0

  elif [[ "$ROLE" == "observer" ]]; then
    echo "Zookeeper service is ready to be upgraded from observer to participant."
    ROLE=participant
    ZKCONFIG=$(zkConfig)
    java -Dlog4j.configuration=file:"$LOG4J_CONF" -jar /root/zu.jar remove $ZKURL $MYID
    sleep 1
    java -Dlog4j.configuration=file:"$LOG4J_CONF" -jar /root/zu.jar add $ZKURL $MYID $ZKCONFIG
    exit 1

  else
    echo "Something has gone wrong. Unable to determinal zookeeper role."
    exit 1

  fi

else
  echo "Zookeeper service is not available for requests"
  exit 1

fi
