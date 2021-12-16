package tunnel

// CliInterface ...
type CliInterface interface {
	AddDevice() error
	AddRoute(cidr string) error
	SetDeviceIP() error
	RemoveDevice() error
}

// Cli ...
type Cli struct {
	TunName  string
	SourceIP string
	DestIP   string
	MaskLen  string
}
