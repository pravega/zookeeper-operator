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

# Extract resource name and this members ordinal value from pod hostname
if [[ $HOST =~ (.*)-([0-9]+)$ ]]; then
    NAME=${BASH_REMATCH[1]}
    ORD=${BASH_REMATCH[2]}
else
    echo "Failed to parse name and ordinal of Pod"
    exit 1
fi

MYID=$((ORD+1))

WRITE_CONFIGURATION=true
REGISTER_NODE=true

# Check validity of on-disk configuration
if [ -f $MYID_FILE ]; then
  EXISTING_ID="`cat $DATA_DIR/myid`"
  if [[ "$EXISTING_ID" == "$MYID" || -f $DYNCONFIG ]]; then
      WRITE_CONFIGURATION=false
  fi

fi

# Determine if there is a ensemble available to join by checking the service domain

set +e
nslookup $DOMAIN
if [[ $? -eq 1 ]]; then
  set -e
  # If an nslookup of the headless service domain fails, then there is no
  # active ensemble
  WRITE_CONFIGURATION=true
  REGISTER_NODE=false

else
  set -e
  # An ensemble exists, check to see if this node is already a member.

  set +e
  ZKURL=$(zkConnectionString)
  set -e
  CONFIG=`java -Dlog4j.configuration=file:"$LOG4J_CONF" -jar /root/zu.jar get-all $ZKURL`
  REGISTERED=`echo "$CONFIG" | grep "server.${MYID}=" | wc -l`

  if [[ $REGISTERED -eq 1 ]]; then
    REGISTER_NODE=false

  else
    # When registering the node to the ensemble, always [re]write the config
    REGISTER_NODE=true
    WRITE_CONFIGURATION=true

  fi
fi

if [[ "$WRITE_CONFIGURATION" == true ]]; then
  echo "Writing myid: $MYID to: $MYID_FILE."
  echo $MYID > $MYID_FILE

  if [[ $MYID -eq 1 && "$REGISTER_NODE" == false ]]; then
    ROLE=participant
    echo Initial initialization of ordinal 0 pod, creating new config.
    ZKCONFIG=$(zkConfig)

    echo Writing bootstrap configuration with the following config:
    echo $ZKCONFIG
    echo $MYID > $MYID_FILE
    echo "server.${MYID}=${ZKCONFIG}" > $DYNCONFIG

  elif [[ $MYID -ne 1 && "$REGISTER_NODE" == false ]]; then
    echo Writing configuration gleaned from zookeeper ensemble
    echo "$CONFIG" | grep -v "^version="> $DYNCONFIG

  elif [[ "$REGISTER_NODE" == true ]]; then
    ROLE=observer
    ZKCONFIG=$(zkConfig)

    echo Registering node and writing local configuration to disk.
    java -Dlog4j.configuration=file:"$LOG4J_CONF" -jar /root/zu.jar add $ZKURL $MYID  $ZKCONFIG $DYNCONFIG
  fi
fi

ZOOCFGDIR=/data/conf
export ZOOCFGDIR

if [ ! -d "$ZOOCFGDIR" ]
then
echo "Creating writable conf directory."
mkdir $ZOOCFGDIR
cp -f /conf/zoo.cfg $ZOOCFGDIR
cp -f /conf/log4j.properties $ZOOCFGDIR
cp -f /conf/log4j-quiet.properties $ZOOCFGDIR
cp -f /conf/env.sh $ZOOCFGDIR
fi

if [[ "$WRITE_CONFIGURATION" == false &&  "$REGISTER_NODE" == false ]]; then
  # We get here only on server restart...
  echo Printing bootstrap dynamic config file
  cat $DYNCONFIG
  echo Printing Static Config
  STATIC_CONFIG=`cat $ZOOCFGDIR/zoo.cfg`
  #Setting dynamicConfigFile=/data/zoo.cfg.dynamic in zoo.cfg
  sed -i 's/dynamicConfigFile=.*/dynamicConfigFile=\/data\/zoo\.cfg\.dynamic/g' $ZOOCFGDIR/zoo.cfg
  cat $ZOOCFGDIR/zoo.cfg
fi

echo "Starting zookeeper service"
zkServer.sh --config $ZOOCFGDIR start-foreground
