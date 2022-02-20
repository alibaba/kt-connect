package command

import "os"

// ActionInterface all action defined
type ActionInterface interface {
	Connect(ch chan os.Signal) error
	Preview(serviceName string, ch chan os.Signal) error
	Exchange(resourceName string, ch chan os.Signal) error
	Mesh(resourceName string, ch chan os.Signal) error
	Recover(serviceName string) error
	Clean() error
}

// Action cmd action
type Action struct {}
