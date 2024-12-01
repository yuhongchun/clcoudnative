# 格式化
# go fmt ./...

# 导出apm环境变量

# 移除旧文件
rm -f main
# 进行编译
CGO_ENABLED=0 go build -o main  main.go 
# 启动 

ELASTIC_APM_ACTIVE=false ./main server -c config/settings.yaml 
