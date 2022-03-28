Ktctl Recover
---

用于立即恢复指定服务被`exchange`或`mesh`命令重定向的流量。基本用法如下：

```bash
ktctl recover <目标服务名>
```

该命令暂无可选参数。

特别说明：

- 该命令仅适用于恢复由KtConnect `0.3.2`及以上版本创建的流量重定向。若集群中有KtConnect `0.3.1`及以下版本的用户，依然建议等待使用者正常退出或异常失联超时后，使用`ktctl clean`命令清理集群残留资源并恢复流量
