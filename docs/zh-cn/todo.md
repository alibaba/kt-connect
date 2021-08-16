# v1.0 版本计划

> 如果你有什么其它好的点子，请在[Github](https://github.com/alibaba/kt-connect/issues/new?assignees=&labels=&template=feature_request.md&title=)中反馈

## 命令行全局能力增强

* 支持后台运行
* 支持全局配置，简化命令行参数
* 支持全局参数和子命令参数任意顺序混用

## 子命令功能增强

* `connect`命令的vpn模式支持自动安装sshuttle依赖
* `exchange`和`mesh`命令支持使用Service名作为目标
* 支持对StatefulSet资源的`exchange`和`mesh`操作
* `mesh`命令支持自动设置流量规则（配合Istio）

## 服务发现能力增强

> KT目前只支持Kubernetes原生的服务发现能力

* Nacos服务发现：Shadow容器能够自动向Nacos服务进行服务实例注册和注销（待调研）
