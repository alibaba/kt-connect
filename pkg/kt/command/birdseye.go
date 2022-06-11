package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/birdseye"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
)

// NewBirdseyeCommand show a summary of cluster service network
func NewBirdseyeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "birdseye",
		Short: "Show summary of services status in cluster",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("too many options specified (%s)", strings.Join(args, ",") )
			}
			return general.Prepare()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Birdseye()
		},
		Example: "ktctl birdseye [command options]",
	}

	cmd.SetUsageTemplate(general.UsageTemplate(false))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Birdseye, opt.BirdseyeFlags())
	return cmd
}

func Birdseye() error {
	pods, apps, svcs, err := birdseye.GetKtPodsAndAllServices(opt.Get().Global.Namespace)
	if err != nil {
		return err
	}

	if opt.Get().Birdseye.ShowConnector {
		birdseye.ShowConnectors(pods, apps)
	}

	// service-name, service-description
	allServices := make([][]string, 0)
	for _, svc := range svcs {
		if cb, exists := svc.Labels[util.ControlBy]; exists && cb == util.KubernetesToolkit {
			for _, p := range pods {
				if util.MapContains(svc.Spec.Selector, p.Labels) {
					if role := p.Labels[util.KtRole]; role == util.RolePreviewShadow {
						allServices = append(allServices, []string{svc.Name, "previewing"})
						continue
					} else if role == util.RoleExchangeShadow {
						user := p.Annotations[util.KtUser]
						if user == "" {
							user = "unknown user"
						}
						allServices = append(allServices, []string{svc.Name, "exchanged by " + user})
						continue
					} else if role == util.RoleMeshShadow {
						allServices = append(allServices, []string{svc.Name, "meshed"})
						continue
					}
				}
			}
		} else if !opt.Get().Birdseye.HideNaturalService {
			allServices = append(allServices, []string{svc.Name, "normal"})
		}
	}
	for _, svc := range allServices {
		log.Info().Msgf("%s - %s", svc[0], svc[1])
	}
	return nil
}
