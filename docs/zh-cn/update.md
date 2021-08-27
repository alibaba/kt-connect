升级说明
---

KtConnect的不兼容版本升级参考

# 0.0.x → 0.1.x

#### ① Windows版本`ktctl connect`新增`socks`模式

该模式将创建基于`socks4`协议的网络代理（Windows支持此协议的全局代理），并自动修改Windows全局代理配置，在`ktctl connect`运行期间，浏览器和curl等命令行工具将能够直接访问集群内的IP地址。

在`0.1.x`版本中已将`socks`作为默认连接模式，若您不希望自动设置系统全局代理，可通过`--method=socks5`参数继续使用`socks5`运行模式，并手工为需要访问集群的进程设置代理。

#### ② Windows版本`ktctl connect`默认支持访问Service域名

在`0.0.x`版本有`--dump2hosts`和`--dump2hostsNS`两个参数，前者用于启用本地通过Service名访问集群服务功能，后者用于指定需要访问的一个或多个namespace。

从`0.1.x`版本开始，Windows环境下默认会启用Service名访问的功能（无需额外参数），使用`--dump2hosts`参数指定要访问的一个或多个namespace（默认为当前namespace）。

同时，`0.1.x`版本增强了服务域名访问功能，支持`<service-name>`、`<service-name>.<namespace>`和`<service-name>.<namespace>.<cluster-domain>`形式的域名。

#### ③ 使用`ktctl provide`命令替代`ktctl run`

`run`命令带有"新启动一个本地服务进程"的歧义，而这个命令实际作用是将本地开发中的服务"提供"给集群中的其他服务调用。

同时将该命令的`--port`参数改为`--expose`，与`exchange`和`mesh`命令保持一致。

#### ④ 调整`exchange`和`mesh`命令端口映射参数顺序

在`0.0.x`版本中，若需要同时指定本地和远端端口，`--expose`参数格式为`<remote-port>:<local-port>`，在`0.1.x`版本调整为`<local-port>:<remote-port>`，匹配`docker`、`kubectl`等工具的统一惯例。

#### ⑤ 其他细节变化，新增`ktctl clean`命令

包括屏幕输出、ShadowPod命名，本地每个`ktctl`进程使用独立PID文件等细节的改变。

新增的`ktctl clean`命令用于清理当`ktctl`进程非正常结束（譬如直接关闭控制台窗口）时，在集群和本地遗留的资源并还原相关系统配置。包括集群里遗留的代理`Deployment`（以及相关联的`Service`和`ConfigMap`）、本地的全局配置修改和Hosts文件修改等。
