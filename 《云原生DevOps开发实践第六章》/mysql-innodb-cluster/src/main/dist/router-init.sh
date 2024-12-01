#!/usr/bin/env bash

./bootstrap
ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime

mv mysql-router-8.0.21-el7-x86_64 mysql-router

# Create MySQL InnoDB Cluster
echo -e "Create MySQL InnoDB Cluster"


# Select cluster_name from mysql_innodb_cluster_metadata.cluster
mysql -hmysql-0-node.${FRAMEWORK_NAME}.autoip.dcos.thisdcos.directory -uroot -p${MYSQL_ROOT_PASSWORD} -e "select cluster_name  from mysql_innodb_cluster_metadata.clusters;" | grep ${CLUSTERNAME}

if [[ $? -ne 0 ]] && [[ ${POD_INSTANCE_INDEX} = 0 ]];then
cat << EOF > init_cluster.js
shell.connect('root@mysql-0-node.${FRAMEWORK_NAME}.autoip.dcos.thisdcos.directory:3306', '${MYSQL_ROOT_PASSWORD}')
var cluster = dba.createCluster('${CLUSTERNAME}', {adoptFromGR: true});
EOF
mysqlsh --no-password --js --file=init_cluster.js
else
    echo -e "mysql cluster has been established or Not router-0-node"
fi

echo -e "Generate mysql-router configuration file"
echo -e "POD_INSTANCE_INDEX:"${POD_INSTANCE_INDEX}

# If the specified second node (read-only) cannot be operated, it will automatically try to connect to the primary node, Add HealthCheck First

for n in {0..2};do
    nodeSelf=`mysql -hmysql-$n-node.${FRAMEWORK_NAME}.autoip.dcos.thisdcos.directory -uroot -p123456 -e "SELECT ta.* ,tb.MEMBER_HOST,tb.MEMBER_PORT,tb.MEMBER_STATE FROM performance_schema.global_status ta,performance_schema.replication_group_members tb  WHERE ta.VARIABLE_NAME='group_replication_primary_member' and ta.VARIABLE_VALUE=tb.MEMBER_ID;" | grep "autoip.dcos.thisdcos" | awk '{print $3}'`
    echo $nodeSelf
    if [ ! -n $nodeSelf ]; then
        echo "MIC Master Node:"$nodeSelf
    fi
done
echo -e "run mysql-router"
./mysql-router/bin/mysqlrouter --bootstrap root:${MYSQL_ROOT_PASSWORD}@$nodeSelf:3306 --user=root
./mysql-router/bin/mysqlrouter -c /mnt/mesos/sandbox/mysql-router/mysqlrouter.conf
