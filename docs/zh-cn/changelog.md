更新日志
---

### 0.2.0

> 发布时间：待定

* Kubernetes最低兼容版本提高到`1.16`
* 新增exchange的`ephemeral`模式（for k8s 1.23+）

### 0.1.2

> 发布时间：2021-08-29

* 自动解析本地DNS配置，移除connect命令的`--localDomain`参数
* 使用vpn模式时自动检测并安装sshuttle，简化初次使用的准备工作
* 解决Exchange和Mesh连接闲置超时报"lost connection to pod"的问题
* 修复Connect命令开启debug模式时无法连接的错误
* 优化Windows环境的屏幕输出，适配非管理员用户场景
* 新增`--imagePullSecret`参数支持指定拉取代理Pod镜像使用的Secret（感谢@[pvtyuan](https://github.com/pvtyuan)）

### 0.1.1

> 发布时间：2021-08-19

* 发布包从`tar.gz`格式改为`zip`格式，方便Windows用户使用
* 新增`--serviceAccount`参数支持指定代理Pod使用的ServiceAccount
* 新增`--useKubectl`参数支持使用本地`kubectl`工具连接集群
* 增强`clean`命令支持清理残留的ConfigMap和注册表数据
* 修复Kubernetes地址有上下文路径会导致无法连接的问题
* 修复执行connect使用sudo导致.ktctl目录owner变成root的问题

### 0.1.0

> 发布时间：2021-08-08

* 增强Windows下的`connect`命令支持
* 移除对本地`kubectl`客户端工具的依赖
* 新增适用于Linux的`tun`连接模式（感谢@[xyz-li](https://github.com/xyz-li)）
* 使用`provide`命令替代`run`命令
* 新增`clean`命令，清理集群中残留的Shadow Pods
* 支持`service.namespace.svc`结构的服务域名解析
* 完善缺失`sshuttle`依赖等运行时错误的报错信息
* `connect`命令的`dump2hosts`参数支持完整服务域名

### 0.0.13-rc13

> 发布时间：2021-07-02

* 提供`kubectl`工具的`exchange`/`mesh`/`run`插件
* Exchange和Mesh命令支持多端口映射
* 消除本地SSH命令行工具依赖
* 用端口检查替代固定延迟，提升Connect命令执行效率
* 支持本地访问StatefulSet的Pod域名
* 兼容OpenShift 4.x

### 0.0.12

> 发布时间：2020-04-13

* 提供`kubectl`工具的`connect`插件
* 支持Dump任意Namespace中的服务路由到本地Hosts文件
* 支持复用Shadow Pod
* 动态生成SSH Key
* 支持`run`命令，直接暴露本地指定端口的服务到集群
* 优化等待Shadow Pod就绪的检查

### 0.0.11

> 发布时间：2020-02-27

* 支持在本地使用`<servicename>.<namespace>`访问集群中的服务
* 添加`check`命令，用于校验本地环境依赖
* 添加`dashboard`命令，支持dashboard的使用
* 修复部分场景下命令不退出的问题

### 0.0.10

> 发布时间：2020-02-02

* ktctl命令参数适配windows操作系统
* 添加`--dump2hosts`参数用于通过service到本地hosts文件，支持socks5模式下在本地使用域名访问 

### 0.0.9

> 发布时间：2020-01-16

* Support Service Name as dns address
* Make sure shadow is clean up after command exit

### 0.0.8

> 发布时间：2019-12-13

* 添加Windows原生支持
* 添加IDEA支持

### 0.0.7

> 发布时间：2019-12-05

* 添加Oidc插件支持TKE集群
* 新增SOCKS5代理模式以支持WSL环境下使用
* 修复了当Node中不包含Pod网段信息时PodIP无法访问的问题

### 0.0.6

> 发布时间：2019-10-10

* 修复ClusterIP cidr地址获取逻辑
* 重新规划托管docker 镜像仓库地址

### 0.0.5

> 发布时间：2019-10-09

* 开源Dashboard相关代码

### 0.0.4 

> 发布时间：2019-06-26

* Dashboard可视化能力支持

### 0.0.3

> 发布时间：2019-06-19

* 添加mesh命令，支持基于Istio的流量调度能力

### 0.0.2

> 发布时间：2019-06-19

* 修复当Namespace启用Istio自动注入后，exchange无法转发请求到本地问题
* exchange命令支持独立运行

### 0.0.1

> 发布时间：2019-06-18

* 拆分connect与exchaneg子命令，支持多应用转发请求到本地
