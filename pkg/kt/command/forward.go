package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/forward"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
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
			return Forward(args)
		},
		Example: "ktctl forward <service-name|remote-address> [<local-port>:<remote-port>] [command options]",
	}

	cmd.SetUsageTemplate(general.UsageTemplate(true))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Forward, opt.ForwardFlags())
	return cmd
}

func Forward(args []string) error {
	ch, err := general.SetupProcess(util.ComponentForward)
	if err != nil {
		return err
	}

	target := args[0]
	localPort, remotePort, err := parsePort(args)
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

func parsePort(args []string) (localPort int, remotePort int, err error) {
	if len(args) < 2 {
		// port not specified
		return -1, -1, nil
	} else if count := strings.Count(args[1], ":"); count == 0 {
		// only local port specified
		localPort, err = strconv.Atoi(args[1])
		if err != nil {
			return -1, -1, fmt.Errorf("port '%s' format invalid", args[1])
		}
	} else if count == 1 {
		// both local port and remote port specified
		parts := strings.Split(args[1], ":")
		localPort, err = strconv.Atoi(parts[0])
		if err != nil {
			return -1, -1, fmt.Errorf("port '%s' format invalid", parts[0])
		}
		remotePort, err = strconv.Atoi(parts[1])
		if err != nil {
			return -1, -1, fmt.Errorf("port '%s' format invalid", parts[1])
		}
	} else {
		return -1, -1, fmt.Errorf("port '%s' format invalid", args[1])
	}
	return localPort, remotePort, nil
}
