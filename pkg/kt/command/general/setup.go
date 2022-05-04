package general

import (
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	k8sRuntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

const UsageTemplate = `Usage:{{if .Runnable}}
  %s{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

// Prepare setup log level, time difference and kube config
func Prepare() error {
	if opt.Get().Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	util.PrepareLogger(opt.Get().Debug)
	k8sRuntime.ErrorHandlers = []func(error){
		func(err error) {
			_, _ = util.BackgroundLogger.Write([]byte(err.Error() + util.Eol))
		},
	}
	klog.SetOutput(util.BackgroundLogger)
	klog.LogToStderr(false)
	if err := combineKubeOpts(); err != nil {
		return err
	}

	log.Info().Msgf("KtConnect %s start at %d (%s %s)",
		opt.Get().RuntimeStore.Version, os.Getpid(), runtime.GOOS, runtime.GOARCH)

	if !opt.Get().SkipTimeDiff {
		if err := cluster.SetupTimeDifference(); err != nil {
			return err
		}
	}
	return nil
}

// SetupProcess write pid file and set component type
func SetupProcess(componentName string) (chan os.Signal, error) {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	opt.Get().RuntimeStore.Component = componentName
	return ch, util.WritePidFile(componentName, ch)
}

// combineKubeOpts set default options of kubectl if not assign
func combineKubeOpts() error {
	if opt.Get().KubeConfig != ""{
		_ = os.Setenv(util.EnvKubeConfig, opt.Get().KubeConfig)
	}
	config, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return fmt.Errorf("failed to parse kubeconfig: %s", err)
	} else if config == nil {
		// should not happen, but issue-275 and issue-285 may cause by it
		return fmt.Errorf("failed to parse kubeconfig")
	}
	if len(opt.Get().KubeContext) > 0 {
		found := false
		for name, _ := range config.Contexts {
			if name == opt.Get().KubeContext {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("context '%s' not exist, check your kubeconfig file please", opt.Get().KubeContext)
		}
		config.CurrentContext = opt.Get().KubeContext
	}
	if len(opt.Get().Namespace) == 0 {
		ctx, exists := config.Contexts[config.CurrentContext]
		if exists && len(ctx.Namespace) > 0 {
			opt.Get().Namespace = config.Contexts[config.CurrentContext].Namespace
		} else {
			opt.Get().Namespace = util.DefaultNamespace
		}
	}
	kubeconfigGetter := func() clientcmd.KubeconfigGetter {
		return func() (*clientcmdapi.Config, error) {
			return config, nil
		}
	}
	restConfig, err := clientcmd.BuildConfigFromKubeconfigGetter("", kubeconfigGetter())
	if err != nil {
		return err
	}
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	opt.Get().RuntimeStore.Clientset = clientSet
	opt.Get().RuntimeStore.RestConfig = restConfig

	clusterName := "none"
	for name, context := range config.Contexts {
		if name == config.CurrentContext {
			clusterName = context.Cluster
			break
		}
	}
	log.Info().Msgf("Using cluster context %s (%s)", config.CurrentContext, clusterName)

	return nil
}
