Ktctl Mesh
---

用于将指定服务的部分流量重定向到本地。基本用法如下：

```bash
ktctl mesh <目标服务名> --expose <本地端口>:<目标服务端口>
```

命令可选参数：

```
--mode value        实现流量重定向的路由方式，可选值为 "auto"（默认）和 "manual"
--expose value      指定目标服务的一个或多个端口，格式为`port`或`local:remote`，多个端口用逗号分隔，例如：7001,8080:80
--versionMark value 指定本地服务路由的版本标签值，格式可以是 `<标签值>`，`<标签名>:` 或 `<标签名>:<标签值>`
--routerImage value （仅用于auto模式）指定Router Pod使用的镜像地址
```

关键参数说明：

- `--mode`提供了两种服务重定向路由的方式。
  默认的`auto`模式采用Router Pod实现HTTP请求的自动路由，无需额外配置服务网格组件，适用于集群中未部署服务网格的场景。
  `manual`模式仅将本地服务"混入"集群中，并打上特定的版本Label，开发者自行通过服务网格组件（如Istio）灵活配置路由规则。
- `--expose`是一个必须的参数，它的值应当与目标Service的`port`属性值相同，若本地运行服务的端口与目标Service的`port`属性值不一致，则应当使用`<本地端口>:<目标Service端口>`的方式来指定。
- `--versionMark`用于指定路由到本地的Header或Label名称和值。默认值为"version:\<随机生成值\>"，可仅指定标签值，如`--versionMark demo`；可用标签名加冒号的格式仅指定标签名，如`--versionMark kt-mark:`；也可以同时指定标签的名称和值，如`--versionMark kt-mark:demo`。
  在`auto`模式下，该值实际上是用于路由的Header。在`manual`模式下，该值为附加在通往本地服务的Shadow Pod上额外的Label。
