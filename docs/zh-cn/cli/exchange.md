Ktctl Exchange
---

用于使用本地服务替换集群中的Service实例。基本用法如下：

```bash
ktctl exchange <目标服务名> --expose <本地端口>:<目标服务端口>
```

命令可选参数：

```text
--mode value             重定向网络请求的方法，可选值为 "selector"（默认），"scale" 和 "ephemeral"（实验性功能）
--expose value           指定置换服务的一个或多个端口，格式为`port`或`local:remote`，多个端口用逗号分隔，例如：7001,8080:80
--skipPortChecking       不必检查指定的本地端口是否有服务监听
--recoverWaitTime value  （仅用于scale模式）指定退出时等待原Pod启动完成的最长秒数（默认值为120）
```

关键参数说明：

- `--mode`提供了三种替换服务的方式。
  默认的`selector`模式的流量切换和回切速度最快，无需重启被切换服务的Pod，但在切换期间会对目标服务的`selector`属性有修改，与Istio不兼容；
  `scale`模式不会改到目标服务属性，但切换过程会使目标服务的Pod重启，且回切时需等待原始Pod重启完成，耗时相对较长；
  `ephemeral`模式能够兼备以上两种模式的优点，但该模式当前功能尚未完备，且仅能够用于Kubernetes v1.23及以上版本，暂不推荐使用。
- `--expose`是一个必须的参数，它的值应当与被替换Service的`port`属性值相同，若本地运行服务的端口与目标Service的`port`属性值不一致，则应当使用`<本地端口>:<目标Service端口>`的方式来指定。
