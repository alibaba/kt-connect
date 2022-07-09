Ktctl Forward
---

将集群服务或地址映射到本地端口。基本用法如下：

```bash
ktctl forward <TargetService>
ktctl forward <TargetService> <LocalPort>
ktctl forward <TargetService|TargetIP> <LocalPort>:<TargetServicePort>
```

命令可选参数：

```
无
```

关键参数说明：

- 当第一个参数为Service名，且目标Service对象仅定义了一个端口时，命令的第二个参数可以省略（表示将Service的端口映射为本地相同端口）或仅指定本地端口（表示Service的端口映射为本地指定端口）
