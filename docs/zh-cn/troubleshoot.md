Connect故障排查
---

本文将通过手动的方式模拟本地与集群之间建立连接的过程

## 在集群中手动部署Shadow容器和验证应用

> 注意，如果使用了自定义shadow镜像请将image替换为自己的镜像即可

```
$ kubectl run nginx --image=nginx
$ kubectl run troubleshoot --image=registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:stable
```

等待troubleshoot容器启动成功

```
$ kubectl get pods -o wide | grep troubleshoot          
NAME                                                              READY   STATUS             RESTARTS   AGE    IP             NODE                                NOMINATED NODE
troubleshoot-6f9958b7f7-g8vgs                                     1/1     Running            0          3m     172.16.1.114
```

## 将Shodw容器的22端口映射到本地2222

> 当前步骤是为了将集群的ssh协议端口转发到本地端口

```
$ kubectl port-forward pod/troubleshoot-6f9958b7f7-g8vgs 2222:22
Forwarding from 127.0.0.1:2222 -> 22
Forwarding from [::1]:2222 -> 22
```

## 建立SSH免密认证

> troubleshoot容器的默认账号是root:root

```
$ ssh-copy-id root@127.0.0.1 -p 2222
```

上传公钥后，验证本地是否能够正常通过ssh登录到root@127.0.0.1 -p 2222

## 验证本地到集群网络通道

<!-- tabs:start -->

#### ** Socks5模式 **

建立socks5连接：

```
$ ssh -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -i ~/.ssh/id_rsa -D 5000 root@127.0.0.1 -p 2222
```

在新的termial中，设置http_proxy并验证连通性:

```
$ export http_proxy=socks5://127.0.0.1:5000
$ curl http://<nginx-pod-ip>:80
```

#### ** sshuttle模式 **

```
sudo sshuttle -v -e "ssh -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -i /Users/user/.ssh/id_rsa" -r root@127.0.0.1:2222 -x 127.0.0.1 172.16.0.0/24
```

* -i为指定本地的私钥的绝对路径
* -x为需要vpn代理的网段，这里以nginx测试容器的podIP所在子网为例

验证连通性：

```
$ curl http://<nginx-pod-ip>:80
```

<!-- tabs:end -->


