package api

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"devops_build/database"
	"devops_build/database/model"
)

// ListRoutes
// @Summary 列出所有项目对应的发版路径route
// @Produce json
// @Param page query int true "第几页"
// @Param size query int true "页大小"
// @Success 200 {object} ReturnData{data=PageData}
// @Failure 400 {object} ReturnData{data=PageData}
// @Router /api/nighting-build/route [get]
func ListRoutes(c *gin.Context) {
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
	routes, err := devopsdb.GetRoutesByPageSize(c, page, size)

	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Error: Get routes failed, err: %s", err)
		return
	}

	if len(routes) == 0 {
		c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "Empty Data"})
		return
	}

	count, err := devopsdb.GetRouteCount(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Error: Get Route Count failed, err: %s", err)
		return
	}

	c.JSON(http.StatusOK, &ReturnData{Err_code: 1, Err_msg: "ok", Data: PageData{
		Count:    count,
		ListData: routes,
	}})
}

// ListRoutesByProIdAndNspIdAndPageSize
// @Summary 筛选route
// @Produce json
// @Param page query int  true "第几页"
// @Param size query int true "页大小"
// @Param proid query int true "项目id"
// @Param nspid query int true "命名空间 可选 不填则只以projectid为条件进行查询"
// @Success 200 {object} ReturnData{data=PageData}
// @Failure 400 {object} ReturnData{data=PageData}
// @Router /api/nighting-build/selectroute [get]
func ListRoutesByProIdAndNspIdAndPageSize(c *gin.Context) {
	str := c.Query("page")
	s := c.Query("size")
	proidstr := c.Query("proid")
	nspidstr := c.Query("nspid")

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

	if proidstr == "" {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  "proid and nspid can't be empty",
		})
		return
	}
	if nspidstr == "" {
		nspidstr = "0"
	}
	proid, err := strconv.Atoi(proidstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}

	nspid, err := strconv.Atoi(nspidstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}

	db := database.GetDevopsDb()
	routes, err := db.GetRoutesByProIdAndNspIdAndPageSize(c, proid, nspid, page, size)
	if err != nil {
		logrus.WithContext(c).Errorf("Error: Get Routes failed, err: %s", err)
		c.JSON(http.StatusBadRequest, &ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}
	for _, r := range routes {
		nsp, err := db.GetNamespaceById(c, r.NamespaceId)
		if err != nil {
			continue
		}
		r.NspName = nsp.Name
	}

	count, err := db.GetRoutesCountByProIdAndNspId(c, proid, nspid)
	if err != nil {
		logrus.WithContext(c).Errorf("Error: Get Routes Count failed, err: %s", err)
		c.JSON(http.StatusBadRequest, &ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ReturnData{
		Err_code: 0,
		Err_msg:  "ok",
		Data: PageData{
			Count:    count,
			ListData: routes,
		},
	})
}

// @Summary 增加某个项目对应的发版路径route
// @Accept application/json
// @param ProjectId  body int true "ProjectId"
// @param RefRep  body string true "该路径发版的分支 使用正则表达式"
// @param Channel body string true "发版路径"
// @param ClusterId  body int true "集群Id"
// @param NspName body string true "NspName"
// @param Enabled body bool true "是否弃用"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/route [put]
func AddRoutes(c *gin.Context) {
	route := &model.Route{}
	err := c.BindJSON(route)
	route.CreateTime = time.Now()
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}

	devopsdb := database.GetDevopsDb()
	// exist, err := devopsdb.IfExistRoute(c, *route)

	// if err != nil {
	// 	nlog.WithContext(c).Errorf("Error: Judge routes exist failed, err: %s", err)
	// 	c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
	// 	return
	// }

	// if exist {
	// 	c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "记录已经存在，插入失败"})
	// 	return
	// }
	var r *model.Route
	if r, err = devopsdb.InsertIntoRoutes(c, *route); err != nil {
		logrus.WithContext(c).Errorf("Error: Insert route failed, err: %s", err)
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, &ReturnData{Err_code: 1, Err_msg: "ok", Data: r.Id})
}

// @Summary 更新某个项目对应的发版路径route
// @Accept application/json
// @param Id  body int true "Id"
// @param ProjectId  body int true "ProjectId"
// @param RefRep  body string true "该路径发版的分支 使用正则表达式"
// @param Channel body string true "弃用 发版路径"
// @param ClusterId  body int true "集群Id"
// @param NspName body string true "NspName"
// @param Enabled body bool true "是否生效 目前都是true"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/route [patch]
func PatchRoutes(c *gin.Context) {
	route := model.Route{}
	err := c.BindJSON(&route)
	route.UpdateTime = time.Now()
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}
	devopsdb := database.GetDevopsDb()
	route.UpdateTime = time.Now()
	result, err := devopsdb.UpdateRoute(c, route)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Patch Route failed err:", err)
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

// DeleteRoute
// @Summary 删除一个路由
// @Param routeid query string true "路由ID"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/route [delete]
func DeleteRoute(c *gin.Context) {
	strid := c.Query("routeid")
	routeid, err := strconv.Atoi(strid)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}
	devopsdb := database.GetDevopsDb()
	_, err = devopsdb.DeleteRouteById(c, routeid)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Delete Route failed err:", err)
		return
	}

	c.JSON(http.StatusOK, &ReturnData{
		Err_code: 1,
		Err_msg:  "ok",
	})
}
