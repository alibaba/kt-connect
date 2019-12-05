Windows支持
==============

## 使用Windows Subsystem for Linux (WSL)

KT Connect 为了能够在Windows下使用KT Connection, 您可以尝试使用Windows Subsystem for Linux。

请根据帮助文档：https://docs.microsoft.com/en-us/windows/wsl/install-win10 在Windows 10操作系统下安装Ubuntu子系统。 在安装完成后就可以像在Linux中一样使用KT Connect。 限制WSL环境中只能使用socks5代理，同时不支持DNS解析

使用方式:

```
$ sudo ktctl -d connect --method socks5
4:31PM INF Daemon Start At 56071
4:31PM INF Client address 30.5.125.75
4:31PM INF Deploying shadow deployment kt-connect-daemon-cusdp in namespace default

4:31PM DBG Shadow Pod status is Pending
4:31PM DBG Shadow Pod status is Running
4:31PM INF Shadow is ready.
4:31PM DBG Success deploy proxy deployment kt-connect-daemon-cusdp in namespace default

4:31PM DBG Child, os.Args = [ktctl -d connect --method socks5]
4:31PM DBG Child, cmd.Args = [kubectl --kubeconfig=/Users/yunlong/.kube/config -n default port-forward kt-connect-daemon-cusdp-6d47d4f594-p9m6j 2222:22]
Forwarding from 127.0.0.1:2222 -> 22
Forwarding from [::1]:2222 -> 22
4:31PM DBG port-forward start at pid: 56114
4:31PM INF ==============================================================
4:31PM INF Start SOCKS5 Proxy: export http_proxy=socks5://127.0.0.1:2223
4:31PM INF ==============================================================
4:31PM DBG Child, os.Args = [ktctl -d connect --method socks5]
4:31PM DBG Child, cmd.Args = [ssh -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -i /Users/yunlong/.kt_id_rsa -D 2223 root@127.0.0.1 -p2222 sh loop.sh]
Handling connection for 2222
Warning: Permanently added '[127.0.0.1]:2222' (ECDSA) to the list of known hosts.
4:31PM DBG vpn(ssh) start at pid: 56190
4:31PM DBG KT proxy start successful
```

设置http_proxy代理，后就可以直接在本地访问集群POD和Service的地址

```
# set http_proxy
$ export http_proxy=socks5://127.0.0.1:2223
$ curl http://<POD_IP>:<PORT>
$ curl http://<CLUSTER_IP>:<PORT>
```