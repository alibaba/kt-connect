Customized KT
---

As part of DevOps practices, we always believe that developers should have self-service access to manage test environments. However, in reality, due to information security and process control reasons, some companies cannot hand over the kubeconfig configuration with cluster resource editing permissions to every developer.

Is it possible to directly bind a specific cluster permission to the `ktctl` command to achieve cluster network connectivity and service replacement without distributing kubeconfig files ?

Not only that, in some enterprises, the test environments may not able to access the public network (so the shadow pod image address needs to be changed), or have special resource audit rules (e.g. pod must have CpuLimit attribute), or special resource usage specifications (e.g. each pod must have NodeSelector uniformly) ... In those cases, every developer needs to configure a series of parameters before they can use the `ktctl` tool normally.

Can the default values of the parameters of `ktctl` be preset so that developers can use it out of the box without configuration ?

To solve these problems, you can start from the source code and customize an exclusive KT version within the enterprise.

## Custom source code

Required tools: [git](https://git-scm.com/downloads)

It is not difficult to directly modify the source code to achieve the above purpose, but for developers and operators who are not so familiar with Golang, KT provides an easy-to-start customization mechanism.

> For people who need to make more complex customizations and understand the potential risk of code conflict, you can directly see the [Compile and Package](en-us/reference/customize.md?id=Compile and package) section

Firstly, download the source code repository of `kt-connect`, and then enter the code directory (if you have already downloaded the source code repository, just update it with `git pull`):

```bash
git clone https://github.com/alibaba/kt-connect.git
cd kt-connect
````

Currently, the `ktctl` tool provides two quick customization points. There is a config file in both "kt" and "kube" subdirectories of the "hack" directory of the kt-connect code.

```sql
hack
├── customize.go
├── kt
│   └── config  <-- global kt-config custom file
└── kube
    └── config  <-- global kube-config custom file
````

- Encapsulate cluster permissions: Copy the kubeconfig file with configured permissions (the ".kube/config" file in the user's home directory by default) to the "kube" subdirectory under the "hack" directory, overwriting the original one
- Modify the default value of ktctl command parameters: generate or manually edit the `ktctl` configuration file (".kt/config" file in the user's home directory) through `ktctl config`, copy it to the "kt" subdirectory under the "hack" directory , overwrite the original one

This completes the built-in permissions and configuration customization. Next, you need to repackage and generate the `ktctl` executable file.

## Compile and package

Required tools: [go](https://go.dev/dl), [upx](https://github.com/upx/upx/releases/latest), [make](https://cmake.org/install/) (optional)

Under MacOS and Linux systems, it is recommended to use the Make toolkit, with following `make` command:

```bash
make mod
TAG=0.3.5 make ktctl
make upx
````

After execution, binary files for MacOS/Linux/Windows systems will be generated in the `artifacts` directory at one time.

The Make toolkit in the Windows environment is relatively cumbersome to use. It is recommended to use the `go` and `upx` commands to complete packaging directly. For specific commands, please refer to the `ktctl` and `upx` tasks in [Makefile](https://github.com/alibaba/kt-connect/blob/master/Makefile).

For example, compile the binary execution file of Windows 64bit environment:

<!-- tabs:start -->

####**cmd**

```bash
set TAG=0.3.5
set GOARCH=amd64
set GOOS=windows
go mod tidy -compat=1.17
go build -ldflags "-s -w -X main.version=%TAG%" -o artifacts\windows\ktctl.exe .\cmd\ktctl
upx -9 artifacts\windows\ktctl.exe
````

#### **PowerShell**

```bash
$env:TAG="0.3.5"
$env:GOARCH="amd64"
$env:GOOS="windows"
go mod tidy -compat=1.17
go build -ldflags "-s -w -X main.version=$env:TAG" -o artifacts\windows\ktctl.exe .\cmd\ktctl
upx -9 artifacts\windows\ktctl.exe
````

####**MINGW**

```bash
export TAG=0.3.5
export GOARCH=amd64
export GOOS=windows
go mod tidy -compat=1.17
go build -ldflags "-s -w -X main.version=${TAG}" -o artifacts/windows/ktctl.exe ./cmd/ktctl
upx -9 artifacts/windows/ktctl.exe
````

<!-- tabs:end -->

Note: The value of the `TAG` variable in the above command is recommended to be consistent with the latest release version of kt-connect, unless you have customized both `global.image` and `mesh.router-image` configurations to the internal image address of the enterprise, otherwise using an unofficial version of the `TAG` value will cause `ktctl` fail to pull required image.

After the generated binary file has been tested and verified, it can be directly distributed to the developer for use : )

> For Windows environment, the `ktctl` tool is relying on the `wintun.dll` library file to run normally. When packaged and distributed, it is recommended to extract the corresponding version of the `wintun.dll` file from the official release package of `kt-connect` and put it together with `ktctl` binary into a zip, so that developers can use the tool directly after unpacking.