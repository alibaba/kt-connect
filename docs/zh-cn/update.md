升级说明
---

KtConnect的不兼容版本升级参考。

## 0.0.x -> 0.1.x

#### ① Windows版本的`ktctl connect`命令新增`socks`模式，并替代原先的`socks5`成为默认运行模式。

该模式将创建基于`socks4`协议的代理，并自动修改Windows全局代理配置，在`ktctl connect`运行期间，浏览器和curl等命令行工具将能够直接访问集群内的IP地址。

如果不希望自动设置系统全局代理，可以使用`--method=socks5`参数使用`socks5`运行模式，并手工为需要访问集群的进程设置代理。

#### ② Windows版本的`ktctl connect`命令默认会将集群服务域名写入本地hosts文件，`--dump2hosts`参数改为用于指定namespace

在`0.0.x`版本有`--dump2hosts`和`--dump2hostsNS`两个参数，前者用于启用本地通过Service名访问集群服务功能，后者用于指定需要访问的一个或多个namespace

从`0.1.x`版本开始，Windows环境下默认会启用Service名访问的功能，保留`--dump2hosts`参数，但作用与原先的`--dump2hostsNS`参数相同

同时，`0.1.x`版本增强了服务域名访问的功能，不仅支持`<service-name>`形式的域名，还支持`<service-name>.<namespace>`和`<service-name>.<namespace>.<cluster-domain>`的形式

#### ③ 使用`ktctl provide`命令替代`ktctl run`

`run`命令带有"新启动一个本地服务进程"的歧义，而这个命令实际作用是将本地开发中的服务"提供"给集群中的其他服务调用。

同时将该命令的`--port`参数改为`--expose`，与`exchange`和`mesh`命令保持一致。

> 由于`run`、`expose`等名称与`kubectl`的子命令存在重名，会导致`ktctl`作为`kubectl`插件使用时发生冲突，故最终改名`provide`。

#### ④ 调整`exchange`和`mesh`命令端口映射参数顺序

在`0.0.x`版本中，若需要同时指定本地和远端端口，`--expose`参数格式为`<remote-port>:<local-port>`，在`0.1.x`版本调整为`<local-port>:<remote-port>`，与`docker`、`kubectl`等工具保持一致。

#### ⑤ 屏幕输出、ShadowPod命名、PID文件等细节修改，新增`ktctl clean`命令

许多细节变化，包括规范化屏幕输出和ShadowPod命名，本地每个`ktctl`进程使用独立PID文档等。

新增的`ktctl clean`命令用于清理当`ktctl`进程非正常结束（譬如直接关闭控制台窗口）时，在集群和本地遗留的资源并还原相关系统配置。包括集群里遗留的代理`Deployment`（以及相关联的`Service`和`ConfigMap`）、本地的全局配置修改和Hosts文件修改等。
