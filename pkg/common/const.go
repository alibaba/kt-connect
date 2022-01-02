package common

const (
	EnvVarLocalDomains      = "LOCAL_DOMAIN"
	KubernetesTool          = "kt"
	ComponentConnect        = "connect"
	ComponentExchange       = "exchange"
	ComponentMesh           = "mesh"
	ComponentProvide        = "provide"
	ConnectMethodShuttle    = "sshuttle"
	ConnectMethodTun        = "tun"
	ExchangeMethodScale     = "scale"
	ExchangeMethodEphemeral = "ephemeral"
	ExchangeMethodSwitch    = "switch"
	MeshMethodAuto          = "auto"
	MeshMethodManual        = "manual"
	YyyyMmDdHhMmSs          = "2006-01-02 15:04:05"
	SshPort                 = 22

	// ControlBy label used for mark shadow pod
	ControlBy = "control-by"
	// KtName label used for wait shadow pod ready
	KtName = "kt-name"
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
	// RoleProvideShadow shadow role
	RoleProvideShadow = "shadow-provide"
	// RoleRouter router role
	RoleRouter = "router"
)
