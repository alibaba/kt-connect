更新日志
---

### 0.3.5

> 发布时间：2022-05-30

- 增加`config`命令用于支持全局默认配置
- `exchange`/`mesh`/`preview`命令支持跳过本地端口检查（感谢 @[wuxs](https://github.com/wuxs)）
- 增加对本地路由只有部分设置成功情况的检查
- 去除`connect`命令对集群Namespace查询权限的依赖
- 支持自定义本地DNS的代理目标地址和顺序
- 支持定制嵌入kubeconfig配置
- 本地配置目录由`.ktctl`更名为`.kt`
- 修复一处hosts文件修改影响内网域名访问的问题（感谢 @[cryice](https://github.com/cryice)）
- 修复Windows下与OpenVPN共存时的DNS顺序问题
- 修复某些Windows环境本地路由未正确移除的问题（issue-294）

### 0.3.4

> 发布时间：2022-05-04

- `connect`/`exchange`/`mesh`/`preview`命令支持断网自动重连
- 在调试模式下，将后台任务日志写到独立日志文件
- 修复0.3.3版本引入的DNS解析失效问题

### 0.3.3

> 发布时间：2022-04-27

- 支持在任意位置使用全局参数
- `mesh`命令对带有未知Header值的请求改为路由到默认环境，不再报"404"错误
- `exchange`和`mesh`命令支持使用端口名称定义的Service Port (issue-172)
- `clean`命令支持清理本地残留的路由表配置 (issue-294)
- 启动时显示当前连接的Kubernetes集群名称和配置的context名称 (issue-305)
- 当无法找到可用端口时，尝试监听随机端口，规避某些环境下端口检查逻辑不正常的问题 (issue-284)
- 修复`mesh`命令退出时，Router Pod没有被正确删除的问题
- 修复`preview`命令创建的Service误用`--expose`参数本地端口号的问题
- 修复由于开发者之间的本地时间不一致，导致误清理集群未过期资源的问题 (issue-297)
- 修复由于Cluster IP段与API Server地址重合，导致DNS端口Port Forward失败的问题 (issue-300)
- 修复代理DNS对CName记录的域名处理不当的问题 (issue-190)

### 0.3.2

> 发布时间：2022-03-28

- 增加`recover`命令用于立即恢复指定服务被`exchange`或`mesh`的流量
- `connect`运行时自动静默清理集群里的过期资源，可部分替代手工执行`clean`命令的功能
- `connect`命令增加`--useShadowDeployment`参数，支持使用Deployment部署Shadow容器
- `connect`命令增加`--podQuota`参数，支持配置ShadowPod和RouterPod的资源限额（issue-277）
- `connect`命令路由规则不再读取节点的PodCIDR配置，去除对节点权限的依赖
- `connect`命令的hosts域名解析模式增加对Service变化的监听和自动适配
- `exchange`/`mesh`的目标被占用时，显示占用者信息
- `mesh`命令的manual模式现在统一使用Service名作为目标参数
- 修复Windows环境在某些情况下路由设置不生效的问题（感谢@[dominicqi](https://github.com/dominicqi)）
- 修复Windows环境下CPU和内存占用时常飙高的问题（issue-291）
- 修复`ktctl`未加子命令时的运行报错（issue-282）

### 0.3.1

> 发布时间：2022-02-20

- 支持多个KubeConfig文件合并（issue-270）
- 支持自定义DNS缓存时长
- 修复`connect`命令使用`localDNS`模式在某些环境下无法解析域名的问题
- 修复`exchange`命令在网络偶发异常恢复后，依然打印连接异常日志的问题
- 修复`mesh`命令`auto`模式偶发Stuntman Service指向Router Pod的问题
- 修复`clean`命令未正确清理`exchange`创建的遗留资源的问题
- 资源心跳包间隔和加锁超时均缩短到3分钟，加速遗留资源清理
- Shadow Pod和Router Pod的容器增加`ports`属性

### 0.3.0

> 发布时间：2022-02-13

- `connect`命令支持`tun2socks`模式
- `connect`命令支持本地DNS，支持同时解析集群服务域名和本地内/外网域名
- `connect`命令支持在所有系统下访问Headless Service
- `exchange`命令默认采用`selector`模式
- `mesh`命令默认采用`auto`模式
- `exchange`和`mesh`统一使用Service作为目标
- 废弃`dashboard`命令
- 废弃`kubectl`插件
- 增加`exchange`/`mesh`命令的目标端口校验
- 修复命令行参数有效性校验

### 0.2.5

> 发布时间：2021-12-30

- 优化`provide`命令的`--expose`参数，支持多端口和端口映射
- 优化`clean`命令的清理标记方式，解决资源清理不彻底问题
- 优化`connect`命令Pod IP范围的计算逻辑，避免对无关IP段的路由影响
- 添加`--nodeSelector`参数，支持指定Shadow Pod到指定节点（issue-185）
- 解决Auto Mesh以后，服务重新部署可能导致Mesh失效的问题
- 修复`clean`命令清理时遇到状态已经是`Terminating`的资源会报错的问题
- 修复本地KubeConfig无全局Pod权限时`connect`命令报错的问题
- 修复`connect`命令在某些异常退出情况下会有资源残留的问题

### 0.2.4

> 发布时间：2021-12-23

- 支持`mesh`命令的`auto`模式使用Service名指定访问目标
- 新增`exchange`命令的`switch`模式，退出时不再需要等待Pod恢复
- 移除Shadow Pod的默认密码，统一使用临时私钥访问，提高安全性
- 修复本地kube config文件未配置Namespace时，`ktctl`运行必须指定Namespace的问题
- 修复使用Ctrl+C中断`exchange`退出等待时概率性未清理Shadow Pod的问题
- 修复启动Sshuttle失败时程序无法自动退出的问题

### 0.2.3

> 发布时间：2021-12-09

* `mesh`命令支持无需依赖istio的`auto`模式
* `exchange`命令退出的过程中，支持用Ctrl+C中断等待
* `connect`命令的`--dump2hosts`参数支持非socks模式
* 规范化错误日志输出，尽可能详细的显示报错信息

### 0.2.2

> 发布时间：2021-11-12

* `exchange`命令等待原服务完全恢复后，再退出shadow pod（issue-257）
* `connect`命令新增`--excludeIps`参数，用于排除指定的非集群IP段
* `connect`命令新增`--proxyAddr`参数，用于指定Socks5代理监听的IP地址
* `exchange`/`mesh`/`provide`命令增加本地端口是否有服务监听的检查
* 修复`connect`命令`--cidr`参数指定多个IP段出错的问题
* 修复当本地服务重启或响应超时以后，`exchange`的连接会自动断开的问题

### 0.2.1

> 发布时间：2021-11-07

* 默认使用kubeconfig当前上下文的namespace（issue #102）
* 修复`connect`使用共享shadow pod时的错误（issue #260）
* 新增`--context`全局参数，支持切换kubeconfig内的上下文（issue #261）

### 0.2.0

> 发布时间：2021-10-17

* Kubernetes最低兼容版本提高到`1.16`
* 使用shadow pod代替shadow deployment
* Windows的`socks`模式默认不再自动设置全局代理，新增开启该功能的`--setupGlobalProxy`参数
* 新增`exchange`命令的`ephemeral`模式（for k8s 1.23+，感谢@[xyz-li](https://github.com/xyz-li)）
* 修复`exchange`命令连接时常卡顿的问题（issues #184，感谢@[xyz-li](https://github.com/xyz-li)）
* 当Port-forward的目标端口被占用时提供更优雅的报错信息（感谢@[xyz-li](https://github.com/xyz-li)）
* 自动根据用户权限控制生成路由的范围，去除`connect`命令的`--global`参数
* 优化Connect命令的`--cidr`参数，支持指定多个IP区段
* 参数`--label`更名为`--withLabel`
* 增加`--withAnnotation`参数为shadow pod增加额外标注
* `connect`命令增加`--disablePodIp`参数支持禁用Pod IP路由
* shadow pod增加`kt-user`标注用于记录本地用户名
* 移除`check`命令

### 0.1.2

> 发布时间：2021-08-29

* 自动解析本地DNS配置，移除`connect`命令的`--localDomain`参数
* 使用vpn模式时自动检测并安装`sshuttle`，简化初次使用的准备工作
* 解决`exchange`和`mesh`命令连接闲置超时报"lost connection to pod"的问题
* 修复`connect`命令开启debug模式时无法连接的错误
* 优化Windows环境的屏幕输出，适配非管理员用户场景
* 新增`--imagePullSecret`参数支持指定拉取代理Pod镜像使用的Secret（感谢@[pvtyuan](https://github.com/pvtyuan)）

### 0.1.1

> 发布时间：2021-08-19

* 发布包从`tar.gz`格式改为`zip`格式，方便Windows用户使用
* 新增`--serviceAccount`参数支持指定代理Pod使用的ServiceAccount
* 新增`--useKubectl`参数支持使用本地`kubectl`工具连接集群
* 增强`clean`命令支持清理残留的ConfigMap和注册表数据
* 修复Kubernetes地址有上下文路径会导致无法连接的问题
* 修复执行connect使用sudo导致.ktctl目录owner变成root的问题

### 0.1.0

> 发布时间：2021-08-08

* 增强Windows下的`connect`命令支持
* 移除对本地`kubectl`客户端工具的依赖
* 新增适用于Linux的`tun`连接模式（感谢@[xyz-li](https://github.com/xyz-li)）
* 使用`provide`命令替代`run`命令
* 新增`clean`命令，清理集群中残留的Shadow Pods
* 支持`service.namespace.svc`结构的服务域名解析
* 完善缺失`sshuttle`依赖等运行时错误的报错信息
* `connect`命令的`--dump2hosts`参数支持完整服务域名

### 0.0.13-rc13

> 发布时间：2021-07-02

* 提供`kubectl`工具的`exchange`/`mesh`/`run`插件
* `exchange`和`mesh`命令支持多端口映射
* 消除本地SSH命令行工具依赖
* 用端口检查替代固定延迟，提升`connect`命令执行效率
* 支持本地访问StatefulSet的Pod域名
* 兼容OpenShift 4.x

### 0.0.12

> 发布时间：2020-04-13

* 提供`kubectl`工具的`connect`插件
* 支持Dump任意Namespace中的服务路由到本地Hosts文件
* 支持复用Shadow Pod
* 动态生成SSH Key
* 支持`run`命令，直接暴露本地指定端口的服务到集群
* 优化等待Shadow Pod就绪的检查

### 0.0.11

> 发布时间：2020-02-27

* 支持在本地使用`<servicename>.<namespace>`访问集群中的服务
* 添加`check`命令，用于校验本地环境依赖
* 添加`dashboard`命令，支持dashboard的使用
* 修复部分场景下命令不退出的问题

### 0.0.10

> 发布时间：2020-02-02

* `ktctl`命令参数适配windows操作系统
* 添加`--dump2hosts`参数用于通过service到本地hosts文件，支持socks5模式下在本地使用域名访问 

### 0.0.9

> 发布时间：2020-01-16

* 支持本地直接访问Service名称
* 修复Shadow pod未正确清理的问题

### 0.0.8

> 发布时间：2019-12-13

* 添加Windows原生支持
* 添加IDEA支持

### 0.0.7

> 发布时间：2019-12-05

* 添加Oidc插件支持TKE集群
* 新增SOCKS5代理模式以支持WSL环境下使用
* 修复了当Node中不包含Pod网段信息时PodIP无法访问的问题

### 0.0.6

> 发布时间：2019-10-10

* 修复ClusterIP CIDR地址获取逻辑
* 重新规划托管Docker镜像仓库地址

### 0.0.5

> 发布时间：2019-10-09

* 开源Dashboard相关代码

### 0.0.4 

> 发布时间：2019-06-26

* Dashboard可视化能力支持

### 0.0.3

> 发布时间：2019-06-19

* 添加`mesh`命令，支持基于Istio的流量调度能力

### 0.0.2

> 发布时间：2019-06-19

* 修复当Namespace启用Istio自动注入后，`exchange`无法转发请求到本地问题
* `exchange`命令支持独立运行

### 0.0.1

> 发布时间：2019-06-18

* 拆分`connect`与`exchaneg`子命令，支持多应用转发请求到本地
* 支持同时对多个服务进行`exchaneg`操作