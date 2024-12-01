#!/usr/bin/env bash
# backup All or Singel Databases
if [ ${BACKUP_DATABASENAME}=="all" ];then
  mysql -hmaster.${FRAMEWORK_NAME}.l4lb.thisdcos.directory -uroot -p${MYSQL_ROOT_PASSWORD} -P 6446 -e "show databases;" | grep -Ev "Database|mysql|sys|information_schema|performance_schema|mysql_innodb_cluster_metadata" > databasename.log
  DATABASENAME=`cat databasename.log | xargs echo -n`
  echo $DATABASENAME
  backupname=`date "+%Y%m%d%H%M%S"-full`

  mysqldump -hmaster.${FRAMEWORK_NAME}.l4lb.thisdcos.directory -uroot -p${MYSQL_ROOT_PASSWORD} -P 6446 --databases $DATABASENAME --set-gtid-purged=OFF >  $backupname.sql
else
  mysqldump -hmaster.${FRAMEWORK_NAME}.l4lb.thisdcos.directory -uroot -p${MYSQL_ROOT_PASSWORD} -P 6446 --databases ${BACKUP_DATABASENAME} --set-gtid-purged=OFF > ${BACKUP_DATABASENAME}.sql
fi

if [[ $? -eq 0 ]];then
  echo -e "backup status:sucess!"
else
  echo -e "backup status:fail"
fi