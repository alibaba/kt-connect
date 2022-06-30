package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/forward"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

// NewForwardCommand return new Forward command
func NewForwardCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forward",
		Short: "Redirect local port to a service or any remote address",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("a service name or target address must be specified")
			} else if len(args) == 1 && strings.Contains(args[0], ".") {
				return fmt.Errorf("a port must be specified because '%s' is not a service name", args[0])
			} else if len(args) > 2 {
				return fmt.Errorf("too many target addresses are spcified (%s)", strings.Join(args, ",") )
			}
			opt.Get().Global.UseLocalTime = true
			return general.Prepare()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return forwardBySvcPort(args[0], "")
			} else if !strings.Contains(args[1], ":") {
				return forwardBySvcPort(args[0], args[1])
			} else {
				return forwardTo(args[0], args[1])
			}
		},
		Example: "ktctl forward <service-name|remote-address> [<local-port>:<remote-port>] [command options]",
	}

	cmd.SetUsageTemplate(general.UsageTemplate(true))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Forward, opt.ForwardFlags())
	return cmd
}

func forwardBySvcPort(svcName, port string) error {
	svc, err := cluster.Ins().GetService(svcName, opt.Get().Global.Namespace)
	if err != nil {
		return err
	}
	if len(svc.Spec.Ports) == 0 {
		return fmt.Errorf("service '%s' has not port available", svcName)
	} else if len(svc.Spec.Ports) > 1 {
		return fmt.Errorf("service '%s' has multiple ports, must specify one", svcName)
	}
	remotePort := int(svc.Spec.Ports[0].Port)
	var localPort int
	if len(port) == 0 {
		localPort = remotePort
	} else {
		localPort, err = strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("port '%s' format invalid", port)
		}
	}
	return forwardFromTo(svcName, localPort, remotePort)
}

func forwardTo(target, port string) (err error) {
	var localPort, remotePort int
	if count := strings.Count(port, ":"); count == 1 {
		parts := strings.Split(port, ":")
		localPort, err = strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("port '%s' format invalid", parts[0])
		}
		remotePort, err = strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("port '%s' format invalid", parts[1])
		}
	} else {
		return fmt.Errorf("port '%s' format invalid", port)
	}
	return forwardFromTo(target, localPort, remotePort)
}

func forwardFromTo(target string, localPort, remotePort int) error {
	ch, err := general.SetupProcess(util.ComponentForward)
	if err != nil {
		return err
	}

	if strings.Contains(target, ".") {
		err = forward.RedirectAddress(target, localPort, remotePort)
		if err != nil {
			return err
		}
		log.Info().Msg("---------------------------------------------------------------")
		log.Info().Msgf(" Now you can access to '%s:%d' via 'localhost:%d'", target, remotePort, localPort)
		log.Info().Msg("---------------------------------------------------------------")
	} else {
		err = forward.RedirectService(target, localPort, remotePort)
		if err != nil {
			return err
		}
		log.Info().Msg("---------------------------------------------------------------")
		log.Info().Msgf(" Now you can access port %d of service '%s' via 'localhost:%d'", remotePort, target, localPort)
		log.Info().Msg("---------------------------------------------------------------")
	}

	// watch background process, clean the workspace and exit if background process occur exception
	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)
	return nil
}
