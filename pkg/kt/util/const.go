package util

import "fmt"

const (
	// EnvKubeConfig environment variable for kube config file
	EnvKubeConfig = "KUBECONFIG"

	// KubernetesToolkit name of this tool
	KubernetesToolkit = "kt"
	// ComponentConnect connect command
	ComponentConnect = "connect"
	// ComponentExchange exchange command
	ComponentExchange = "exchange"
	// ComponentMesh mesh command
	ComponentMesh = "mesh"
	// ComponentPreview preview command
	ComponentPreview = "preview"

	// ImageKtShadow default shadow image
	ImageKtShadow = "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow"
	// ImageKtRouter default router image
	ImageKtRouter = "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-router"
	// ImageKtNavigator default navigator image
	ImageKtNavigator = "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-navigator"

	// ConnectModeShuttle sshuttle mode
	ConnectModeShuttle = "sshuttle"
	// ConnectModeTun2Socks tun2socks mode
	ConnectModeTun2Socks = "tun2socks"
	// ExchangeModeScale scale mode
	ExchangeModeScale = "scale"
	// ExchangeModeEphemeral ephemeral mode
	ExchangeModeEphemeral = "ephemeral"
	// ExchangeModeSelector selector mode
	ExchangeModeSelector = "selector"
	// MeshModeAuto auto mode
	MeshModeAuto = "auto"
	// MeshModeManual manual mode
	MeshModeManual = "manual"
	// DnsModeLocalDns local dns mode
	DnsModeLocalDns = "localDNS"
	// DnsModePodDns pod dns mode
	DnsModePodDns = "podDNS"
	// DnsModeHosts hosts dns mode
	DnsModeHosts = "hosts"
	// DnsOrderCluster proxy to cluster dns
	DnsOrderCluster = "cluster"
	// DnsOrderUpstream proxy to upstream dns
	DnsOrderUpstream = "upstream"

	// ControlBy label used for mark shadow pod
	ControlBy = "control-by"
	// KtTarget label used for service selecting shadow or route pod
	KtTarget = "kt-target"
	// KtRole label used for auto mesh roles
	KtRole = "kt-role"
	// KtConfig annotation used for clean up context
	KtConfig = "kt-config"
	// KtUser annotation used for record independent username
	KtUser = "kt-user"
	// KtSelector annotation used for record service origin selector
	KtSelector = "kt-selector"
	// KtRefCount annotation used for count of shared pod / service
	KtRefCount = "kt-ref-count"
	// KtLastHeartBeat annotation used for timestamp of last heart beat
	KtLastHeartBeat = "kt-last-heart-beat"
	// KtLock annotation used for avoid auto mesh conflict
	KtLock = "kt-lock"

	// PostfixRsaKey postfix of local private key name
	PostfixRsaKey = ".key"
	// RouterBin path to router executable
	RouterBin = "/usr/sbin/router"
	// SshBitSize ssh bit size
	SshBitSize = 2048
	// SshAuthKey auth key name
	SshAuthKey = "authorized"
	// SshAuthPrivateKey ssh private key
	SshAuthPrivateKey = "privateKey"
	// DefaultNamespace default namespace
	DefaultNamespace = "default"
	// KtExchangeContainer name of exchange ephemeral container
	KtExchangeContainer = "kt-exchange"
	// DefaultContainer default container name
	DefaultContainer = "standalone"
	// StuntmanServiceSuffix suffix of stuntman service name
	StuntmanServiceSuffix = "-kt-stuntman"
	// RouterPodSuffix suffix of router pod name
	RouterPodSuffix = "-kt-router"
	// ExchangePodInfix exchange pod name
	ExchangePodInfix = "-kt-exchange-"
	// MeshPodInfix mesh pod and mesh service name
	MeshPodInfix = "-kt-mesh-"
	// RectifierPodPrefix rectifier pod name
	RectifierPodPrefix = "kt-rectifier-"
	// RoleConnectShadow shadow role
	RoleConnectShadow = "shadow-connect"
	// RoleExchangeShadow shadow role
	RoleExchangeShadow = "shadow-exchange"
	// RoleMeshShadow shadow role
	RoleMeshShadow = "shadow-mesh"
	// RolePreviewShadow shadow role
	RolePreviewShadow = "shadow-preview"
	// RoleRouter router role
	RoleRouter = "router"
	// TunNameWin tun device name in windows
	TunNameWin = "KtConnectTunnel"
	// TunNameLinux tun device name in linux
	TunNameLinux = "kt0"
	// TunNameMac tun device name in MacOS
	TunNameMac = "utun"
	// AlternativeDnsPort alternative port for local dns
	AlternativeDnsPort = 10053

	// ResourceHeartBeatIntervalMinus interval of resource heart beat
	ResourceHeartBeatIntervalMinus = 2
	// PortForwardHeartBeatIntervalSec interval of port-forward heart beat
	PortForwardHeartBeatIntervalSec = 60

)

var (
	KtHome = fmt.Sprintf("%s/.kt", UserHome)
	KtKeyDir = fmt.Sprintf("%s/key", KtHome)
	KtPidDir = fmt.Sprintf("%s/pid", KtHome)
	KtLockDir = fmt.Sprintf("%s/lock", KtHome)
	KtProfileDir = fmt.Sprintf("%s/profile", KtHome)
	KtConfigFile = fmt.Sprintf("%s/config", KtHome)
)