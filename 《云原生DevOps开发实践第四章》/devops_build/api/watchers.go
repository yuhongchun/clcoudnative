package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AddWatchersParam struct {
	UUID        string `mapstructure:"uuid"`         //用户uuid 必填
	ProjectName string `mapstructure:"project_name"` //项目名 必填
	ClusterId   int    `mapstructure:"cluster_id"`   //集群 选填
	Namespace   string `mapstructure:"namespace"`    //命名空间 选填 默认关注此项目所有的环境变更
}

func AddNamespaceWatcher(c *gin.Context) {
	params := &AddWatchersParam{}
	err := c.BindJSON(params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": 1, "err_msg": "参数绑定错误！"})
		return
	}

}
