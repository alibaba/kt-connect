package common

const (
	EnvVarLocalDomain   = "LOCAL_DOMAIN"
	ClientTunIP         = "CLIENT_TUN_IP"
	ServerTunIP         = "SERVER_TUN_IP"
	TunMaskLength       = "TUN_MASK_LEN"
	ControlBy           = "control-by"
	KubernetesTool      = "kt"
	ComponentConnect    = "connect"
	ComponentExchange   = "exchange"
	ComponentMesh       = "mesh"
	ComponentProvide    = "provide"
	ConnectMethodVpn    = "vpn"
	ConnectMethodTun    = "tun"
	ConnectMethodSocks  = "socks"
	ConnectMethodSocks5 = "socks5"
	PostfixRsaKey       = "_id_rsa"
	KTLastHeartBeat     = "kt-last-heart-beat"
	YyyyMmDdHhMmSs      = "2006-01-02 15:04:05"
	SshPort             = 22
	Socks4Port          = 1080

	// KTVersion label used for fetch shadow mark in UI
	KTVersion = "kt-version"
	// KTComponent label used for distinguish shadow type
	KTComponent = "kt-component"
	// KTRemoteAddress label used for fetch pod IP in UI
	KTRemoteAddress = "kt-remote-address"
	// KTName label used for wait shadow pod ready
	KTName = "kt-name"
	// KTConfig annotation used for clean up context
	KTConfig = "kt-config"

	// SSHPrivateKeyName ssh private key name
	SSHPrivateKeyName = "kt_%s" + PostfixRsaKey
	// SSHBitSize ssh bit size
	SSHBitSize = 2048
	// SSHAuthKey auth key name
	SSHAuthKey = "authorized"
	// SSHAuthPrivateKey ssh private key
	SSHAuthPrivateKey = "privateKey"
	// RefCount the count of shared
	RefCount = "refCount"
	// DefNamespace default namespace
	DefNamespace = "default"
)

var (
	// AllKtComponents kt commands available
	AllKtComponents = [4]string{"connect", "exchange", "mesh", "provide"}
)
