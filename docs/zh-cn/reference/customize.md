定制专属KT
---

虽然我们始终认为，作为DevOps实践的一部分，开发者应当对测试环境具有自助访问和管理的权限。然而现实中，由于信息安全和流程管控的原因，有些企业并不能将带有集群资源编辑权限的kubeconfig配置交给每一位开发者。

能否直接将特定的集群权限与`ktctl`命令绑定，在无需分发kubeconfig文件的情况下，实现集群网络打通、服务置换功能呢？

不仅如此，在一些企业中还存在测试环境无法访问公网（需要改ShadowPod镜像地址）、特殊的资源审核规则（如Pod必须有CpuLimit属性）或特殊的资源使用规范（如需统一设定NodeSelector）等情况，使得每位开发者都需要先进行一连串参数配置才能正常使用`ktctl`工具。

能否调整`ktctl`的参数的默认值，让开发者开箱即用无需配置呢？

需要解决这些问题，可以从源码入手，定制一个企业内部的专属KT版本。

## 定制源码

所需工具：[git](https://git-scm.com/downloads)

直接修改源代码实现上述目的并不难，但对于还不那么熟悉Golang的开发者和运维人员来说，KT提供了一种更简单的自定义版本机制。

> 对于需要进行更复杂定制，并且了解潜在的代码冲突风险的同学，可以直接看[编译打包](zh-cn/reference/customize.md?id=编译打包)部分

首先下载`kt-connect`的源码仓库，然后进入代码目录（如果已经下载过源码仓库，直接`git pull`更新即可）：

```bash
git clone https://github.com/alibaba/kt-connect.git
cd kt-connect
```

目前，`ktctl`工具提供了两个快速定制点。在kt-connect代码的"hack"目录的"kt"和"kube"子目录中分别有一个config文件。

```sql
hack
├── customize.go
├── kt
│   └── config   <-- 全局kt-config定制文件
└── kube
    └── config   <-- 全局kube-config定制文件
```

- 封装集群权限：将已配置好权限的kubeconfig文件（默认是用户主目录下的".kube/config"文件）复制到"hack"目录下的"kube"子目录中，覆盖原本的config文件
- 修改ktctl命令参数默认值：通过`ktctl config`生成或手工编辑`ktctl`配置文件（用户主目录下的".kt/config"文件），复制到"hack"目录下的"kt"子目录中，覆盖原本的config文件

这样就完成了权限内置和配置定制化，接下来需要重新打包生成`ktctl`可执行文件。

## 编译打包

所需工具：[go](https://go.dev/dl)、[upx](https://github.com/upx/upx/releases/latest)、[make](https://cmake.org/install/)（可选）

在MacOS和Linux系统下，推荐使用Make工具打包，执行以下`make`命令：

```bash
make mod
TAG=0.3.5 make ktctl
make upx
```

执行后将在`artifacts`目录下一次性生成MacOS/Linux/Windows系统所用的二进制文件。

Windows环境下的Make工具使用起来相对繁琐，建议直接使用`go`和`upx`命令来完成打包，具体命令可以参考[Makefile](https://github.com/alibaba/kt-connect/blob/master/Makefile)中的`ktctl`和`upx`任务。

以打Windows环境的二进制包为例：

<!-- tabs:start -->

#### ** CMD **

```bash
set TAG=0.3.5
set GOARCH=amd64
set GOOS=windows
go mod tidy -compat=1.17
go build -ldflags "-s -w -X main.version=%TAG%" -o artifacts\windows\ktctl.exe .\cmd\ktctl
upx -9 artifacts\windows\ktctl.exe
```

#### ** PowerShell **

```bash
$env:TAG="0.3.5"
$env:GOARCH="amd64"
$env:GOOS="windows"
go mod tidy -compat=1.17
go build -ldflags "-s -w -X main.version=$env:TAG" -o artifacts\windows\ktctl.exe .\cmd\ktctl
upx -9 artifacts\windows\ktctl.exe
```

#### ** MINGW **

```bash
export TAG=0.3.5
export GOARCH=amd64
export GOOS=windows
go mod tidy -compat=1.17
go build -ldflags "-s -w -X main.version=${TAG}" -o artifacts/windows/ktctl.exe ./cmd/ktctl
upx -9 artifacts/windows/ktctl.exe
```

<!-- tabs:end -->

注意：上述命令中的`TAG`变量值建议与kt-connect的最新发行版本保持一致，除非您已经将`global.image`和`mesh.router-image`配置都定制为了企业内部镜像地址，否则使用非正式版本的`TAG`值会导致运行时无法拉取所需的镜像。

生成二进制文件在经过测试验证无误后，就可以分发给开发同学使用啦 : )

> 对于Windows环境，`ktctl`工具需要依赖`wintun.dll`库文件才能正常运行，在打包分发时，建议从`kt-connect`的正式版发行包中提取匹配版本的`wintun.dll`文件并一起成zip打包，便于开发者解包后可以直接使用`ktctl`工具。
