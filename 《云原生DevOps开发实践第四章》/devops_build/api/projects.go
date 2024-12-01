package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"devops_build/database"
	"devops_build/database/model"
	m "devops_build/database/model"
	"devops_build/internal/gitlabapi"
	"devops_build/internal/project"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type FuzzyData struct {
	Fuzzystr string `db:"fuzzystr"`
}

// ListProjects
// @Summary 列出所有项目
// @Produce json
// @Param page query string true "第几页"
// @Param size query int true "页大小"
// @Success 200 {object} ReturnData{data=PageData}
// @Failure 400 {object} ReturnData{data=PageData}
// @Router /api/nighting-build/project [get]
func ListProjects(c *gin.Context) {
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
	projects, err := devopsdb.GetProjectsByPageSize(c, page, size)

	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Error: Get projects failed, err: %s", err)
		return
	}

	if len(projects) == 0 {
		c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "Empty Data"})
		return
	}

	count, err := devopsdb.GetProjectCount(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Error: Get Project Count failed, err: %s", err)
		return
	}

	c.JSON(http.StatusOK, &ReturnData{Err_code: 1, Err_msg: "ok", Data: PageData{
		Count:    count,
		ListData: projects,
	}})
}

// ListProjectsByFuzzyFind
// @Summary 模糊查询得出所需的项目
// @Produce json
// @Param fuzzystr query string true "输入的项目查询字符串"
// @Success 200 {object} ReturnData
// @Router /api/nighting-build/projectname [get]
func ListProjectsFuzzy(c *gin.Context) {
	fstr := c.Query("fuzzystr")

	if fstr == "" {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "Need fuzzystr Param"})
		return
	}

	devopsdb := database.GetDevopsDb()
	projects, err := devopsdb.GetProjectsByFuzzyFind(c, fstr)

	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Error: Get projects failed, err: %s", err)
		return
	}

	if len(projects) == 0 {
		c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "Empty Data"})
		return
	}

	//	projects, err := devopsdb.GetProjectsByFuzzyFind(c,fstr)

	c.JSON(http.StatusOK, &ReturnData{Err_code: 1, Err_msg: "ok", Data: projects})
}

// @Summary 增加项目
// @Accept application/json
// @Param project_name body string false "ProjectName"
// @Param repo_type body string true "仓库类型  应该有个下拉框 目前只有gitlab一种"
// @Param repo_url body string true "仓库地址 有此参数的话会自动同步其他git相关的信息（projectname,gitlab_id,topic,descript等）"
// @Param tags body string true "项目类型 type=ops or tenant or center"
// @Param enabled body bool true "Enabled"
// @Param project_token body string true "项目令牌 目前都是 cicd??????"
// @Param descript body string false "Descript"
// @Param topic body string  false "Topic"
// @Param gitlab_id body string false "GitlabId"
// @Param group body string false "Group"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/project [put]
func AddProjects(c *gin.Context) {
	//todo
	var project = m.Project{}
	devopsdb := database.GetDevopsDb()
	err := c.ShouldBindJSON(&project)
	fmt.Println(project)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}
	if project.RepoType == "gitlab" {
		project.RepoUrl = strings.TrimSpace(project.RepoUrl)
		repoUrlSplit := strings.Split(strings.TrimSuffix(project.RepoUrl, ".git"), "http://10.1.0.200:30808/")
		fmt.Println("repoUrlSplit", repoUrlSplit)
		if len(repoUrlSplit) != 2 {
			c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "仓库地址不合法！"})
			return
		}
		groupPro := strings.Trim(repoUrlSplit[1], "/")
		groupProSplit := strings.Split(groupPro, "/")
		fmt.Println("groupProSplit", groupProSplit)
		if len(groupProSplit) < 2 {
			c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "仓库地址不合法！"})
			return
		}

		gitlabPro, err := gitlabapi.GitlabApiObj.GetSingleProject(c, groupProSplit[0], groupProSplit[1])
		fmt.Println("gitlabPro", gitlabPro)
		if err != nil {
			logrus.Error(err)
			c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "从gitlab获取项目信息出错！"})
			return
		}
		project.GitlabId = gitlabPro.ID
		project.Group = groupProSplit[0]
		if len(project.ProjectName) == 0 {
			project.ProjectName = gitlabPro.Name
		}
		project.Enabled = true
		topic := ""
		for _, t := range gitlabPro.Topics {
			topic = topic + t + ","
		}
		topic = strings.Trim(topic, ",")
		project.Topic = topic
		project.Descript = gitlabPro.Description
		project.ProjectToken = "cicd??????"

		fmt.Println(project)
		fmt.Println("ID:", project.GitlabId)
		if project.GitlabId == 0 {
			c.JSON(200, gin.H{"err_code": 0, "err_msg": "从gitlab获取项目失败"})
			return
		}
		exist, err := devopsdb.IfExistProject(c, project)
		if err != nil {
			logrus.WithContext(c).Errorf("Error: Judge Project exist failed,err: %s", err)
			c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
			return
		}

		if exist {
			c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "记录已经存在，插入失败"})
			return
		}
	}

	if len(project.Tags) == 0 {
		project.Tags = "type=tenant,"
	}
	pro, err := devopsdb.InsertIntoProjects(c, project)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Error: Insert Project failed, err: %s", err)
		return
	}
	//初始化配置项

	// apolloConfig := m.Config{
	// 	ProjectId:        pro.Id,
	// 	ConfigName: project.ProjectName,
	// 	FileName:         "settings.yaml",
	// 	ConfigmapName:    project.ProjectName,
	// 	DeploymentId: ,
	// }
	// _, err = devopsdb.InsertIntoApolloConfig(c, apolloConfig)
	// if err != nil {
	// 	nlog.Error("初始化配置信息失败！")
	// }
	//初始化webhook
	//添加mr gitlab_automatic hook
	//gitlabAutomaticHook := "https://automatic.ops-dev.miemie.la/we-work/webhook"
	//gitlabAutomaticHookEvent := true
	//err = gitlabsdk.AddWebHook(c, project.GitlabId, &gitlab.AddProjectHookOptions{
	//	MergeRequestsEvents: &gitlabAutomaticHookEvent,
	//	URL:                 &gitlabAutomaticHook,
	//})
	//if err != nil {
	//	logrus.Error("添加webhook失败！", err)
	//}
	//
	////添加dev/test cicd webhook
	//gitlabNighingBuildHook := "https://devopsyyds.ops-dev.miemie.la/api/nighting-build/gitlab_callback"
	//gitlabNightingBuildHookEvent := true
	//gitlabNightingBuildToken := "cicd??????"
	//err = gitlabsdk.AddWebHook(c, project.GitlabId, &gitlab.AddProjectHookOptions{
	//	PipelineEvents: &gitlabNightingBuildHookEvent,
	//	URL:            &gitlabNighingBuildHook,
	//	Token:          &gitlabNightingBuildToken,
	//})
	//if err != nil {
	//	logrus.Error("添加webhook失败！", err)
	//}
	//
	c.JSON(http.StatusOK, &ReturnData{Err_code: 1, Err_msg: "ok", Data: pro.Id})
}

// @Summary 增加项目
// @Accept application/json
// @Param Id body int true "Id"
// @Param ProjectName body string true "ProjectName"
// @Param RepoType body string true "目前用不到"
// @Param RepoUrl body string true "目前用不到"
// @Param Tags body string true "项目类型 type=ops or tenant or center"
// @Param Enabled body bool true "Enabled"
// @Param ProjectToken body string true "项目令牌 目前都是 cicd??????"
// @Param EnabledBranchs body string true "可发版的分支，使用正则表达式"
// @Param Descript body string true "Descript"
// @Param Topic body string  true "Topic"
// @Param GitlabId body string true "GitlabId"
// @Param Group body string true "Group"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/project [patch]
func PatchProjects(c *gin.Context) {
	pro := m.Project{}
	err := c.BindJSON(&pro)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}
	devopsdb := database.GetDevopsDb()
	result, err := devopsdb.UpdateProject(c, pro)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Patch Project failed err:", err)
		return
	}

	//if result {
	//	c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "ok"})
	//	return
	//} else {
	//	c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "update failed"})
	//	return
	//}
	if result {
		c.JSON(http.StatusOK, &ReturnData{Err_code: 1, Err_msg: "ok"})
	}
}

func AddNewProResourceDefault(c *gin.Context) {
	newProResourceDefault := project.NewProResourceDefault{}
	err := c.BindJSON(&newProResourceDefault)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "参数绑定错误！"})
		return
	}
	res, err := project.AddNewProResourceFromTemp(c, newProResourceDefault)
	if err != nil {
		c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: err.Error(), Data: res})
		return
	}
	c.JSON(http.StatusOK, &ReturnData{Err_code: 1, Err_msg: "ok", Data: res})
}

func SyncProjectFromGitlab(c *gin.Context) {
	project := &m.Project{}
	err := c.BindJSON(project)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": 0, "err_msg": "参数绑定错误"})
		return
	}
	if project.RepoType == "gitlab" {
		project.RepoUrl = strings.TrimSpace(project.RepoUrl)
		repoUrlSplit := strings.Split(strings.TrimSuffix(project.RepoUrl, `.git`), "gitlab.yiban.io/")
		if len(repoUrlSplit) != 2 {
			c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "仓库地址不合法！"})
			return
		}
		groupPro := strings.Trim(repoUrlSplit[1], "/")
		groupProSplit := strings.Split(groupPro, "/")
		if len(groupProSplit) < 2 {
			c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "仓库地址不合法！"})
			return
		}

		gitlabPro, err := gitlabapi.GitlabApiObj.GetSingleProject(c, groupProSplit[0], groupProSplit[1])
		if err != nil {
			logrus.Error(err)
			c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "从gitlab获取项目信息出错！"})
			return
		}
		project.GitlabId = gitlabPro.ID
		project.Group = groupProSplit[0]
		project.ProjectName = gitlabPro.Name
		project.Enabled = true
		topic := ""
		for _, t := range gitlabPro.Topics {
			topic = topic + t + ","
		}
		topic = strings.Trim(topic, ",")
		project.Topic = topic
		project.Descript = gitlabPro.Description
	}
	devopsdb := database.GetDevopsDb()
	ok, err := devopsdb.UpdateProject(c, *project)
	if err != nil || !ok {
		logrus.Error(err)
		c.JSON(200, gin.H{"err_code": 0, "err_msg": "查询错误"})
		return
	}
	c.JSON(200, gin.H{"err_code": 1, "err_msg": "ok", "data": project})

}

// DeleteProject
// @Summary 删除一个项目
// @Param proid query string true "项目ID"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/project [delete]
func DeleteProject(c *gin.Context) {
	strid := c.Query("proid")
	proid, err := strconv.Atoi(strid)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}
	devopsdb := database.GetDevopsDb()
	_, err = devopsdb.DeleteProjectById(c, proid)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Delete Project failed err:", err)
		return
	}

	c.JSON(http.StatusOK, &ReturnData{
		Err_code: 1,
		Err_msg:  "ok",
	})
}

// AddDingTalkBot
// @Summary 添加一个钉钉机器人
// @Param dingtalk_bot_hook body string  true "钉钉机器人连接"
// @Param descript body string true "描述信息"
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/dingtalk_bot [put]
func AddDingTalkBot(c *gin.Context) {
	dingbot := &model.DingTalkBot{}
	err := c.BindJSON(dingbot)
	if err != nil {
		c.JSON(400, ReturnData{Err_code: 0, Err_msg: "参数绑定错误"})
		return
	}

	devopsdb := database.GetDevopsDb()
	_, err = devopsdb.InsertIntoDingtalkBot(c, *dingbot)
	if err != nil {
		c.JSON(500, ReturnData{Err_code: 0, Err_msg: "插入错误"})
		return
	}
	c.JSON(http.StatusOK, ReturnData{Err_code: 1, Err_msg: "ok"})
}

// ListDingTalkBot
// @Summary 展示钉钉机器人
// @Param project_id query string true "项目ID"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/list_dingtalk_bot [get]
func ListDingTalkBot(c *gin.Context) {
	proIdstr := c.Query("project_id")
	proId, err := strconv.Atoi(proIdstr)
	if err != nil {
		c.JSON(400, "参数错误")
		return
	}
	devopsdb := database.GetDevopsDb()
	dingtalkbots, err := devopsdb.SelectDingtalkBotByPro(c, proId)
	if err != nil {
		c.JSON(500, "查询错误1")
		return
	}
	c.JSON(200, ReturnData{Err_code: 1, Err_msg: "ok", Data: dingtalkbots})
}

// DeleteDingtalkBot
// @Summary 删除一个钉钉机器人
// @Param dingtalk_bot_id query string true "钉钉机器人ID"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/dingtalk_bot [delete]
func DeleteDingTalkBot(c *gin.Context) {
	dingIdstr := c.Query("dingtalk_bot_id")
	fmt.Println(dingIdstr)
	dingId, err := strconv.Atoi(dingIdstr)
	if err != nil {
		logrus.Error(err)
		c.JSON(400, ReturnData{Err_code: 0, Err_msg: "参数错误"})
		return
	}
	devopsdb := database.GetDevopsDb()
	_, err = devopsdb.DeleteDingtalkBotById(c, dingId)
	if err != nil {
		c.JSON(400, ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}
	c.JSON(200, ReturnData{Err_code: 1, Err_msg: "ok"})
}