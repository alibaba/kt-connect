升级说明
---

KtConnect的不兼容版本升级参考

# 0.2.x → 0.3.x

#### ① 各子命令的`--method`参数统一改为`--mode`

同时重命名了部分子命令的运行模式，如Connect的`vpn`模式改为`sshuttle`模式，功能保持一致。

#### ② `ktctl connect`命令的`--cidr`参数更名为`--includeIps`，`--dump2hosts`参数被`--dnsMode`参数替代

`--includeIps`参数与`--excludeIps`参数相匹配，更直观的反映该参数功能。

`--dnsMode`参数的`hosts`模式替代了`--dump2hosts`的功能，同时该参数提供了更丰富的域名解析方式选择。

#### ③ `ktctl provide`命令更名为`ktctl preview`

"预览(preview)"服务比"提供(provide)"服务与该命令的使用场景更匹配。

#### ④ Exchange和Mesh默认指定service

原先的Exchange命令默认是指定Deployment名称（`scale`模式）或Pod名称（`ephemeral`模式），但命令实际的直观效果是替换了相应Service的流量，直接指定Service名称在思维方式上更加符合直觉。

同时该命令依然兼容指定deployment/pod，使用如`ktctl exchange --mode scale deployment/tomcat`或`ktctl exchange --mode ephemeral pod/tomcat`格式即可。

#### ⑤ 默认mesh标签从kt-version改为version

原先的`kt-version`标签与Istio的惯例路由标签不符，导致规则配置不便，因此改为更符合普遍通用的`version`标签。

#### ⑥ 移除`ktctl dashboard`命令和`kubectl`命令插件

由于实际使用场景较少，Dashboard命令和Kubectl插件已不再维护。

---

# 0.1.x → 0.2.x

#### ① Kubernetes最低兼容版本提高到`1.16`

不再支持`1.15`以及之前版本的Kubernetes集群。

#### ② 使用shadow pod代替shadow deployment

相比使用shadow deployment间接创建的shadow pod，直接创建的shadow pod更加轻量，同时对用户业务的资源视图干扰也更少。

注意：新版本的`ktctl clean`命令只会清理集群中的shadow pod，如需清理旧版本在集群中遗留的shadow deployment，请降级至v0.1.x版本。

#### ③ Windows版本`ktctl connect`命令的`socks`模式默认不再自动设置全局代理

可使用`--setupGlobalProxy`参数手工开启自动设置全局代理功能，此参数对`socks5`模式也适用。

#### ④ `ktctl connect`命令使用`--withLabel`参数替代`--label`参数

`--withLabel`参数作用和值格式与原`--label`参数相同，同时增加`--withAnnotation`参数用于为shadow pod指定额外标注。

#### ⑤ 移除`ktctl connect`命令命令的`--global`参数

KtConnect现在能够自动根据用户是否具有全局权限，自动适配查询的Namespace范围，不再需要手工设定。

#### ⑥ 移除`ktctl check`命令

KtConnect现在会在执行相关命令时自动检测并尝试安装缺失的组件，不再需要手工执行检查。

---

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
