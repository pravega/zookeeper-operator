#!/bin/bash

echo "--------------- Check required parameters ---------------"
_req_envs="
Required parameters: \n
BACKUPDIR: $BACKUPDIR \n
ZOOKEEPERDATADIR: $ZOOKEEPERDATADIR \n
BACKUPS_TO_KEEP: $BACKUPS_TO_KEEP \n
"
echo -e "$_req_envs"

if [ -z "$BACKUPDIR" ] || [ -z "$ZOOKEEPERDATADIR" ] || [ -z "$BACKUPS_TO_KEEP" ]; then
  echo -e "Some required env variables aren't defined.\n"
  exit 1
fi

# Create backup directory if absent.
# ----------------------------------
echo "--------------- Check backup/tmp dirs ---------------"
if [ ! -d "$BACKUPDIR" ] && [ ! -e "$BACKUPDIR" ]; then
    mkdir -p "$BACKUPDIR"
else
    printf "Backup directory $BACKUPDIR is existed.\n"
fi

# TO DO: provide additional check of zookeeper health

# Backup and create tar archive
# ------------------------------
echo "--------------- Backup ---------------"
TIMESTAMP=$( date +"%Y%m%d%H%M%S" )
# Include the timestamp in the filename
FILENAME="$BACKUPDIR/zookeeper-$TIMESTAMP.tar.gz"

tar -zcvf $FILENAME -P $ZOOKEEPERDATADIR > /dev/null 2>&1
RC=$?

if [ $RC -ne 0 ]; then
    printf "Error generating tar archive.\n"
    exit 1
else
    printf "Successfully created a backup tar file.\n"
fi


# Cleanup old backups
# -------------------
echo "--------------- Cleanup ---------------"
echo "List of backups:"
cd $BACKUPDIR

BACKUPS_AMOUNT=`find . -path "*/zookeeper-*.tar.gz" -type f -printf "\n%AD %AT %p" | wc -l`
TO_DELETE=$(( $BACKUPS_AMOUNT - $BACKUPS_TO_KEEP ))
if [ $TO_DELETE -gt 0 ] ; then
    echo "Keeping only $BACKUPS_TO_KEEP full backups"
    ls -t | tail -n -$TO_DELETE | xargs -d '\n' rm -rf
else
    echo "There are less backups than required, nothing to delete."
fi

# Cleanup old backups
# -------------------
echo "--------------- Current backups ---------------"
ls -lt
