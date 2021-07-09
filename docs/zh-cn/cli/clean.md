## Command: ktctl clean

清理集群里已超期的KT代理容器

### 示例

```
ktctl clean
```

### 参数

```
--dryRun                  Only print name of deployments to be deleted
--thresholdInMinus value  Length of allowed disconnection time before a unavailing shadow pod be deleted (default: 30)
```

### 从父命令集成的参数

```
--namespace value, -n value   (default: "default")
--kubeconfig value, -c value  (default: env from KUBECONFIG)
--image value, -i value       Custom proxy image (default: "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:stable")
--debug, -d                   debug mode
--label value, -l value       Extra labels on proxy pod e.g. 'label1=val1,label2=val2'
--help, -h                    show help
--version, -v                 print the version
```
