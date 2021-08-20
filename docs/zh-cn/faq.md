常见问题
---

#### Q：有多个测试集群时，如何选择连接到哪个集群？

A：`ktctl`采用本地`kubectl`工具的集群配置，即`~/.kube/config`文件中的默认集群。

#### Q：ktctl命令行工具所需的最小RBAC权限？

A：配置在`~/.kube/config`中的账户需要具有对Deployment、Service、ConfigMap资源类型的操作权限，[这个YAML配置](https://github.com/alibaba/kt-connect/blob/feature/minimum-permissions/docs/deploy/rbac/clusterrole.yaml) 展示了`ktctl`工具所需的最小可用权限 。

#### Q：在MacOS/Linux下遇到`too many open files`报错？

A：这是由于系统文件句柄数上限不足导致的，解决方法参考：[MacOS](https://www.jianshu.com/p/d6f7d1557f20) / [Linux](https://zhuanlan.zhihu.com/p/75897823)
