package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"devops_build/database"
	"github.com/sirupsen/logrus"
)

// ListNspByClusterId
// @Summary 筛选命名空间
// @Produce json
// @Param clusterid query int true "集群Id"
// @Success 200 {object} ReturnData{data=PageData}
// @Failure 400 {object} ReturnData{data=PageData}
// @Router /api/nighting-build/namespace [get]
func ListNspByClusterId(c *gin.Context) {
	ids := c.Query("clusterid")
	if ids == "" {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "Need clusterid Param"})
		return
	}
	id, err := strconv.Atoi(ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}
	db := database.GetDevopsDb()
	nsps, err := db.GetNspByClusterId(c, id)
	if err != nil {
		logrus.WithContext(c).Error(err)
		c.JSON(http.StatusBadRequest, &ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, &ReturnData{
		Err_code: 1,
		Err_msg:  "ok",
		Data: PageData{
			Count:    len(nsps),
			ListData: nsps,
		},
	})
}

// ListNspByClusterId
// @Summary 筛选命名空间
// @Produce json
// @Param namespace_id query int true "集群Id"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/single_namespace [get]
func GetNamespaceById(c *gin.Context) {
	id, err := strconv.Atoi(c.Request.URL.Query().Get("namespace_id"))
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": 0, "err_msg": "参数错误"})
		return
	}
	devopsdb := database.GetDevopsDb()
	res, err := devopsdb.GetNamespaceById(c, id)
	if err != nil {
		c.JSON(200, gin.H{"err_code": 0, "err_msg": "查询错误！"})
	}
	c.JSON(200, gin.H{"err_code": 1, "err_msg": "ok", "data": res})
}
