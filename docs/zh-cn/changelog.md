## 更新日志

### 0.0.9

> 发布时间： 2020-01-16

* Support Service Name as dns address
* Make sure shadow is clean up after command exit


### 0.0.8

> 发布时间： 2019-12-13

* 添加Windows原生支持
* 添加IDEA支持

### 0.0.7

> 发布时间： 2019-12-05

* 添加Oidc插件支持TKE集群
* 新增SOCKS5代理模式以支持WSL环境下使用
* 修复了当Node中不包含Pod网段信息时PodIP无法访问的问题

### 0.0.6

> 发布时间： 2019-10-10

* 修复ClusterIP cidr地址获取逻辑
* 重新规划托管docker 镜像仓库地址

### 0.0.5

> 发布时间： 2019-10-09

* 开源Dashboard相关代码

### 0.0.4 

> 发布时间： 2019-06-26

* Dashboard可视化能力支持

### 0.0.3

> 发布时间： 2019-06-19

* 添加mesh命令，支持基于Istio的流量调度能力

### 0.0.2

> 发布时间： 2019-06-19

* 修复当Namespace启用Istio自动注入后，exchange无法转发请求到本地问题
* exchange命令支持独立运行

### 0.0.1

> 发布时间： 2019-06-18

* 拆分connect与exchaneg子命令，支持多应用转发请求到本地