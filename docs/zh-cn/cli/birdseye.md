Ktctl Birdseye
---

用于查看当前Namespace中，各个服务的路由状态。基本用法如下：

```bash
ktctl birdseye
```

命令可选参数：

```
--sortBy value        展示服务的排序方式，可选值为 "status"（默认）和 "name"
--showConnector       展示此时连接到集群的所有用户
--hideNaturalService  隐藏未被exchange/mesh的普通服务
```

> 命令中显示出的用户名为开发者本地计算机的登录名

关键参数说明：

- `--sortBy`参数用于指定服务的展示顺序。默认值`status`将依次展示流量被完全代理（`exchange`）的服务、流量被部分代理（`mesh`）的服务、流量未被代理的服务、从本地暴露到集群（`preview`）的服务。可选值`name`将按照服务名的字母顺序依次展示各服务。
