package options

import (
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"reflect"
	"strconv"
)

// ConnectOptions ...
type ConnectOptions struct {
	Global           bool
	DisablePodIp     bool
	DisableTunDevice bool
	DisableTunRoute  bool
	ProxyPort        int
	DnsPort          int
	DnsCacheTtl      int
	IncludeIps       string
	ExcludeIps       string
	Mode             string
	DnsMode          string
	ShareShadow      bool
	ClusterDomain    string
	SkipCleanup      bool
}

// ExchangeOptions ...
type ExchangeOptions struct {
	Mode             string
	Expose           string
	RecoverWaitTime  int
	SkipPortChecking bool
}

// MeshOptions ...
type MeshOptions struct {
	Mode             string
	Expose           string
	VersionMark      string
	RouterImage      string
	SkipPortChecking bool
}

// RecoverOptions ...
type RecoverOptions struct {
}

// PreviewOptions ...
type PreviewOptions struct {
	External bool
	Expose   string
	SkipPortChecking bool
}

// CleanOptions ...
type CleanOptions struct {
	DryRun           bool
	ThresholdInMinus int64
	LocalOnly        bool
}

// ConfigOptions ...
type ConfigOptions struct {
}

// GlobalOptions ...
type GlobalOptions struct {
	AsWorker            bool
	Kubeconfig          string
	Namespace           string
	ServiceAccount      string
	Debug               bool
	Image               string
	ImagePullSecret     string
	NodeSelector        string
	WithLabel           string
	WithAnnotation      string
	PortForwardTimeout  int
	PodCreationTimeout  int
	UseShadowDeployment bool
	ForceUpdate         bool
	UseLocalTime        bool
	Context             string
	PodQuota            string
	ListenCheck         bool
}

// DaemonOptions cli options
type DaemonOptions struct {
	Connect  *ConnectOptions
	Exchange *ExchangeOptions
	Mesh     *MeshOptions
	Preview  *PreviewOptions
	Recover  *RecoverOptions
	Clean    *CleanOptions
	Config   *ConfigOptions
	Global   *GlobalOptions
}

var opt *DaemonOptions

// Get fetch options instance
func Get() *DaemonOptions {
	if opt == nil {
		opt = &DaemonOptions{
			Global:   &GlobalOptions{},
			Connect:  &ConnectOptions{},
			Exchange: &ExchangeOptions{},
			Mesh:     &MeshOptions{},
			Preview:  &PreviewOptions{},
			Recover:  &RecoverOptions{},
			Clean:    &CleanOptions{},
			Config:   &ConfigOptions{},
		}
		if customize, exist := GetCustomizeKtConfig(); exist {
			mergeOptions(opt, []byte(customize))
		}
		if configData, err := ioutil.ReadFile(util.KtConfigFile); err == nil {
			mergeOptions(opt, configData)
		}
	}
	return opt
}

func mergeOptions(opt *DaemonOptions, data []byte) {
	config := make(map[string]map[string]string)
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		log.Warn().Msgf("Invalid config content, skipping ...")
		return
	}
	for group, item := range config {
		for key, value := range item {
			groupField := reflect.ValueOf(opt).Elem().FieldByName(util.Capitalize(group))
			if groupField.IsValid() {
				itemField := groupField.Elem().FieldByName(util.Capitalize(key))
				if itemField.IsValid() {
					switch itemField.Kind() {
					case reflect.String:
						itemField.SetString(value)
					case reflect.Int:
						if v, err2 := strconv.Atoi(value); err2 == nil {
							itemField.SetInt(int64(v))
						} else {
							log.Warn().Msgf("Config item '%s.%s' value is not integer: %s", group, key, value)
						}
					case reflect.Bool:
						if v, err2 := strconv.ParseBool(value); err2 == nil {
							itemField.SetBool(v)
						} else {
							log.Warn().Msgf("Config item '%s.%s' value is not bool: %s", group, key, value)
						}
					default:
						log.Warn().Msgf("Config item '%s.%s' of invalid type: %s",
							group, key, itemField.Kind().String())
					}
					log.Debug().Msgf("Loaded %s.%s = %s", group, key, value)
				}
			}
		}
	}
}
