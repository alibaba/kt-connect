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
		Long: "A utility tool to help you work with Kubernetes dev environment more efficiently",
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

	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false
	rootCmd.PersistentFlags().StringVarP(&opt.Get().Namespace, "namespace", "n", "", "Specify target namespace (otherwise follow kubeconfig current context)")
	rootCmd.PersistentFlags().StringVarP(&opt.Get().KubeConfig, "kubeconfig", "c", "", "Specify path of KubeConfig file")
	rootCmd.PersistentFlags().StringVarP(&opt.Get().Image, "image", "i", fmt.Sprintf("%s:v%s", util.ImageKtShadow, opt.Get().RuntimeStore.Version), "Customize shadow image")
	rootCmd.PersistentFlags().StringVar(&opt.Get().ImagePullSecret, "imagePullSecret", "", "Custom image pull secret")
	rootCmd.PersistentFlags().StringVar(&opt.Get().ServiceAccount, "serviceAccount", "default", "Specify ServiceAccount name for shadow pod")
	rootCmd.PersistentFlags().StringVar(&opt.Get().NodeSelector, "nodeSelector", "", "Specify location of shadow and route pod by node label, e.g. 'disk=ssd,region=hangzhou'")
	rootCmd.PersistentFlags().BoolVarP(&opt.Get().Debug, "debug", "d", false, "Print debug log")
	rootCmd.PersistentFlags().StringVarP(&opt.Get().WithLabels, "withLabel", "l", "", "Extra labels on shadow pod e.g. 'label1=val1,label2=val2'")
	rootCmd.PersistentFlags().StringVar(&opt.Get().WithAnnotations, "withAnnotation", "", "Extra annotation on shadow pod e.g. 'annotation1=val1,annotation2=val2'")
	rootCmd.PersistentFlags().IntVar(&opt.Get().PortForwardWaitTime, "portForwardTimeout", 10, "Seconds to wait before port-forward connection timeout")
	rootCmd.PersistentFlags().IntVar(&opt.Get().PodCreationWaitTime, "podCreationTimeout", 60, "Seconds to wait before shadow or router pod creation timeout")
	rootCmd.PersistentFlags().BoolVar(&opt.Get().UseShadowDeployment, "useShadowDeployment", false, "Deploy shadow container as deployment")
	rootCmd.PersistentFlags().BoolVar(&opt.Get().SkipTimeDiff, "useLocalTime", false, "Use local time for resource heartbeat timestamp")
	rootCmd.PersistentFlags().BoolVarP(&opt.Get().AlwaysUpdateShadow, "forceUpdate", "f", false, "Always update shadow image")
	rootCmd.PersistentFlags().StringVar(&opt.Get().KubeContext, "context", "", "Specify current context of kubeconfig")
	rootCmd.PersistentFlags().StringVar(&opt.Get().PodQuota, "podQuota", "", "Specify resource limit for shadow and router pod, e.g. '0.5c,512m'")

	rootCmd.PersistentFlags().BoolVar(&opt.Get().RunAsWorkerProcess, "asWorker", false, "Run as worker process")
	_ = rootCmd.PersistentFlags().MarkHidden("asWorker")
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	// process will hang here
	if err := rootCmd.Execute(); err != nil {
		log.Error().Msgf("Failed to start: %s", err)
	}
	general.CleanupWorkspace()
}
