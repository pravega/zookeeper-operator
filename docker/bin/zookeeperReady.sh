#!/usr/bin/env bash
set -ex

source /conf/env.sh
source /usr/local/bin/zookeeperFunctions.sh

OK=$(echo ruok | nc 127.0.0.1 $CLIENT_PORT)

# Check to see if zookeeper service answers
if [[ "$OK" == "imok" ]]; then

  # Check to see if zookeeper service for this node is a participant
  ZKURL=$(zkConnectionString)
  MYID=`cat $MYID_FILE`
  ROLE=`java -jar -Dlog4j.configuration=file:"$LOG4J_CONF" /root/zu.jar get-role $ZKURL $MYID`

  if [[ "$ROLE" == "participant" ]]; then
    echo "Zookeeper service is available and an active participant"
    exit 0

  elif [[ "$ROLE" == "observer" ]]; then
    echo "Zookeeper service is ready to be upgraded from observer to participant."
    ROLE=participant
    ZKCONFIG=$(zkConfig)
    java -jar -Dlog4j.configuration=file:"$LOG4J_CONF" /root/zu.jar remove $ZKURL $MYID
    sleep 1
    java -jar -Dlog4j.configuration=file:"$LOG4J_CONF" /root/zu.jar add $ZKURL $MYID $ZKCONFIG
    exit 1

  else
    echo "Something has gone wrong. Unable to determinal zookeeper role."
    exit 1

  fi

else
  echo "Zookeeper service is not available for requests"
  exit 1

fi
