技术原理
---

## Connect命令

简单来说，Connect命令的作用是建立了一个从本地到集群的”单向VPN“。
根据VPN原理，打通两个内网必然需要借助一个公共中继节点，KtConnect工具巧妙的利用Kubernetes原生的PortForward能力，简化了建立连接的过程（其实是Kubernetes API Server间接起到了中继节点的作用）。
Connect命令的两种模式本质大同小异，只是在建立连接的方式上稍有不同。

#### Tun2Socks模式

首先在集群中创建一个Shadow Pod，它的作用是提供SSH和DNS服务，然后利用PortForward将Shadow Pod的SSH端口映射到本地并创建通往Shadow Pod的Socks5协议代理服务，监听本地2223端口（可通过`--proxyPort`参数修改）。

接着在本地创建一个tun设备（可以简单理解为一个临时的虚拟网卡），将发往该设备的数据通过Socks5协议发送给Shadow Pod，由Shadow Pod代理访问集群里的其他服务。

最后配置本地路由表，并修改本地DNS配置，将访问属于集群IP段的请求路由到上一步创建的tun设备，同时使用KtConnect提供的临时DNS服务来解析本地服务发起DNS查询。

当ktctl connect命令退出时，将自动销毁Shadow Pod和本地tun设备，并恢复本地路由和DNS配置。

#### Sshuttle模式

与Tun2Socks模式的主要区别在于，不创建本地tun设备和Socks5代理服务，而是利用一个Sshuttle脚本直接将本地请求通过SSH协议发送给运行在Shadow Pod上的接收脚本，通过后者代理访问集群里的服务。

由于没有tun设备可作为路由表目标，Sshuttle模式利用`iptables`/`ipfwadm`/`nftables`工具（Linux系统）或`pfctl`工具（MacOS系统）来实现让目的地址是集群资源IP的请求发往代理服务。Windows系统中不存在类似`iptables`这样功能强大的路由控制工具，因而该模式尚无法支持Windows系统。

![connect](https://img.alicdn.com/imgextra/i3/O1CN010F3ixF1rYXjpVfHuq_!!6000000005643-0-tps-2482-886.jpg)

## Exchange命令

Exchange/Mesh/Preview都使用Shadow Pod作为流量从集群发往本地的代理，利用Kubernetes和SSH协议的两次PortForward将发往Shadow Pod指定端口的请求转发到本地的指定端口。

#### Selector模式

启动Shadow Pod并建立与本地的连接后，KtConnect会修改原Service的`selector`属性，使得它直接指向Shadow Pod，同时将`selector`的原始值记录到Service的`kt-selector` Annotation里，因此访问目标Service的流量就被重定向到了本地端口。

当ktctl exchange命令退出时，自动还原目标Service的`selector`属性，使流量恢复回原本的Pod。

#### Scale模式

与Selector模式的不同之处在于，它不会改变Service的属性，而是将Shadow Pod的Label设置为与目标Service相匹配，同时缩容目标Service原本的Pod实例数量到0（KtConnect假设原服务是通过Deployment部署的，否则将报错并进入退出流程），从而让流量自然进入Shadow Pod。

当ktctl exchange命令退出时，自动将Service原本的Pod实例扩容回原始数量，使流量恢复正常。

![exchange](https://img.alicdn.com/imgextra/i4/O1CN01oZHLhc1YxIRdf3Oa6_!!6000000003125-0-tps-2486-908.jpg)

## Mesh命令

Mesh借助Router Pod或服务网格（如Istio）的能力实现按规则重定向流量。

#### Auto模式

启动Shadow Pod和Shadow Pod并检查是否存在与目标Service匹配的Router Pod，若未找到则会创建Router Pod，修改目标Service的`selector`属性指向Router Pod，同时再创建一个Stuntman Service指向原Service的Pod实例。

然后更新Router Pod的路由规则，将包含特定Header的请求转发到Shadow Service，其余请求则通过Stuntman Service回到测试环境原本的Pod实例。

当ktctl mesh命令退出时，会移除Router Pod中通往当前Shadow Pod的路由规则。若Router Pod中已经没有其他路由，则同时删除Router Pod和Stuntman Service，并恢复原Service的`selector`属性内容。

#### Manual模式

与Auto模式的差异在于，创建出的Shadow Pod除了用于标识版本的`version`标签外（可通过`--versionMark`参数修改），其余Label均与目标Service的原始Pod实例相同，且不会自动创建Shadow Service、Router Pod和Stuntman Service。

开发者可以通过Pod的版本标识，自行创建服务网格规则，来控制哪些流量应该进入开发者的本地环境。

![mesh](https://img.alicdn.com/imgextra/i4/O1CN01cCUk1Z1xnYPYuGTlB_!!6000000006488-0-tps-2486-986.jpg)

## Preview命令

Preview命令只是简单的创建Shadow Pod，并使用用户指定的名称创建Shadow Service，访问Shadow Service的请求被自然的路由到本地指定端口。

当ktctl preview命令退出时，会直接移除相应的Shadow Pod和Shadow Service。

![preview](https://img.alicdn.com/imgextra/i4/O1CN01sEZYfx1RF8wxtPZUg_!!6000000002081-0-tps-2484-878.jpg)

## Clean命令

除了访问Shadow Pod使用的临时私钥和Pid文件，KtConnect不会在本地留下任何关于运行状态的记录文件，Clean命令主要依靠记录在Shadow Pod、Router Pod、Shadow Service和被重定向的Service上的Annotation内容来恢复环境的原始状态（请不要手工修改以"kt-"开头的Annotation值）。

譬如，本地进程有时会由于异常退出或网络断连而无法主动清理创建到集群里的资源对象，因此KtConnect进程会在创建的对象上使用`kt-last-heart-beat` Annotation定期刷新时间戳，ktctl clean命令就可以依据这个Annotation找到集群中已经不再使用的残留资源并予以清除。
