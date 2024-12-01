package database

import (
	"devops_release/database/relational"
	"devops_release/database/relational/postgres"
)

//获取当前实现的数据库实例
func GetDevopsDb() relational.DevopsDb {
	return postgres.DevopsDb
}
