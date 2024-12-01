package main

import (
	"context"

	"devops_release/config"
	"devops_release/database"
	"devops_release/database/relational/postgres"
	"devops_release/internal/buildv2"
)

func main() {
	config.Setup("config/settings.yaml")
	postgres.PostgresUtils.SetUp(postgres.Options{
		Dsn: config.PostgresConfig.Dsn,
	})
	devopsdb := database.GetDevopsDb()
	if devopsdb == nil {
		panic("devopsdb 为空！")
	}
	buildv2.RestartWithConfig(context.Background(), "weiban-dev-new", "weiban-tenant-1-dev", "activity-management")
}
