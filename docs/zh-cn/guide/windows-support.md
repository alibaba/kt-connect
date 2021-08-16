Windows支持
=============

KtConnect目前只支持在Windows环境使用基于Socks代理的连接方法，分为`socks`和`socks5`两种模式。

## 使用`socks`模式连接

`socks`是KtConnect在Windows环境的默认工作模式，无需增加额外参数：

```bash
$ ktctl connect
```

此模式下，KtConnect会启动一个连接Kubernetes集群的Socks4协议代理进程，并自动设置本地系统的全局代理配置和`http_proxy`环境变量指向该代理，此后启动的浏览器、控制台等进程均会自动通过该代理获得访问集群Pod和Service地址的能力。

> 由于Windows的全局代理配置并不对所有应用程序都具有约束性，譬如通常用户自行开发的软件并不会遵循此配置

当KtConnect进程正常结束时，会自动还原系统代理配置和环境变量，若遇到KtConnect进程异常关闭而未能恢复系统配置时，可以事后手工执行`ktctl clean`命令进行清理。

## 使用`socks5`模式连接

`socks5`模式需要手工设置代理环境变量或配合三方工具使用，执行命令：

```bash
$ ktctl connect --method socks5
```

然后根据日志提示，在任意`CMD`或`PowerShell`命令行中设置`http_proxy`环境变量。譬如在`CMD`中，命令行日志输出：

```
4:31PM INF ----------------------------------------------------------
4:31PM INF Start SOCKS5 Proxy: set http_proxy=socks5://127.0.0.1:2223
4:31PM INF ----------------------------------------------------------
```

打开一个新的`CMD`窗口，设置环境变量：

```bash
$ set http_proxy=socks5://127.0.0.1:2223
```

此时在命令行中就可以用任意可识别该代理环境变量的工具（譬如`curl`）来访问Kubernetes集群里的Pod和Service了。

对于IDEA用户，请参考[在IDEA中使用IDEA](zh-cn/guide/how-to-use-in-idea)。

除了使用环境变量，也可以配合支持Socks5的代理软件（推荐[Proxifier](https://www.proxifier.com)）来开启全局流量代理，实现本地所有软件皆可访问集群Pod和Service的目的。