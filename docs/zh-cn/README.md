介绍
---

kt-connect是一款能让开发者在本地运行的服务与Kubernetes集群中的服务双向互通的工具。

## 特性

* 直接访问Kubernetes集群

开发者通过KT可以直接连接Kubernetes集群内部网络，在不修改代码的情况下完成本地联调测试

* 转发集群流量到本地

开发者可以将集群中的流量转发到本地，使得集群中的服务无需额外配置即可访问本地服务

* Service Mesh支持

对于使用Istio的开发者，KT提供指向本地服务Version标签和自定义标签来精细控制流量

* Windows/MacOS/Linux系统支持

通过ssh-vpn/socks-proxy/tun-device等多种通信通道，解决开发者在不同操作系统下的网络访问问题

* 作为插件集成到Kubectl

开发者也可以直接将KT集成到`kubectl`中，直接作为子命令使用

## 联系我们

请加入`kt-dev`钉钉群：

<img src="https://virtual-environment.oss-cn-zhangjiakou.aliyuncs.com/image/dingtalk-group-zh-cn.jpg" width="40%"></img>
