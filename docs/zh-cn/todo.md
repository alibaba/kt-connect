# TODO:

> 如果你有什么其它好的点子，请在[Github](https://github.com/alibaba/kt-connect/issues/new?assignees=&labels=&template=feature_request.md&title=)中反馈

## 命令行能力增强

* 命令行交互优化，支持进度展示
* Remote SSH Volume
* ~~kubectl dashboard: 命令行打开KT Connect Dashboard (自动安装)~~
* ~~Socks5代理模式支持本地DNS访问服务集群（自动Dump DNS到hosts文件~~

## Dashboard能力增强

> 原生K8S相关

* 支持可视化配置单个服务Istio流量转发规则
* 支持Namespace级别拓扑可视化能力

> KT Virtual Env相关

* 支持自动安装KT VirtualEnv扩展
* 当KT VirtualEnv启用后支持Virtual Env的可视化管理

## 服务发现能力增强

> KT目前只支持Kubernetes原生的服务发现能力

* Nacos服务发现：Shadow容器能够自动向Nacos服务进行服务实例注册和注销（调研）

## 文档和最佳实践

* 提供各语言的最佳实践帮助文档
* ~~Cli References~~