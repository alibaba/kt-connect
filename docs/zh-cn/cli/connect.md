Ktctl Connect
---

用于从本地连接到集群。基本用法如下：

```bash
ktctl connect
```

命令可选参数：

```text
--mode value           与集群建立虚拟连接的方式，可选值为 "tun2socks"（默认）和 "sshuttle"（仅限Linux/Mac）
--dnsMode value        指定解析集群服务域名的方式，可选值为 "localDNS"（默认），"podDNS"（仅用于sshuttle模式）和 "hosts"
--shareShadow          使用在同Namespace下共享的Shadow Pod
--clusterDomain value  指定集群的域名尾缀（默认值为"cluster.local"）
--disablePodIp         禁用Pod IP访问，只能访问服务的Cluster IP或服务域名
--skipCleanup          禁止自动清理集群中残留的过期对象
--includeIps value     将指定IP段指定为集群网段，多个IP段用逗号分隔，IP段格式如 '172.2.0.0/16'
--excludeIps value     将指定IP段指定为非集群网段，多个IP段用逗号分隔，可指定单个IP如 '192.168.64.2' 或IP段如 '192.168.64.0/24'
--disableTunDevice     （仅用于`tun2socks`模式）仅创建Socks5代理，不创建本地tun设备
--disableTunRoute      （仅用于`tun2socks`模式）仅创建tun设备，不自动设置本地路由规则
--proxyPort value      （仅用于`tun2socks`模式）指定Socks5代理监听的端口（默认值为2223）
--dnsCacheTtl value    （仅用于`localDNS`模式）指定DNS缓存的超时秒数（默认值为60）
```

关键参数说明：

- `--mode`提供了两种连接集群的方式。除非由于特定原因无法使用默认的`tun2socks`模式或需要排除某些IP段的路由，否则不建议修改此参数。
- `--dnsMode`提供了三种解析集群服务域名的方式。
 `localDNS`模式将在本地启动临时的域名解析服务，它会先尝试在集群中查找目标域名，若未找到再通过系统的上游DNS查找，可通过`localDNS:<dns1>,<dns2>`格式指定查找顺序，其中<dns>值可以为`IP地址:端口`格式，或特殊值`upstream`(系统上游DNS)和`cluster`(集群DNS)；
 `podDNS`模式将使用集群的DNS服务解析所有域名，
 `hosts`模式用于限定本地只允许访问指定Namespace的服务域名，可通过`hosts:<namespaces>`格式指定可访问的Namespace列表，逗号分隔，如`--dnsMode hosts:default,dev,test`，默认只能访问Shadow Pod所在Namespace的服务。
- `--shareShadow`参数允许所有在同一个Namespace下工作的开发者共用一个Shadow Pod，这种方式能够在一定程度上节约集群资源，但在Shadow Pod偶然发生崩溃时，会同时影响到所有开发者。
