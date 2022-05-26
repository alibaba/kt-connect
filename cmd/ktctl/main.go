package main

import (
	"github.com/alibaba/kt-connect/pkg/kt/command"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"os"
)

var (
	version = "dev"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: util.IsWindows()})
	for _, dir := range []string{util.KtKeyDir, util.KtPidDir, util.KtLockDir, util.KtProfileDir} {
		_ = util.CreateDirIfNotExist(dir)
		_ = util.FixFileOwner(dir)
	}
	_ = util.FixFileOwner(util.KtConfigFile)
	// TODO: 0.4 - auto remove old kt home folder .ktctl
}

func main() {
	// this line must go first
	opt.Store.Version = version
	cobra.EnableCommandSorting = false

	var rootCmd = &cobra.Command{
		Use:   "ktctl",
		Version: version,
		Short: "A utility tool to help you work with Kubernetes dev environment more efficiently",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		Example: "ktctl <command> [command options]",
	}

	rootCmd.AddCommand(command.NewConnectCommand())
	rootCmd.AddCommand(command.NewExchangeCommand())
	rootCmd.AddCommand(command.NewMeshCommand())
	rootCmd.AddCommand(command.NewPreviewCommand())
	rootCmd.AddCommand(command.NewRecoverCommand())
	rootCmd.AddCommand(command.NewCleanCommand())
	rootCmd.AddCommand(command.NewConfigCommand())
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.SetUsageTemplate(general.UsageTemplate(false))
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	opt.SetOptions(rootCmd, rootCmd.PersistentFlags(), opt.Get().Global, opt.GlobalFlags())

	// process will hang here
	if err := rootCmd.Execute(); err != nil {
		log.Error().Msgf("Exit: %s", err)
	}
	general.CleanupWorkspace()
}
