# 介绍

![](../_media/demo-1.gif)

## 特性

* 直接访问Kubernetes集群

开发者通过KT可以直接连接Kubernetes集群内部网络，在不修改代码的情况下完成本地开发与联调测试

* 转发集群流量到本地

开发者可以将集群中的流量转发到本地，从而使得集群中的其它服务可以联调本地

* Service Mesh支持

对于使用Istio的开发者，KT支持创建一个指向本地的Version版本

* 基于SSH的轻量级VPN网络

KT使用shhuttle作为网络连接实现，实现轻量级的SSH VPN网络

* 作为kubectl插件，集成到Kubectl

开发者也可以直接将ktctl集成到kubectl中

## 更新日志

### 0.0.6

* 添加Windows支持文档说明
* 修复ClusterIP cidr地址获取逻辑
* 重新规划托管docker 镜像仓库地址

### 0.0.5

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