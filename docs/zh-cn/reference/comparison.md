相关工具比较
---

`kt-connect`是一款源自于云效内部技术实践的云原生研发辅助工具，其希望解决的问题与网络工具`telepresence`比较相似，但由于技术实现方式的差异，两者在不同场景下的能力也各有特色。

从基本运作模式上来看。`kt-connect`属于内部工具开源，所有功能完全免费，支持企业[快速定制](zh-cn/reference/customize.md)。`telepresence`是开源的商业软件，部分功能需订阅付费后方可使用。从开源定制协议来看，`kt-connect`采用的GPL v3协议不支持将该项目源码用于商业软件（详见[常见问题](zh-cn/reference/faq.md)文档），而`telepresence`的源码无此限制。

|           | kt-connect  |   telepresence   |
| ---       | ---         | ---              |
| 主要开发者  | 阿里云云效   |  Ambassador Labs |
| 开发语言   | Golang      | Golang           |
| 开源协议   | GPL v3      | Apache v2.0      |
| 是否付费   | 完全免费     | 有付费功能         |
| 定制方式   | 配置化定制或修改源码 | 修改源码     |

从功能和实现方式来看，两者的提供的核心能力类似，即解决云原生集群网络双向打通和流量转发问题。在此之上，`telepresence`由于采用了Sidecar机制，能够提供部分类似Istio的网络治理能力和相应的平台服务（需付费订阅），但其网络流量规则与Istio Agent存在不兼容的情况。

正常情况下，`kt-connect`会在客户端退出时完全清除在集群里创建的所有资源，资源侵入性较低。`telepresence`在客户端退出后依然会在集群里留下`ambassador` Namespace以及若干Service和Pod，需要开发者在事后自行清理。在上行链路上（从本地到集群），`kt-connect`基于`socks5`协议封装数据，而`telepresence`使用`grpc`协议，且支持IPv6地址，理论上后者的通信效率略高（没有实测过）。在下线链路上（从集群到本地），`kt-connect`在Service层对流量进行转发，在建立流量通道时不会产生其他副作用，建立连接的速度也更快。`telepresence`在Pod层基于Sidecar对流量进行转发，在建立流量通道时会导致服务相应的所有Pod重启，建立连接的时长与Pod重启到就绪的时长有关，且由于Sidecar转发存在轻微的性能损耗，其理论通信效率更低。

在资源成本方面，`kt-connect`默认为每个用户的上行链路创建独立代理Pod（可配置为共享），`telepresence`的所有上行流量都共用一个代理Pod，后者的资源消耗更少。但对于下行流量，`kt-connect`为每个需要重定向的服务创建一个反向代理Pod，`telepresence`会为每个Pod都增加Sidecar，后者的资源消耗更大。

|               |  kt-connect  |    telepresence  |
| ---           | ---          | ---              |
| 集群资源驻留    | 无           | 一个Namespace和若干Service、Pod |
| 上行连接副作用  | 无            | 无               |
| 代理协议       | socks5       | grpc             |
| IP v4地址     | 支持          | 支持              |
| IP v6地址     | 不支持        | 支持              |
| 下行连接副作用  | 无           | 所有Pod重启        |
| 仅回流部分流量  | 支持         | 支持              |
| 流量上行资源成本 | 较高，每个用户使用独立代理Pod | 较低，所有流量复用一个代理Pod    |
| 流量下行资源成本 | 较低，每个服务一个反向代理Pod | 较高，每个Pod增加一个Sidecar容器 |
| 上下行流量关系  | 可独立使用    | 必须先建立上行通道，才能继续创建下行通道 |
| 暴露本地服务    | 支持         | 支持，需登录       |
| 兼容IstioAgent | 支持         | 不支持            |

在使用方面，`kt-connect`基于云效自身使用过程中遇到的问题，做了许多细节优化，譬如支持本地访问集群服务时使用不带`namespace`尾缀的纯服务名称作为域名（`telepresence`仅支持`service-name.namespace`形式的短域名），支持完全内网环境集群（自定义代理Pod镜像），支持全局默认配置项（ktctl config命令）等等。如果您需要一个小巧实用的Kubernetes云原生研发网络互通工具，`kt-connect`将会是一个不错的选择。
