Ktctl Preview
---

用于将本地服务添加到Kubernetes集群。基本用法如下：

```bash
ktctl preview <新建服务名> --expose <本地端口>:<新建服务端口>
```

命令可选参数：

```
--expose value       指定本地服务监听的端口，格式为`port`或`local:remote`，多个端口用逗号分隔，例如：7001,8080:80
--external           创建`LoadBalancer`类型的Service（生成可暴露到集群外的服务地址）
--skipPortChecking   不必检查指定的本地端口是否有服务监听
```

关键参数说明：

- `--expose`是一个必须的参数，它的值应当与本地运行服务的端口一致，若希望创建的Service使用与本地服务不同的端口，则应当使用`<本地端口>:<预期Service端口>`的方式来指定。
