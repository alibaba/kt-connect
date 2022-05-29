集群权限
---

`ktctl`命令行工具在于Kubernetes集群交互建立网络通道的过程中，需要具有对集群Pod、Service、ConfigMap等资源类型的操作权限。

如果您在使用`ktctl`时遇到了与权限有关的错误，比如"User \"xxx\" cannot list resource \"yyy\" in API group \"zzz\""，说明您当前使用的Role或ClusterRole缺少KT运行所需的权限。

以下列举了几种典型功能集合KT所需的最小RBAC权限配置：

- 极小权限，仅支持`connect`命令使用默认参数运行：[connect-only-mini.yaml](https://github.com/alibaba/kt-connect/blob/master/docs/deploy/rbac/connect-only-mini.yaml)
- 支持使用`connect`命令的完整功能：[connect-only-full.yaml](https://github.com/alibaba/kt-connect/blob/master/docs/deploy/rbac/connect-only-full.yaml)
- 支持所有命令使用默认参数运行：[all-commands-mini.yaml](https://github.com/alibaba/kt-connect/blob/master/docs/deploy/rbac/all-commands-mini.yaml)
- 支持所有命令的全部完整功能：[all-commands-full.yaml](https://github.com/alibaba/kt-connect/blob/master/docs/deploy/rbac/all-commands-full.yaml)

对于有权限管控要求而无法直接分发kubeconfig配置文件给开发者的企业，也可以使用KT的[定制编译版本](zh-cn/reference/customize.md)，将具有相应集群操作权限的kubeconfig配置内置到`ktctl`二进制文件中。
