package main

import (
	"fmt"
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
	_ = util.CreateDirIfNotExist(util.KtHome)
	util.FixFileOwner(util.KtHome)
}

func main() {
	// this line must go first
	opt.Get().RuntimeStore.Version = version

	var rootCmd = &cobra.Command{
		Use:   "ktctl",
		Version: version,
		Short: "A utility tool to help you work with Kubernetes dev environment more efficiently",
	}

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		_ = rootCmd.Help()
	}

	action := &command.Action{}
	rootCmd.AddCommand(command.NewConnectCommand(action))
	rootCmd.AddCommand(command.NewExchangeCommand(action))
	rootCmd.AddCommand(command.NewMeshCommand(action))
	rootCmd.AddCommand(command.NewPreviewCommand(action))
	rootCmd.AddCommand(command.NewCleanCommand(action))
	rootCmd.AddCommand(command.NewRecoverCommand(action))
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.SetUsageTemplate(fmt.Sprintf(general.UsageTemplate, "ktctl <command> [command options]"))
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	opt.SetOptions(rootCmd, rootCmd.PersistentFlags(), opt.Get(), opt.GlobalFlags())

	// process will hang here
	if err := rootCmd.Execute(); err != nil {
		log.Error().Msgf("Failed to start: %s", err)
	}
	general.CleanupWorkspace()
}
