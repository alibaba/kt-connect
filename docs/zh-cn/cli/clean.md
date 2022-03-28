Ktctl Clean
---

用于清理集群里已超期的KT代理容器。基本用法如下：

```bash
ktctl clean
```

命令可选参数：

```
--dryRun                  只打印要删除的Kubernetes资源名称，不删除资源
--thresholdInMinus value  清理至少已失联超过多长时间的Kubernetes资源 (单位：分钟，默认值：15)
```

关键参数说明：

- `--thresholdInMinus`参数值通常不宜小于KT资源的默认心跳间隔时长（5分钟），否则可能导致误删正在使用中的正常资源。
