package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	urfave "github.com/urfave/cli"
)

// NewRecoverCommand return new recover command
func NewRecoverCommand(action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "recover",
		Usage: "drop back traffic to specified kubernetes service changed by exchange or mesh",
		UsageText: "ktctl recover [command options]",
		Flags: general.RecoverActionFlag(opt.Get()),
		Action: func(c *urfave.Context) error {
			if opt.Get().Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(); err != nil {
				return err
			}

			if len(c.Args()) == 0 {
				return fmt.Errorf("name of service to recover is required")
			}

			return action.Recover(c.Args().First())
		},
	}
}

// Recover delete unavailing shadow pods
func (action *Action) Recover(serviceName string) error {
	return nil
}

