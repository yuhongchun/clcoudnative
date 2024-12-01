package server

import (
	"context"
	"errors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"devops_build/config"
	"devops_build/database/relational/postgres"
	"devops_build/internal/router"
	"devops_build/pkg"
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "start the service",
	Long:  "start the whole service",
	// PreRun: func(cmd *cobra.Command, args []string) {
	// },
	PreRun: func(cmd *cobra.Command, args []string) {
		usage()
		setup()

	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return run()
	},
}

var cfgFile string

func init() {
	ServerCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./config/settings_default.yaml", "Start service with provided configuration file.")
}

func usage() {
	log.Println("starting server...")
}

func setup() {
	// setup application config from config file
	config.SetUp(cfgFile)
	postgres.PostgresUtils.SetUp(postgres.Options{
		Dsn: config.PostgresConfig.Dsn,
	})
	// init etcd configuration
	// database.SetUp()
	// setup watcher for etcd
	// go database.WatchETCD()
}

func run() error {
	switch config.ApplicationConfig.Mode {
	case pkg.ModeDev:
		gin.SetMode(gin.DebugMode)
		logger.Infof("now in mode %s\n", pkg.ModeDev)
	case pkg.ModeTest:
		gin.SetMode(gin.TestMode)
		logger.Infof("now in mode %s\n", pkg.ModeTest)
	default:
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	// 暂无前端，暂时去掉用户登陆验证环节 || init the usage of sessions
	//r = middleware.InitMiddleWare(r)
	// 创建基于cookie的存储引擎，yangyanxing 参数是用于加密的密钥
	store := cookie.NewStore([]byte("yuhongchun"))
	// 设置session中间件，参数mysession，指的是session的名字，也是cookie的名字
	// store是前面创建的存储引擎，我们可以替换成其他存储引擎
	r.Use(sessions.Sessions("mysession", store))
	r = router.InitRouter(r)

	srv := &http.Server{
		Addr:         config.ApplicationConfig.Host + ":" + config.ApplicationConfig.Port,
		Handler:      r,
		ReadTimeout:  time.Duration(config.ApplicationConfig.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.ApplicationConfig.WriterTimeout) * time.Second,
	}

	go func() {
		// 服务器连接
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("Listen: %s\n", err)
		}
	}()

	logger.Infof("%s Server Run http://%s:%s/ \r\n",
		time.Now().String(),
		config.ApplicationConfig.Host,
		config.ApplicationConfig.Port)
	logger.Infof("%s Enter Control + C Shutdown Server \r\n", time.Now().String())

	// end server gracefully
	quit := make(chan os.Signal, 1)
	// Notify Used os.Interrupt, should be buffed.
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.Infof("%s 服务器正在关闭......\n", time.Now().String())
	// record.ReleaseQueue()
	// database.ETCDClient.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("服务器关闭:", err)
	}
	logger.Info("服务器已关闭")

	return nil
}
