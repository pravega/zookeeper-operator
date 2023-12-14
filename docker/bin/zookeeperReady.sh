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
STATIC_CONFIG=/data/conf/zoo.cfg

OK=$(echo ruok | socat stdio tcp:localhost:$CLIENT_PORT)

# Check to see if zookeeper service answers
if [[ "$OK" == "imok" ]]; then
  set +e
  getent hosts $DOMAIN
  if [[ $? -ne 0 ]]; then
    set -e
    echo "There is no active ensemble, skipping readiness probe..."
    exit 0
  else
    set -e
    # An ensemble exists, check to see if this node is already a member.
    # Check to see if zookeeper service for this node is a participant
    set +e
    # Extract resource name and this members' ordinal value from pod hostname
    HOST=`hostname -s`
    if [[ $HOST =~ (.*)-([0-9]+)$ ]]; then
        NAME=${BASH_REMATCH[1]}
        ORD=${BASH_REMATCH[2]}
    else
        echo Failed to parse name and ordinal of Pod
        exit 1
    fi
    MYID=$((ORD+1))
    ONDISK_CONFIG=false
    if [ -f $MYID_FILE ]; then
      EXISTING_ID="`cat $DATA_DIR/myid`"
      if [[ "$EXISTING_ID" == "$MYID" && -f $STATIC_CONFIG ]]; then
      #If Id is correct and configuration is present under `/data/conf`
          ONDISK_CONFIG=true
          DYN_CFG_FILE_LINE=`cat $STATIC_CONFIG|grep "dynamicConfigFile\="`
          DYN_CFG_FILE=${DYN_CFG_FILE_LINE##dynamicConfigFile=}
          SERVER_FOUND=`cat $DYN_CFG_FILE | grep "server.${MYID}=" | wc -l`
          if [[ "$SERVER_FOUND" == "0" ]]; then
            echo "Server not found in ensemble. Exiting ..."
            exit 1
          fi
          SERVER=`cat $DYN_CFG_FILE | grep "server.${MYID}="`
          if [[ "$SERVER" == *"participant"* ]]; then
              ROLE=participant
          elif [[ "$SERVER" == *"observer"* ]]; then
              ROLE=observer
          fi
      fi
    fi

    if [[ "$ROLE" == "participant" ]]; then
      echo "Zookeeper service is available and an active participant"
      exit 0
    elif [[ "$ROLE" == "observer" ]]; then
      echo "Zookeeper service is ready to be upgraded from observer to participant."
      ROLE=participant
      ZKURL=$(zkConnectionString)
      ZKCONFIG=$(zkConfig)
      java -Dlog4j.configuration=file:"$LOG4J_CONF" -jar /opt/libs/zu.jar remove $ZKURL $MYID
      sleep 1
      java -Dlog4j.configuration=file:"$LOG4J_CONF" -jar /opt/libs/zu.jar add $ZKURL $MYID $ZKCONFIG
      exit 0
    else
      echo "Something has gone wrong. Unable to determine zookeeper role."
      exit 1
    fi
  fi

else
  echo "Zookeeper service is not available for requests"
  exit 1
fi
