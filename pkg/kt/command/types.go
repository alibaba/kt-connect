package command

import (
	"github.com/alibaba/kt-connect/pkg/kt"
)

// ActionInterface all action defined
type ActionInterface interface {
	Connect(cli kt.CliInterface) error
	Preview(serviceName string, cli kt.CliInterface) error
	Exchange(resourceName string, cli kt.CliInterface) error
	Mesh(deploymentName string, cli kt.CliInterface) error
	Clean(cli kt.CliInterface) error
}

// Action cmd action
type Action struct {}
