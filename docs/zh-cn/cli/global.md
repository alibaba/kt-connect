全局参数
---

`ktctl`的命令参数分为"全局参数"和"子命令参数"，全局参数应当直接跟在`ktctl`后，子命令参数应当放在具体的子命令后，例如：

```bash
$ ktctl --namespace demo connect --includeIps 10.1.0.0/16
        | <- 全局参数 -> |        | <---- 子命令参数 ----> |
```

可用的全局参数包括：

```text
--namespace value, -n value   指定目标服务的Kubernetes Namespace（若未指定，则使用本地KubeConfig配置的默认Namespace）
--kubeconfig value, -c value  指定本地KubeConfig配置文件路径（默认为"/Users/flin/.kube/config"）
--image value, -i value       指定Shadow Pod使用的镜像（默认为"registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:v0.3.0"）
--imagePullSecret value       指定下载Shadow Pod镜像使用的Secret
--serviceAccount value        指定下载Shadow Pod镜像使用的ServiceAccount（默认为"default"）
--nodeSelector value          指定运行Shadow Pod的节点选择标签，多个标签使用逗号分隔，例如"disk=ssd,region=hangzhou"
--debug, -d                   显示调试日志
--withLabel value, -l value   为Shadow Pod指定额外的标签，多个标签使用逗号分隔，例如"label1=val1,label2=val2"
--withAnnotation value        为Shadow Pod指定额外的注解，多个注解使用逗号分隔，例如"annotation1=val1,annotation2=val2"
--portForwardTimeout value    等待PortForward建立的超时时长，单位秒（默认值是10）
--podCreationTimeout value    等待Shadow Pod和Router Pod创建完成的超时时长，单位秒（默认值是60）
--useShadowDeployment         使用Deployment方式部署Shadow容器
--useLocalTime                使用本地时间（而非集群时间）作为KT资源的心跳包时间戳
--forceUpdate, -f             总是从镜像仓库重新拉取最新的Shadow Pod和Router Pod镜像
--context value               使用本地KubeConfig配置里的指定Context
--podQuota value              指定Shadow Pod和Router Pod的CPU和内存限制（逗号分隔，例如"0.5c,512m"）
--help, -h                    显示帮助信息
--version, -v                 显示命令版本
```

关键参数说明：

- `--namespace`实际是指定将Shadow Pod运行在哪个Namespace。
  对于`connect`、`preview`命令来说，它将影响服务的访问方式，即可以直接通过`<服务名>`访问与Shadow Pod在同一个Namespace的服务，而访问其他Namespace的服务则必须使用`<服务名>.<Namespace>`作为域名。
  对于`exchange`、`mesh`命令来说，必须指定使用与需置换目标服务相同的Namespace。
- `--podQuota`使用`c`表示CPU配额（单位为"核"），使用`k`/`m`/`g`表示内存配额（单位分别为"KB"/"MB"/"GB"）
