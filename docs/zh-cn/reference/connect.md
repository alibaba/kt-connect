Ktctl Connect原理
---

简单来说，Connect命令的作用是建立了一个从本地到集群的”单向VPN“。
根据VPN原理，打通两个内网必然需要借助一个公共中继节点，KtConnect工具巧妙的利用Kubernetes原生的PortForward能力，简化了建立连接的过程（其实是Kubernetes API Server间接起到了中继节点的作用）。
Connect命令的两种模式本质大同小异，只是在建立连接的方式上稍有不同。

## Tun2Socks模式

首先在集群中创建一个Shadow Pod，它的作用是提供SSH和DNS服务，然后利用PortForward将Shadow Pod的SSH端口映射为本地`2222`端口（可通过`--sshPort`参数修改）。

第二步是利用得到的SSH本地端口再创建出通往Shadow Pod的Socks5协议代理服务，监听本地2223端口（可通过`--proxyPort`参数修改）。

接着在本地创建一个tun设备（可以简单理解为增加了一个临时虚拟网卡），将发往该设备的数据通过Socks5协议发送给Shadow Pod，由Shadow Pod代理访问集群里的其他服务。

最后配置本地路由表，并修改本地DNS配置，将访问属于集群IP段的请求路由到上一步创建的tun设备，同时使用KtConnect提供的临时DNS服务来解析本地服务发起DNS查询。

当KtConnect进程退出时，将自动销毁Shadow Pod和本地tun设备，并恢复本地路由和DNS配置。

## Sshuttle模式

与Tun2Socks模式的主要区别在于，不创建本地tun设备和Socks5代理服务，而是利用一个Sshuttle脚本直接将本地请求通过SSH协议发送给运行在Shadow Pod上的接收脚本，通过后者代理访问集群里的服务。

由于没有tun设备可作为路由表目标，Sshuttle模式利用`iptables`/`ipfwadm`/`nftables`工具（Linux系统）或`pfctl`工具（MacOS系统）来实现让目的地址是集群资源IP的请求发往代理服务。Windows系统中不存在类似`iptables`这样功能强大的路由控制工具，因而该模式尚未支持Windows系统。
