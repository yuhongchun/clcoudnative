### 帮助文档

devops-build目录是devops-build应用的相关源码，作了无害处理，但保留了核心代码
* devops_build/api: devops-build的核心API接口


devops-release目录是devops-release应用的相关源码，作了无害处理，但保留了核心代码,devops-relase应用较为核心有：

* devops_release/util/apollo: apollo_sdk相关核心代码
* devops_release/internal/buildv2 : 针对Kubernetes集群的相关操作均通过buildv2包来实现

#### Go版本解释
最初devops-build和devops-release是基于go 1.17开发，后续发版机器的Go升级到1.20，也是能正常编译运行的。


