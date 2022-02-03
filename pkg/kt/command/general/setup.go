package general

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
	"runtime"
	"strings"
)

// SetupProcess write pid file and print setup message
func SetupProcess(componentName string, ch chan os.Signal) error {
	opt.Get().RuntimeStore.Component = componentName
	log.Info().Msgf("KtConnect %s start at %d (%s)", opt.Get().RuntimeStore.Version, os.Getpid(), runtime.GOOS)
	return util.WritePidFile(componentName, ch)
}

// validateKubeOpts support like '-n default | --kubeconfig=/path/to/kubeconfig'
func validateKubeOpts(opts []string) error {
	errMsg := "Kubectl option %s invalid, check it by 'kubectl options'"
	for _, opt := range opts {
		// validate like '--kubeconfig=/path/to/kube/config'
		if strings.Contains(opt, "=") && len(strings.Fields(opt)) != 1 {
			return fmt.Errorf(errMsg, opt)
		}
		// validate like '-n default'
		if strings.Contains(opt, " ") && len(strings.Fields(opt)) != 2 {
			return fmt.Errorf(errMsg, opt)
		}
	}
	return nil
}

// CombineKubeOpts set default options of kubectl if not assign
func CombineKubeOpts() error {
	config, err := clientcmd.LoadFromFile(opt.Get().KubeConfig)
	if err != nil {
		return err
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
		if len(config.Contexts[config.CurrentContext].Namespace) > 0 {
			opt.Get().Namespace = config.Contexts[config.CurrentContext].Namespace
		} else {
			opt.Get().Namespace = common.DefaultNamespace
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

	return nil
}
