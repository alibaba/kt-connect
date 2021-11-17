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
	// KTVersion label used for fetch shadow mark in UI
	KTVersion = "kt-version"
	// KTComponent label used for distinguish shadow type
	KTComponent = "kt-component"
	// KTRemoteAddress label used for fetch pod IP in UI
	KTRemoteAddress = "kt-remote-address"
	// KTName label used for wait shadow pod ready
	KTName = "kt-name"
	// KTRole label used for auto mesh roles
	KTRole = "kt-role"
	// KTConfig annotation used for clean up context
	KTConfig = "kt-config"
	// KtUser annotation used for record independent username
	KtUser = "kt-user"
	// KtSelector label used for record service origin selector
	KtSelector = "kt-selector"
	// KTRefCount annotation used for count of shared pod / service
	KTRefCount = "kt-ref-count"
	// KTLastHeartBeat annotation used for timestamp of last heart beat
	KTLastHeartBeat = "kt-last-heart-beat"

	// SSHPrivateKeyName ssh private key name
	SSHPrivateKeyName = "kt_%s" + PostfixRsaKey
	// SSHBitSize ssh bit size
	SSHBitSize = 2048
	// SSHAuthKey auth key name
	SSHAuthKey = "authorized"
	// SSHAuthPrivateKey ssh private key
	SSHAuthPrivateKey = "privateKey"
	// DefaultNamespace default namespace
	DefaultNamespace = "default"
	// KtExchangeContainer name of exchange ephemeral container
	KtExchangeContainer = "kt-exchange"
	// DefaultContainer default container name
	DefaultContainer = "standalone"
	// OriginServiceSuffix suffix of origin service name
	OriginServiceSuffix = "-kt-origin"
)

var (
	// AllKtComponents kt commands available
	AllKtComponents = [4]string{"connect", "exchange", "mesh", "provide"}
)
