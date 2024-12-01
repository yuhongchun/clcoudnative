package api

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"devops_build/database"
	"devops_build/database/model"
)

// ListServices
// @Summary 列出某个项目发版路径下的deployment对应的service
// @Produce json
// @Param page query string true "第几页"
// @Param size query string true "一页几个"
// @Success 200 {object} ReturnData{data=PageData}
// @Failure 400 {object} ReturnData{data=PageData}
// @Router /api/nighting-build/service [get]
func ListServices(c *gin.Context) {
	str := c.Query("page")
	s := c.Query("size")

	if str == "" || s == "" {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "Need Page and Size Param"})
		return
	}

	page, err := strconv.Atoi(str)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}

	size, err := strconv.Atoi(s)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}

	if page <= 0 || size <= 0 {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "Invalid Page or Size"})
		return
	}

	devopsdb := database.GetDevopsDb()
	services, err := devopsdb.GetServicesByPageSize(c, page, size)

	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Error: Get Services failed, err: %s", err)
		return
	}

	if len(services) == 0 {
		c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "Empty Data"})
		return
	}

	count, err := devopsdb.GetServiceCount(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Error: Get Service Count failed, err: %s", err)
		return
	}

	c.JSON(http.StatusOK, &ReturnData{Err_code: 1, Err_msg: "ok", Data: PageData{
		Count:    count,
		ListData: services,
	}})
}

// ListServicesByDeploymentId
// @Summary 筛选service
// @Produce json
// @Param depid query int true "deploymentId"
// @Success 200 {object} ReturnData{data=PageData}
// @Failure 400 {object} ReturnData{data=PageData}
// @Router /api/nighting-build/selectservice [get]
func ListServicesByDeploymentId(c *gin.Context) {
	depids := c.Query("depid")

	if depids == "" {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "Need depid Param"})
		return
	}

	depid, err := strconv.Atoi(depids)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}

	db := database.GetDevopsDb()
	services, err := db.GetServicesByDeploymentId(c, depid)
	if err != nil {
		logrus.WithContext(c).Errorf("Err:get service failed,err:%s")
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}

	count, err := db.GetServicesCountByDeploymentId(c, depid)
	if err != nil {
		logrus.WithContext(c).Errorf("Err:get service failed,err:%s")
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, &ReturnData{
		Err_code: 1,
		Err_msg:  "ok",
		Data: PageData{
			Count:    count,
			ListData: services,
		},
	})
	return
}

// @Summary 增加某个项目发版路径下的deployment对应的service
// @Accept application/json
// @Param DeploymentId body string true "对应的deploymentId"
// @Param Content body string true "Content"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/service [put]
func AddServices(c *gin.Context) {
	service := model.Service{}
	err := c.BindJSON(&service)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}
	devopsdb := database.GetDevopsDb()

	// exist, err := devopsdb.IfExistService(c, service)
	// if err != nil {
	// 	nlog.WithContext(c).Errorf("Error: Judge services exist failed, err: %s", err)
	// 	c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
	// 	return
	// }

	// if exist {
	// 	c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "记录已经存在，插入失败"})
	// 	return
	// }

	s, err := devopsdb.InsertIntoService(c, service)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Error: Insert Service failed, err: %s", err)
		return
	}
	c.JSON(http.StatusOK, &ReturnData{Err_code: 1, Err_msg: "ok", Data: s.Id})
}

// @Summary 更新某个项目发版路径下的deployment对应的service
// @Accept application/json
// @Param Id body int true "Id"
// @Param DeploymentId body string true "对应的deploymentId"
// @Param Content body string true "Content"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/service [patch]
func PatchServices(c *gin.Context) {
	service := model.Service{}
	err := c.BindJSON(&service)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}
	devopsdb := database.GetDevopsDb()
	result, err := devopsdb.UpdateService(c, service)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Patch Service failed err:", err)
		return
	}

	if result {
		c.JSON(http.StatusOK, &ReturnData{Err_code: 1, Err_msg: "ok"})
		return
	} else {
		c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "update failed"})
		return
	}
}

// DeleteService
// @Summary 删除一个项目
// @Param serviceid query string true "ID"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/service [delete]
func DeleteService(c *gin.Context) {
	strid := c.Query("serviceid")
	serviceid, err := strconv.Atoi(strid)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}
	devopsdb := database.GetDevopsDb()
	_, err = devopsdb.DeleteServiceById(c, serviceid)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Delete Service failed err:", err)
		return
	}

	c.JSON(http.StatusOK, &ReturnData{
		Err_code: 1,
		Err_msg:  "ok",
	})
}
