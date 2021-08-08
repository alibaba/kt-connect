## Command: ktctl clean

清理集群里已超期的KT代理容器

### 示例

```
ktctl clean
```

### 常用参数

```
--dryRun                  只打印要删除的Kubernetes资源名称，不删除资源
--thresholdInMinus value  清理至少已失联超过多长时间的Kubernetes资源 (单位：分钟，默认值：30)
```

### 从父命令集成的参数

```
--namespace value, -n value   目标Namespace名称 (默认值：default)
--kubeconfig value, -c value  使用的Kubernetes集群配置文件 (默认值：~/.kube/config，若存在KUBECONFIG变量则从该变量读取)
--image value, -i value       指定使用的代理镜像 (默认值：registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:<当前版本>)
--debug, -d                   开启调试日志
--label value, -l value       为代理Pod增加额外标签，例如 'label1=val1,label2=val2'
--forceUpdate                 创建代理Pod时强制更新最新镜像版本
```
