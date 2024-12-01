#!/usr/bin/env bash
# cron job
MESOS_SANDBOX=/mnt/mesos/sandbox
MYSQL_ROOT_PASSWORD=123456
nodeSelf="localhost"
export PATH=${MESOS_SANDBOX}/percona-xtrabackup-2.4.21-Linux-x86_64.glibc2.12/bin:$PATH


function mysqlExec () {
    local mysqlHost=$1
    local options=$2
    local execSql=${@: 3}

    mysql --connect-timeout 3 -h${mysqlHost} -uroot -p${MYSQL_ROOT_PASSWORD} ${options} -e "$execSql" 2>/dev/null
}

# CleanUp Backup Data
find ${MESOS_SANDBOX}/BKP_DATA  -mtime +10 -name "*" -exec rm -rf {} \;

PRIMARY=`mysqlExec ${nodeSelf} -sN "SHOW STATUS LIKE 'group_replication_primary_member';" | awk '{print $2}'`
UUID=`mysqlExec ${nodeSelf} -sN "SHOW GLOBAL VARIABLES LIKE 'server_uuid';" | awk '{print $2}'`
echo -e "PRIMARY:"$PRIMARY
echo -e "UUID":$UUID
if [[ $PRIMARY = $UUID ]];then
    echo -e "localhost UUID:"$UUID
    echo "start local backup.." > local_bkp.log
        backuppath=`date "+%Y%m%d%H%M%S"-full`
        mkdir -p BKP_DATA/${FRAMEWORK_NAME}/conf/
        mkdir -p BKP_DATA/${FRAMEWORK_NAME}/data/${backuppath}
        echo "${backuppath}" >${MESOS_SANDBOX}/BKP_DATA/${FRAMEWORK_NAME}/conf/inc.txt

        innobackupex --user=root --password=123456 --port=3306 --socket=/var/lib/mysql/mysql.sock --datadir=${MESOS_SANDBOX}/MYSQL_DATA BKP_DATA/${FRAMEWORK_NAME}/data/${backuppath}/  >> local_bkp.log
        echo "end local backup.." >> local_bkp.log
else
    echo -e "primary is not localhost,skip local backup"
fi
sleep 10s
