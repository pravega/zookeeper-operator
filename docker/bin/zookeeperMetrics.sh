#!/usr/bin/env bash
set -ex

source /conf/env.sh

echo mntr | nc localhost $CLIENT_PORT >& 1
