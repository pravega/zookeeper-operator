#!/usr/bin/env bash
set -ex

source /conf/env.sh

OK=$(echo ruok | nc 127.0.0.1 $CLIENT_PORT)

# Check to see if zookeeper service answers
if [[ "$OK" == "imok" ]]; then
  exit 0

else
  exit 1

fi
