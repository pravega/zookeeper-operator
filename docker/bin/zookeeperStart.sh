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

START_DELAY=${START_DELAY:-0}


HOST=`hostname -s`
DATA_DIR=/data
MYID_FILE=$DATA_DIR/myid
LOG4J_CONF=/conf/log4j-quiet.properties
DYNCONFIG=$DATA_DIR/zoo.cfg.dynamic
STATIC_CONFIG=/data/conf/zoo.cfg

# used when zkid starts from value grater then 1, default 1
OFFSET=${OFFSET:-1}

# Extract resource name and this members ordinal value from pod hostname
if [[ $HOST =~ (.*)-([0-9]+)$ ]]; then
    NAME=${BASH_REMATCH[1]}
    ORD=${BASH_REMATCH[2]}
else
    echo Failed to parse name and ordinal of Pod
    exit 1
fi

MYID=$(($ORD+$OFFSET))

# use SEED_NODE to bootstrap the current zookeeper cluster, else default to local cluster
# CLIENT_HOST is used in zkConnectionString function already to create zkURL
CLIENT_HOST=${SEED_NODE:-$CLIENT_HOST}


# use FQDN_TEMPLATE to create an OUTSIDE_NAME that is going to be used to establish quorum election
# this should be used along with the quorumListenOnAllIPs set to true
# % from the FQDN_TEMPLATE will be replaced with pod index+1
if [ -n "$FQDN_TEMPLATE" ]; then
  OUTSIDE_NAME=$(echo ${FQDN_TEMPLATE} | sed "s/%/$(($ORD+1))/g")
fi

# domain should be the OUTSIDE_NAME for when it's set
DOMAIN=${SEED_NODE:-$DOMAIN}

# wait for loadbalancer registration and skip the first one
if [ $MYID -gt $OFFSET ]; then
  sleep $START_DELAY
fi

# Values for first startup
WRITE_CONFIGURATION=true
REGISTER_NODE=true
ONDISK_MYID_CONFIG=false
ONDISK_DYN_CONFIG=false

# Check validity of on-disk configuration
if [ -f $MYID_FILE ]; then
  EXISTING_ID="`cat $DATA_DIR/myid`"
  if [[ "$EXISTING_ID" == "$MYID" && -f $STATIC_CONFIG ]]; then
    # If Id is correct and configuration is present under `/data/conf`
      ONDISK_MYID_CONFIG=true
  fi
fi

if [ -f $DYNCONFIG ]; then
  ONDISK_DYN_CONFIG=true
fi

set +e
# Check if envoy is up and running
if [[ -n "$ENVOY_SIDECAR_STATUS" ]]; then
  COUNT=0
  MAXCOUNT=${1:-30}
  HEALTHYSTATUSCODE="200"
  while true; do
    COUNT=$(expr $COUNT + 1)
    SC=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:15000/ready)
    echo "waiting for envoy proxy to come up";
    sleep 1;
    if (( "$SC" == "$HEALTHYSTATUSCODE" || "$MAXCOUNT" == "$COUNT" )); then
      break
    fi
  done
fi
set -e

# Determine if there is an ensemble available to join by checking the service domain
set +e
getent hosts $DOMAIN  # This only performs a dns lookup
if [[ $? -eq 0 ]]; then
  ACTIVE_ENSEMBLE=true
else 
  # If dns lookup of the headless service domain fails, then there is no
  # active ensemble yet, but in certain cases the dns lookup of headless service
  # takes a while to come up even if there is active ensemble
  ACTIVE_ENSEMBLE=false
  declare -i count=20
  while [[ $count -ge 0 ]]
  do
    sleep 2
    ((count=count-1))
    getent hosts $DOMAIN
    if [[ $? -eq 0 ]]; then
      ACTIVE_ENSEMBLE=true
      break
    fi
  done
fi

if [[ "$ONDISK_MYID_CONFIG" == true && "$ONDISK_DYN_CONFIG" == true ]]; then
  # If Configuration is present, we assume, there is no need to write configuration.
    WRITE_CONFIGURATION=false
else
    WRITE_CONFIGURATION=true
fi

if [[ "$ACTIVE_ENSEMBLE" == false ]]; then
  # This is the first node being added to the cluster or headless service not yet available
  REGISTER_NODE=false
else
  # An ensemble exists, check to see if this node is already a member.
  if [[ "$ONDISK_MYID_CONFIG" == false || "$ONDISK_DYN_CONFIG" == false ]]; then
    REGISTER_NODE=true
  else
    REGISTER_NODE=false
  fi
fi

if [[ "$WRITE_CONFIGURATION" == true ]]; then
  echo "Writing myid: $MYID to: $MYID_FILE."
  echo $MYID > $MYID_FILE
  if [[ $MYID -eq $OFFSET && -z "$SEED_NODE" ]]; then
    ROLE=participant
    echo Initial initialization of ordinal 0 pod, creating new config.
    ZKCONFIG=$(zkConfig $OUTSIDE_NAME)
    echo Writing bootstrap configuration with the following config:
    echo $ZKCONFIG
    echo $MYID > $MYID_FILE
    echo "server.${MYID}=${ZKCONFIG}" > $DYNCONFIG
  fi
fi

if [[ "$REGISTER_NODE" == true ]]; then
    ROLE=observer
    ZKURL=$(zkConnectionString)
    ZKCONFIG=$(zkConfig $OUTSIDE_NAME)
    set -e
    echo Registering node and writing local configuration to disk.
    java -Dlog4j.configuration=file:"$LOG4J_CONF" -jar /opt/libs/zu.jar add $ZKURL $MYID  $ZKCONFIG $DYNCONFIG
    set +e
fi

ZOOCFGDIR=/data/conf
export ZOOCFGDIR
echo Copying /conf contents to writable directory, to support Zookeeper dynamic reconfiguration
if [[ ! -d "$ZOOCFGDIR" ]]; then
  mkdir $ZOOCFGDIR
  cp -f /conf/zoo.cfg $ZOOCFGDIR
else
  echo Copying the /conf/zoo.cfg contents except the dynamic config file during restart
  echo -e "$( head -n -1 /conf/zoo.cfg )""\n""$( tail -n 1 "$STATIC_CONFIG" )" > $STATIC_CONFIG
fi
cp -f /conf/log4j.properties $ZOOCFGDIR
cp -f /conf/log4j-quiet.properties $ZOOCFGDIR
cp -f /conf/env.sh $ZOOCFGDIR

if [ -f $DYNCONFIG ]; then
  # Node registered, start server
  echo Starting zookeeper service
  zkServer.sh --config $ZOOCFGDIR start-foreground
else
  echo "Node failed to register!"
  exit 1
fi
