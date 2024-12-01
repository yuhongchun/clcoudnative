#!/usr/bin/env bash
set -m

declare nodeSelf="localhost"

function mysqlExec () {
    local mysqlHost=$1
    local options=$2
    local execSql=${@: 3}

    mysql --connect-timeout 3 -h${mysqlHost} -uroot -p${MYSQL_ROOT_PASSWORD} ${options} -e "$execSql" 2>/dev/null
}


./bootstrap
ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime
serverID=$(( ${POD_INSTANCE_INDEX} + 1 ))
echo -e "MySQL Server ID: $serverID"
#hostIP=${MESOS_CONTAINER_IP}


# Calculate the mysql pool buffer size
#MYSQL_INNODB_BUFFER_POOL_SIZE=$(echo "${NODE_MEM} * ${MYSQL_INNODB_BUFFER_POOL_SIZE_RATIO}"|bc)
MYSQL_INNODB_BUFFER_POOL_SIZE=`echo | awk "{print $NODE_MEM * $MYSQL_INNODB_BUFFER_POOL_SIZE_RATIO}"`

# my.cnf for related parameter tuning
cp -v ./config-templates/my.cnf  ./my.cnf

sed -i "s@serverid@$serverID@g" ./my.cnf
sed -i "s@INDEX@${POD_INSTANCE_INDEX}@g" ./my.cnf
sed -i "s@FRAMEWORK_NAME@${FRAMEWORK_NAME}@g" ./my.cnf
sed -i "s@maxconn@${MYSQL_MAX_CONNECTIONS}@g" ./my.cnf
sed -i "s@maxuserconn@${MYSQL_MAX_USER_CONNECTIONS}@g" ./my.cnf
sed -i "s@buffersize@${MYSQL_INNODB_BUFFER_POOL_SIZE}@g" ./my.cnf
sed -i "s@timehour@${MYSQL_TIME_HOUR}@g" ./my.cnf
sed -i "s@expiredays@${MYSQL_EXPIRE_LOGS_DAYS}@g" ./my.cnf
sed -i "s@timeoutconn@${MYSQL_TIMEOUT}@g" ./my.cnf
sed -i "s@maxerr@${MYSQL_CONNECT_ERRORS}@g" ./my.cnf

echo "report_host=mysql-${POD_INSTANCE_INDEX}-node.${FRAMEWORK_NAME}.autoip.dcos.thisdcos.directory" >> ./my.cnf
rm -rf /etc/my.cnf
cp ./my.cnf /etc/my.cnf
chown -R mysql:mysql /mnt && chmod -R 777 /tmp && chmod -R 777 /mnt

# start mysqld as background job 1
pStatus=$?
/entrypoint.sh mysqld --log-error &

if [[ ${pStatus} -ne 0 ]]; then
    echo -e "Failed to start mysqld, process exit code: $pStatus"
    exit ${pStatus}
else
    until mysqlExec ${nodeSelf} -sN "SELECT 1"; do sleep 3; done
fi

#sleep 30
echo -e "Post-flight status: START"
#rtnCode=0


# update the password of user root
# mysql -uroot -p -e "update user set password=password('123456');"


# install group_replication plugins

mysqlExec ${nodeSelf} -sN "\
SET SQL_LOG_BIN=0;\
GRANT ALL PRIVILEGES ON *.* TO 'root'@'%' IDENTIFIED BY '${MYSQL_ROOT_PASSWORD}' WITH GRANT OPTION;\
GRANT ALL PRIVILEGES ON mysql_innodb_cluster_metadata.* TO root@'%' WITH GRANT OPTION;\
GRANT RELOAD, SHUTDOWN, PROCESS, FILE, SUPER, REPLICATION SLAVE, REPLICATION CLIENT, CREATE USER ON *.* TO root@'%' WITH GRANT OPTION;\
FLUSH PRIVILEGES;\
SET SQL_LOG_BIN=1;"

# Install MySQL Group Replication MGR
echo -e "install MySQL Group Replication MGR"

if [[ ${POD_INSTANCE_INDEX} -eq 0 ]]; then
    echo "POD":${POD_INSTANCE_INDEX}
    mysqlExec ${nodeSelf} -sN "\
    SET SQL_LOG_BIN=0;\
    CREATE USER rpl_user@'%';\
    GRANT REPLICATION SLAVE ON *.* TO rpl_user@'%' IDENTIFIED BY 'rpl_pass';\
    SET SQL_LOG_BIN=1;\
    CHANGE MASTER TO MASTER_USER='rpl_user', MASTER_PASSWORD='rpl_pass' FOR CHANNEL 'group_replication_recovery';\
    INSTALL PLUGIN group_replication SONAME 'group_replication.so';\
    set global group_replication_bootstrap_group = ON;\
    START GROUP_REPLICATION;\
    set global group_replication_bootstrap_group = OFF;"
else
    echo "POD":${POD_INSTANCE_INDEX}
    mysqlExec ${nodeSelf} -sN "\
    SET SQL_LOG_BIN=0;\
    CREATE USER rpl_user@'%';\
    GRANT REPLICATION SLAVE ON *.* TO rpl_user@'%' IDENTIFIED BY 'rpl_pass';\
    SET SQL_LOG_BIN=1;\
    CHANGE MASTER TO MASTER_USER='rpl_user', MASTER_PASSWORD='rpl_pass' FOR CHANNEL 'group_replication_recovery';\
    INSTALL PLUGIN group_replication SONAME 'group_replication.so';\
    set global group_replication_allow_local_disjoint_gtids_join=ON;\
    START GROUP_REPLICATION;"
fi


#Only print failure status without exiting, to prevent scenarios where Pod restart cannot start
if [[ $? -ne 0 ]]; then
    echo -e "install Group Replication MGR status: FAIL"
    source ./config-templates/rejoin-instance.sh
    #exit $rtnCode
    #fi
else
    echo -e "install Group Replication MRG status: sucesses"

fi


#bring background job 1 to foreground
echo -e "Bring mysqld process to foreground"
fg 1
