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
DYNCONFIG=$DATA_DIR/zoo.cfg.dynamic
STATIC_CONFIG=/data/conf/zoo.cfg

# Extract resource name and this members ordinal value from pod hostname
if [[ $HOST =~ (.*)-([0-9]+)$ ]]; then
    NAME=${BASH_REMATCH[1]}
    ORD=${BASH_REMATCH[2]}
else
    echo Failed to parse name and ordinal of Pod
    exit 1
fi

MYID=$((ORD+1))

# Values for first startup
WRITE_CONFIGURATION=true
REGISTER_NODE=true
ONDISK_CONFIG=false

# Check validity of on-disk configuration
if [ -f $MYID_FILE ]; then
  EXISTING_ID="`cat $DATA_DIR/myid`"
  if [[ "$EXISTING_ID" == "$MYID" && -f $STATIC_CONFIG ]]; then
    # If Id is correct and configuration is present under `/data/conf`
      ONDISK_CONFIG=true
  fi
fi

# Determine if there is a ensemble available to join by checking the service domain
set +e
nslookup $DOMAIN
if [[ $? -eq 1 ]]; then
  # If an nslookup of the headless service domain fails, then there is no
  # active ensemble
  ACTIVE_ENSEMBLE=false
else
  ACTIVE_ENSEMBLE=true
fi

if [[ "$ONDISK_CONFIG" == true ]]; then
  # If Configuration is present, we assume, there is no need to write configuration.
    WRITE_CONFIGURATION=false
else
    WRITE_CONFIGURATION=true
fi

if [[ "$ACTIVE_ENSEMBLE" == false ]]; then
  # This is the first node being added to the cluster
  REGISTER_NODE=false
else
  # An ensemble exists, check to see if this node is already a member.
  if [[ "$ONDISK_CONFIG" == false ]]; then
    REGISTER_NODE=true
  else
    REGISTER_NODE=false
  fi
fi

if [[ "$WRITE_CONFIGURATION" == true ]]; then
  echo "Writing myid: $MYID to: $MYID_FILE."
  echo $MYID > $MYID_FILE
  if [[ $MYID -eq 1 ]]; then
    ROLE=participant
    echo Initial initialization of ordinal 0 pod, creating new config.
    ZKCONFIG=$(zkConfig)
    echo Writing bootstrap configuration with the following config:
    echo $ZKCONFIG
    echo $MYID > $MYID_FILE
    echo "server.${MYID}=${ZKCONFIG}" > $DYNCONFIG
  else
    echo Writing configuration gleaned from zookeeper ensemble
    echo "$CONFIG" | grep -v "^version="> $DYNCONFIG
  fi
fi

if [[ "$REGISTER_NODE" == true ]]; then
    ROLE=observer
    ZKURL=$(zkConnectionString)
    ZKCONFIG=$(zkConfig)
    echo Registering node and writing local configuration to disk.
    java -Dlog4j.configuration=file:"$LOG4J_CONF" -jar /root/zu.jar add $ZKURL $MYID  $ZKCONFIG $DYNCONFIG
fi

ZOOCFGDIR=/data/conf
export ZOOCFGDIR
if [[ ! -d "$ZOOCFGDIR" ]]; then
  echo Copying /conf contents to writable directory, to support Zookeeper dynamic reconfiguration
  mkdir $ZOOCFGDIR
  cp -f /conf/zoo.cfg $ZOOCFGDIR
  cp -f /conf/log4j.properties $ZOOCFGDIR
  cp -f /conf/log4j-quiet.properties $ZOOCFGDIR
  cp -f /conf/env.sh $ZOOCFGDIR
fi

echo Starting zookeeper service
zkServer.sh --config $ZOOCFGDIR start-foreground
