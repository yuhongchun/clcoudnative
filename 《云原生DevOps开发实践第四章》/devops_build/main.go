package main

import "devops_build/cmd"

// @title devops-build
// @version 1.0
// @description DevOps发版平台
// @termsOfService http://swagger.io/terms/

// @contact.name
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:5001
// @BasePath
func main() {
	cmd.Execute()
}
