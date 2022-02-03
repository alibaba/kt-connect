package common

const (
	KubernetesTool        = "kt"
	ComponentConnect      = "connect"
	ComponentExchange     = "exchange"
	ComponentMesh         = "mesh"
	ComponentPreview      = "preview"
	ConnectModeShuttle    = "sshuttle"
	ConnectModeTun2Socks  = "tun2socks"
	ExchangeModeScale     = "scale"
	ExchangeModeEphemeral = "ephemeral"
	ExchangeModeSelector  = "selector"
	MeshModeAuto          = "auto"
	MeshModeManual        = "manual"
	DnsModeLocalDns       = "localDNS"
	DnsModePodDns         = "podDNS"
	DnsModeHosts          = "hosts"
	Localhost             = "127.0.0.1"
	YyyyMmDdHhMmSs        = "2006-01-02 15:04:05"
	SshPort               = 22

	// EnvVarLocalDomains environment variable for local domain config
	EnvVarLocalDomains = "KT_LOCAL_DOMAIN"
	// EnvVarDnsProtocol environment variable for shadow pod dns protocol
	EnvVarDnsProtocol = "KT_DNS_PROTOCOL"
	// EnvVarLogLevel environment variable for shadow pod log level
	EnvVarLogLevel = "KT_LOG_LEVEL"
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
	PostfixRsaKey = "_id_rsa"
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
	// OriginServiceSuffix suffix of origin service name
	OriginServiceSuffix = "-kt-origin"
	// RouterPodSuffix suffix of router pod name
	RouterPodSuffix = "-kt-router"
	// ExchangePodInfix exchange pod name
	ExchangePodInfix = "-kt-exchange-"
	// MeshPodInfix mesh pod and mesh service name
	MeshPodInfix = "-kt-mesh-"
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
	// StandardDnsPort standard dns port
	StandardDnsPort = 53
	// AlternativeDnsPort alternative port for local dns
	AlternativeDnsPort = 10053
)
