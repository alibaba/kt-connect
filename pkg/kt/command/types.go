package command

// ActionInterface all action defined
type ActionInterface interface {
	Connect() error
	Preview(serviceName string) error
	Exchange(resourceName string) error
	Mesh(resourceName string) error
	Recover(serviceName string) error
	Clean() error
}

// Action cmd action
type Action struct {}
