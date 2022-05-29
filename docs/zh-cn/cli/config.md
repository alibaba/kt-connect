Ktctl Config
---

用于预先设定`ktctl`命令参数的默认值。包含8个子命令：

- `show`：查看所有已配置的参数
- `get`：查看指定配置参数的值
- `set`：为指定参数设置默认值
- `unset`：删除指定参数的默认值
- `list-profile`：列出所有已保存的配置集合
- `save-profile`：将当前配置内容保存为配置集合
- `load-profile`：加载指定的配置集合
- `drop-profile`：删除指定的配置集合

基本用法如下：

```bash
ktctl config show
ktctl config get <参数名>
ktctl config set <参数名> <参数值>
ktctl config unset <参数名>
ktctl config list-profile
ktctl config save-profile <存档名>
ktctl config load-profile <存档名>
ktctl config drop-profile <存档名>
```

参数的配置格式为"<命令>.<参数>"，其中参数部分应使用全小写的横线分隔格式。例如将`ktctl connect`命令的`--excludeIps`值默认设置为`172.2.1.0/24,172.2.2.0/24`，则相应配置命令如下：

```bash
ktctl config set connect.exclude-ips 172.2.1.0/24,172.2.2.0/24
                |<-命令->||<-参数->|  |<---- 期望的默认值 ---->|
```

对于全局参数，"命令"部分固定使用`global`，例如配置ShadowPod的镜像地址：

```bash
ktctl config set global.image your-repository-address/kt-connect-shadow:latest
```

使用带`--all`参数的`config show`命令可以查看所有可配置的参数清单：

```bash
ktctl config show --all
```

配置的内容会以YAML格式存储在用户主目录下的".kt/config"文件里。

`config`命令自身的可选参数如下：

```
config show
--all, -a 列出所有可用参数，包括未配置默认值的参数

config get
无参数

config set
无参数

config unset
--all, -a 删除所有参数的默认配置

config list-profile
无参数

config save-profile
无参数

config load-profile
--dryRun 仅显示配置集合的内容，不加载

config drop-profile
无参数
```
