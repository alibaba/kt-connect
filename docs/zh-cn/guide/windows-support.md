Windows支持
---

KtConnect目前只支持在Windows环境使用基于Socks代理的连接方法，分为`socks`和`socks5`两种模式。

## 使用`socks`模式连接

`socks`是KtConnect在Windows环境的默认工作模式，无需增加额外参数：

```bash
$ ktctl connect
```

此模式下，KtConnect会启动一个连接Kubernetes集群的Socks4协议代理进程，本地服务通过该代理即可访问集群中的IP地址。

在Windows下让进程自动使用代理访问网络的方法因程序而异，两种相对常见的方法是`http_proxy`环境变量和系统代理设置。

`http_proxy`环境变量通常对各类命令行工具有效（但并非是所有命令行工具），如`curl`和`wget`等，此方法对`socks`和`socks5`模式都适用。

系统代理仅适用于`socks`模式，对浏览器以及部分系统内置服务有效，具体设置方法在各Windows版本中稍有差异，例如Windows 10可参考[这篇文章](https://zhuanlan.zhihu.com/p/111541987)，应将代理地址设置为`socks=127.0.0.1`，端口设置为`2223`。

## 使用`socks5`模式连接

`socks5`模式需使用`--method socks5`参数开启：

```bash
$ ktctl connect --method socks5
```

它的工作原理与`socks`模式类似，同样需采用例如设置`http_proxy`环境变量等方法，让目标应用程序使用代理访问集群。

需注意的是，Windows的系统代理设置不支持`socks5`模式的代理。在主流浏览器中，只有Firefox内置了Socks5代理协议功能。

不过，市面上有一些全局流量代理工具（譬如[Proxifier](https://www.proxifier.com)）也能够支持Socks5协议，从而使本地所有软件皆可访问集群中的IP和域名。

## 自动设置全局代理

默认情况下，使用`socks`或`socks5`模式运行`ktctl connect`命令时，会在命令行（CMD或PowerShell）输出用于设置`http_proxy`环境变量的命令。譬如在`CMD`中，显示内容如下：

```
00:00PM INF ----------------------------------------------------------
00:00PM INF Start SOCKS5 Proxy: set http_proxy=socks5://127.0.0.1:2223
00:00PM INF ----------------------------------------------------------
```

除了手工设置该环境变量，也可以使用`--setupGlobalProxy`参数，让KtConnect自动将其设置为全局环境变量：

```bash
> ktctl connect --setupGlobalProxy   # 或 ktctl connect --method socks5 --setupGlobalProxy
```

此时，控制台将不再显示前述的手工设置环境变量命令。在`socks`模式下，KtConnect还会自动将Windows的系统代理也配置为其本地代理服务地址。

由于在Windows里，环境变量和系统代理的修改对已运行的进程无效，因此只有在KtConnect运行以后再启动的浏览器、控制台等进程才会自动通过该代理获得访问集群Pod和Service地址的能力。

当KtConnect进程正常结束时，会自动还原系统代理配置和环境变量，若遇到KtConnect进程异常关闭而未能恢复系统配置时，可以事后手工执行`ktctl clean`命令进行清理。

> 注意：在Windows下，除非使用某些第三方全局代理软件，否则无法真正做到让所有程序均通过代理访问网络。
> Windows内置的"全局代理配置"并不对所有应用程序都具有约束性，仅对浏览器和部分系统软件有效。
> 对于其他情况，需查看软件自身是否提供有代理配置功能。譬如进行基于Spring的Java应用开发时，可参考[在IDEA中联调](zh-cn/guide/how-to-use-in-idea.md)文档。

## 两种连接模式的对比

| 连接模式 | 代理协议 | 是否可用命令行工具 | 是否可用于浏览器 | 主要优点       |
| ------ | ------- | --------------- | -------------- | ------------ |
| `socks`  | Socks4  | 是，基于http_proxy环境变量 | 是，任意浏览器 | Windows内置Socks4代理支持，可直接用于各种浏览器 |
| `socks5` | Socks5  | 是，基于http_proxy环境变量 | 是，仅Firefox | 可配合三方工具（如Proxifier）实现真正的全局代理 |
