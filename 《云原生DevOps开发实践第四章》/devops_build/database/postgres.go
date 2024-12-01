package database

import (
	"devops_build/database/relational"
	"devops_build/database/relational/postgres"
)

func GetDevopsDb() relational.DevopsDb {
	return postgres.DevopsDb
}
