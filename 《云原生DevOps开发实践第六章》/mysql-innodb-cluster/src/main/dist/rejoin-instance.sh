#!/usr/bin/env bash

function mysqlExec () {
    local mysqlHost=$1
    local options=$2
    local execSql=${@: 3}

    mysql --connect-timeout 3 -h${mysqlHost} -uroot -p${MYSQL_ROOT_PASSWORD} ${options} -e "$execSql" 2>/dev/null
}


for n in {0..2};do
  nodeSelf=mysql-$n-node.${FRAMEWORK_NAME}.autoip.dcos.thisdcos.directory
  PRIMARY=`mysqlExec ${nodeSelf} -sN "SHOW STATUS LIKE 'group_replication_primary_member';" | awk '{print $2}'`
  UUID=`mysqlExec ${nodeSelf} -sN "SHOW GLOBAL VARIABLES LIKE 'server_uuid';" | awk '{print $2}'`
  echo -e "PRIMARY:"$PRIMARY
  echo -e "UUID":$UUID
  if [[ $PRIMARY = $UUID ]];then
    MASTERNODE=mysql-$n-node.${FRAMEWORK_NAME}.autoip.dcos.thisdcos.directory
    echo "MASTERNODE:"$MASTERNODE
    break
  fi
sleep 1
done


cat << EOF > rejoin_cluster.js
shell.connect('$MASTERNODE', '${MYSQL_ROOT_PASSWORD}')
var cluster=dba.getCluster()
cluster.removeInstance('root@mysql-${POD_INSTANCE_INDEX}-node.${FRAMEWORK_NAME}.autoip.dcos.thisdcos.directory:3306',{'force':'true'})
cluster.addInstance('root@mysql-${POD_INSTANCE_INDEX}-node.${FRAMEWORK_NAME}.autoip.dcos.thisdcos.directory:3306', {'password': '123456'})
EOF

mysqlsh --no-password --js --file=rejoin_cluster.js