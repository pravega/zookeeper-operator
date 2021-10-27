#!/bin/sh
set -xe
LOG_PATH=/opt/zookeeper/logs/zookeeper.log*
LOG_RECOVER_PATH=/opt/zookeeper/logs/recovery.log
usage() {
    cat <<EOF
usage: ${0} [OPTIONS]

The following flags are required.
       --type, -t             service procedure type(precheck, recover, postcheck).
EOF
    exit 1
}

precheck() {
  echo "zookeeper precheck called at " $(date) >>  ${LOG_RECOVER_PATH}
  if grep -q "Peer state changed" ${LOG_PATH}
  then
    echo "zookeeper has ongoing sync task" >>  ${LOG_RECOVER_PATH}
    if ( grep "Peer state changed" ${LOG_PATH} | sort -k 2 | tail -1| grep -q "following - broadcast" \
    || grep "Peer state changed" ${LOG_PATH} | sort -k 2 | tail -1| grep -q  "leading - broadcast" ) \
    && ! grep -q "digest mismatch" ${LOG_PATH}
    then
      echo "sync succeeded" >>  ${LOG_RECOVER_PATH}
      exit 0
    else
      echo "sync did not succeed" >>  ${LOG_RECOVER_PATH}
      exit 1
     fi
  else
    echo "zookeeper has no ongoing sync task" >>  ${LOG_RECOVER_PATH}
    if ! grep -q "digest mismatch" ${LOG_PATH}
    then
      echo "no ongoing sync, no mismatch" >>  ${LOG_RECOVER_PATH}
      exit 0
    else
      echo "digest mismatch!" >>  ${LOG_RECOVER_PATH}
      exit 1
    fi
  fi
}

postcheck() {
  echo "zookeeper postcheck called at " $(date) >>  ${LOG_RECOVER_PATH}
  if grep -q "Peer state changed" ${LOG_PATH}
  then
    echo "zookeeper has ongoing sync task" >>  ${LOG_RECOVER_PATH}
    if ( grep "Peer state changed" ${LOG_PATH} | sort -k 2 | tail -1| grep -q "following - broadcast" \
    || grep "Peer state changed" ${LOG_PATH} | sort -k 2 | tail -1| grep -q  "leading - broadcast" ) \
    && ! grep -q "digest mismatch" ${LOG_PATH}
    then
      echo "sync succeeded" >>  ${LOG_RECOVER_PATH}
      exit 0
    else
      echo "sync did not succeed" >>  ${LOG_RECOVER_PATH}
      exit 1
     fi
  else
    echo "zookeeper has no ongoing sync task" >>  ${LOG_RECOVER_PATH}
    if ! grep -q "digest mismatch" ${LOG_PATH}
    then
      echo "no ongoing sync, no mismatch" >>  ${LOG_RECOVER_PATH}
      exit 0
    else
      echo "digest mismatch!" >>  ${LOG_RECOVER_PATH}
      exit 1
    fi
  fi
}


recovery() {
  echo "zookeeper recovery called at " $(date) >>  ${LOG_RECOVER_PATH}
  exit 0
}

while [ $# -gt 0 ]; do
    case ${1} in
        -t|--type)
            type="$2"
            shift
            shift
            ;;
        *)
            usage
            shift
            ;;
    esac
done

case ${type} in
    precheck)
        precheck
        ;;
    postcheck)
        postcheck
        ;;
    recover)
        recovery
        ;;
    *)
        usage
        ;;
esac