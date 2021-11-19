package common

const (
	EnvVarLocalDomains      = "LOCAL_DOMAIN"
	ClientTunIP             = "CLIENT_TUN_IP"
	ServerTunIP             = "SERVER_TUN_IP"
	TunMaskLength           = "TUN_MASK_LEN"
	KubernetesTool          = "kt"
	ComponentConnect        = "connect"
	ComponentExchange       = "exchange"
	ComponentMesh           = "mesh"
	ComponentProvide        = "provide"
	ConnectMethodVpn        = "vpn"
	ConnectMethodTun        = "tun"
	ConnectMethodSocks      = "socks"
	ConnectMethodSocks5     = "socks5"
	ExchangeMethodScale     = "scale"
	ExchangeMethodEphemeral = "ephemeral"
	MeshMethodAuto          = "auto"
	MeshMethodManual        = "manual"
	PostfixRsaKey           = "_id_rsa"
	YyyyMmDdHhMmSs          = "2006-01-02 15:04:05"
	SshPort                 = 22
	Socks4Port              = 1080

	// ControlBy label used for mark shadow pod
	ControlBy = "control-by"
	// KtVersion label used for fetch shadow mark in UI
	KtVersion = "kt-version"
	// KtComponent label used for distinguish shadow type
	KtComponent = "kt-component"
	// KtRemoteAddress label used for fetch pod IP in UI
	KtRemoteAddress = "kt-remote-address"
	// KtName label used for wait shadow pod ready
	KtName = "kt-name"
	// KtRole label used for auto mesh roles
	KtRole = "kt-role"
	// KtConfig annotation used for clean up context
	KtConfig = "kt-config"
	// KtUser annotation used for record independent username
	KtUser = "kt-user"
	// KtSelector label used for record service origin selector
	KtSelector = "kt-selector"
	// KtRefCount annotation used for count of shared pod / service
	KtRefCount = "kt-ref-count"
	// KtLastHeartBeat annotation used for timestamp of last heart beat
	KtLastHeartBeat = "kt-last-heart-beat"
	// KtLock annotation used for avoid auto mesh conflict
	KtLock = "kt-lock"

	// SshPrivateKeyName ssh private key name
	SshPrivateKeyName = "kt_%s" + PostfixRsaKey
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
	// KtMeshInfix of mesh pod and mesh service name
	KtMeshInfix = "-kt-mesh-"
	// RoleShadow shadow role
	RoleShadow = "shadow"
	// RoleRouter router role
	RoleRouter = "router"
)

var (
	// AllKtComponents kt commands available
	AllKtComponents = [4]string{"connect", "exchange", "mesh", "provide"}
)
