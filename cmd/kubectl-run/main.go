package main

import (
	"os"

	"github.com/alibaba/kt-connect/pkg/kt/cmd"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	version = "dev"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-mesh", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := cmd.NewRunCommand(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, version)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
