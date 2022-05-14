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
	// must merge config file before any other code which may use the option object
	mergeConfigFile()

	// then setup logs
	SetupLogger()

	if err := combineKubeOpts(); err != nil {
		return err
	}

	log.Info().Msgf("KtConnect %s start at %d (%s %s)",
		opt.Store.Version, os.Getpid(), runtime.GOOS, runtime.GOARCH)

	if !opt.Get().Global.UseLocalTime {
		if err := cluster.SetupTimeDifference(); err != nil {
			return err
		}
	}
	return nil
}

func SetupLogger() {
	if opt.Get().Global.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	util.PrepareLogger(opt.Get().Global.Debug)
	k8sRuntime.ErrorHandlers = []func(error){
		func(err error) {
			_, _ = util.BackgroundLogger.Write([]byte(err.Error() + util.Eol))
		},
	}
	klog.SetOutput(util.BackgroundLogger)
	klog.LogToStderr(false)
}

// SetupProcess write pid file and set component type
func SetupProcess(componentName string) (chan os.Signal, error) {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	opt.Store.Component = componentName
	return ch, util.WritePidFile(componentName, ch)
}

// mergeConfigFile merge config file and command line options
func mergeConfigFile() {
	// this method is invoke before logger setup, don't do any logging here
	// TODO: read config file
}

// combineKubeOpts set default options of kubectl if not assign
func combineKubeOpts() error {
	if opt.Get().Global.Kubeconfig != ""{
		_ = os.Setenv(util.EnvKubeConfig, opt.Get().Global.Kubeconfig)
	}
	config, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if len(customizeKubeConfig) > 50 {
		config, err = clientcmd.Load([]byte(customizeKubeConfig))
	}
	if err != nil {
		return fmt.Errorf("failed to parse kubeconfig: %s", err)
	} else if config == nil {
		// should not happen, but issue-275 and issue-285 may cause by it
		return fmt.Errorf("failed to parse kubeconfig")
	}
	if len(opt.Get().Global.Context) > 0 {
		found := false
		for name, _ := range config.Contexts {
			if name == opt.Get().Global.Context {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("context '%s' not exist, check your kubeconfig file please", opt.Get().Global.Context)
		}
		config.CurrentContext = opt.Get().Global.Context
	}
	if len(opt.Get().Global.Namespace) == 0 {
		ctx, exists := config.Contexts[config.CurrentContext]
		if exists && len(ctx.Namespace) > 0 {
			opt.Get().Global.Namespace = config.Contexts[config.CurrentContext].Namespace
		} else {
			opt.Get().Global.Namespace = util.DefaultNamespace
		}
	}
	kubeConfigGetter := func() clientcmd.KubeconfigGetter {
		return func() (*clientcmdapi.Config, error) {
			return config, nil
		}
	}
	restConfig, err := clientcmd.BuildConfigFromKubeconfigGetter("", kubeConfigGetter())
	if err != nil {
		return err
	}
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	opt.Store.Clientset = clientSet
	opt.Store.RestConfig = restConfig

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
