package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	cfg "devops_release/config"
	"devops_release/database/relational/postgres"
	"devops_release/internal/buildv2"
	"devops_release/util/apollo"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"devops_release/pkg"
	"devops_release/router"

	nlog "github.com/sirupsen/logrus"
)

var (
	config   string
	port     string
	StartCmd = &cobra.Command{
		Use:     "server",
		Short:   "启动服务",
		Example: "./nighting-release server config/settings.yaml",
		PreRun: func(cmd *cobra.Command, args []string) {
			usage()
			setup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			return run()
		},
	}
)

func init() {
	StartCmd.PersistentFlags().StringVarP(&config, "config", "c", "config/settings.yaml", "Start server with provided configuration file")
	StartCmd.PersistentFlags().StringVarP(&port, "port", "p", "5000", "Tcp port server listening on")
}

func usage() {
	usageStr := `starting server`
	nlog.Infof("%s\n", usageStr)
}

func setup() {
	//初始化配置文件
	cfg.Setup(config)
	postgres.PostgresUtils.SetUp(postgres.Options{
		Dsn: cfg.PostgresConfig.Dsn,
	})
}

func run() error {
	// 从配置 读取运行模式
	nlog.Infof("mode:%s\n", cfg.ApplicationConfig.Mode)
	switch cfg.ApplicationConfig.Mode {
	case pkg.ModeProd:
		gin.SetMode(gin.ReleaseMode)
	case pkg.ModeTest:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	r := router.InitRouter()

	// 注册http
	srv := &http.Server{
		Addr:         cfg.ApplicationConfig.Host + ":" + cfg.ApplicationConfig.Port,
		Handler:      r,
		ReadTimeout:  time.Second * time.Duration(cfg.ApplicationConfig.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(cfg.ApplicationConfig.WriterTimeout),
	}

	go func() {
		// 服务器连接
		if err := srv.ListenAndServe(); err != nil {
			nlog.Error("Listen: %s\n", err)
		}
	}()
	//监听weiban下的配置，传入配置变更回调

	go func() {
		apollo.StartListening(buildv2.RestartWithConfig)
	}()

	nlog.Infof("%s Server Run http://%s:%s/ \r\n",
		time.Now().String(),
		cfg.ApplicationConfig.Host,
		cfg.ApplicationConfig.Port)
	nlog.Infof("%s Enter Control + C Shutdown Server \r\n", time.Now().String())

	// end server gracefully
	quit := make(chan os.Signal, 1)
	// Notify Used os.Interrupt, should be buffed.
	signal.Notify(quit, os.Interrupt)
	<-quit
	nlog.Infof("%s 服务器正在关闭......\n", time.Now().String())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		nlog.Fatalf("服务器关闭:", err)
	}
	nlog.Info("服务器已关闭")
	return nil
}
