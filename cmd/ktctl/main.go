package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alibaba/kt-connect/pkg/kt/action"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	// global params
	kubeconfig string
	namespace  string
	debug      bool
	image      string

	// connect
	disableDNS   bool
	localSSHPort int
	cidr         string

	// exchange
	expose string

	//context
	pidFile  string
	userHome string
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	userHome := util.HomeDir()
	appHome := fmt.Sprintf("%s/.ktctl", userHome)
	util.CreateDirIfNotExist(appHome)
	pidFile = fmt.Sprintf("%s/pid", appHome)

	app := cli.NewApp()
	app.Name = "KT Connect"
	app.Usage = ""
	app.Version = "0.0.6"
	app.Authors = []cli.Author{
		cli.Author{
			Name: "rdc incubator",
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "namespace,n",
			Value:       "default",
			Destination: &namespace,
		},
		cli.StringFlag{
			Name:        "kubeconfig,c",
			Value:       filepath.Join(userHome, ".kube", "config"),
			Destination: &kubeconfig,
		},
		cli.StringFlag{
			Name:        "image,i",
			Usage:       "Custom proxy image",
			Value:       "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:stable",
			Destination: &image,
		},
		cli.BoolFlag{
			Name:        "debug,d",
			Usage:       "debug mode",
			Destination: &debug,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "connect",
			Usage: "connection to kubernetes cluster",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:        "port",
					Value:       2222,
					Usage:       "Local SSH Proxy port",
					Destination: &localSSHPort,
				},
				cli.BoolFlag{
					Name:        "disableDNS",
					Usage:       "Disable Cluster DNS",
					Destination: &disableDNS,
				},
				cli.StringFlag{
					Name:        "cidr",
					Usage:       "Custom CIDR eq '172.2.0.0/16'",
					Destination: &cidr,
				},
			},
			Action: func(c *cli.Context) error {
				action := action.Action{
					Kubeconfig: kubeconfig,
					Namespace:  namespace,
					Debug:      debug,
					Image:      image,
					PidFile:    pidFile,
					UserHome:   userHome,
				}
				if debug {
					zerolog.SetGlobalLevel(zerolog.DebugLevel)
				}
				action.Connect(localSSHPort, disableDNS, cidr)
				return nil
			},
		},
		{
			Name:  "exchange",
			Usage: "exchange kubernetes deployment to local",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "expose",
					Usage:       "expose port",
					Destination: &expose,
				},
			},
			Action: func(c *cli.Context) error {
				action := action.Action{
					Kubeconfig: kubeconfig,
					Namespace:  namespace,
					Debug:      debug,
					Image:      image,
					PidFile:    pidFile,
					UserHome:   userHome,
				}
				if debug {
					zerolog.SetGlobalLevel(zerolog.DebugLevel)
				}
				action.Exchange(c.Args().First(), expose, userHome, pidFile)
				return nil
			},
		},
		{
			Name:  "mesh",
			Usage: "mesh kubernetes deployment to local",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "expose",
					Usage:       "expose port",
					Destination: &expose,
				},
			},
			Action: func(c *cli.Context) error {
				action := action.Action{
					Kubeconfig: kubeconfig,
					Namespace:  namespace,
					Debug:      debug,
					Image:      image,
					PidFile:    pidFile,
					UserHome:   userHome,
				}
				if debug {
					zerolog.SetGlobalLevel(zerolog.DebugLevel)
				}
				action.Mesh(c.Args().First(), expose, userHome, pidFile)
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err)
	}
}
