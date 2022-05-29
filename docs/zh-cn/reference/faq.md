常见问题
---

#### Q：有多个测试集群时，如何选择连接到哪个集群 ？

A：`ktctl`采用本地`kubectl`工具的集群配置，默认为用户主目录下的`.kube/config`文件中配置的集群。可通过`KUBECONFIG`环境变量或`--kubeconfig`运行参数指定使用其他配置文件路径。

#### Q：在MacOS/Linux下遇到`too many open files`报错 ？

A：这是由于系统文件句柄数上限不足导致的，解决方法参考：[MacOS](https://www.jianshu.com/p/d6f7d1557f20) / [Linux](https://zhuanlan.zhihu.com/p/75897823)

#### Q：执行Exchange或Mesh后，通过访问目标Service报503错误，且返回Header中包含"server: envoy"内容 ？

A：Exchange默认的`selector`模式和Mesh默认的`auto`模式与Istio服务网格不兼容，如果使用了Istio组件，请使用Exchange的`scale`模式和Mesh的`manual`模式。如果切换后依然存在上述错误，请检查该服务上的VirtualService和DestinationRule规则为何无法选择到KT创建的Shadow Pod。

#### Q：执行`ktctl`命令报错 "unable to do port forwarding: socat not found" 或 "ssh: handshake failed: EOF" ？

A：Ktctl的端口映射功能依赖于集群主机上的`socat`工具，请在集群的各个节点上预先安装（Debian/Ubuntu发行版安装命令：`apt-get install socat`，CentOS/RedHat发行版安装命令：`yum install socat`）

#### Q：启动`ktctl connect`以后，使用域名访问集群中的服务提示 "Could not resolve host" ?

A：使用`--debug`参数重新启动`ktctl connect`命令，在访问域名时观察`ktctl`控制台上是否有相关的域名检查日志输出。若有 "domain <你访问的域名> not exists" 错误，请检查您所连接的集群和使用的服务域名是否正确（可到集群中的Pod访问该域名进行验证）；若无任何与所查域名相关的输出，则说明系统DNS配置未生效，请提交 [issue](https://github.com/golang/go/issues) 告诉我们，并写明本地操作系统版本和使用的`ktctl`版本信息。
