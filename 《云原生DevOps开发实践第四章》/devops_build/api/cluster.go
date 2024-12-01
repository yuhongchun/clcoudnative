package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"devops_build/database"
	"github.com/sirupsen/logrus"
)

// ListCluster
// @Produce json
// @Success 200 {object} ReturnData{data=PageData}
// @Failure 400 {object} ReturnData{data=PageData}
// @Router /api/nighting-build/cluster [get]
func ListCluster(c *gin.Context) {
	db := database.GetDevopsDb()
	clusters, err := db.GetCluster(c)
	if err != nil {
		logrus.WithContext(c).Errorf("Err:get cluser error:", err)
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, ReturnData{
		Err_code: 1,
		Err_msg:  "ok",
		Data: PageData{
			Count:    len(clusters),
			ListData: clusters,
		},
	})
	return
}
