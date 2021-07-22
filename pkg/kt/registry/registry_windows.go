package registry

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
)

const InternetSettings = "SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Internet Settings"
const notExist = "<NotExist>"

func SetGlobalProxy(port int, config *ProxyConfig) error {
	internetSettings, err := registry.OpenKey(registry.CURRENT_USER, InternetSettings, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer internetSettings.Close()

	val, _, err := internetSettings.GetIntegerValue("ProxyEnable")
	if err == nil {
		config.ProxyEnable = uint32(val)
	} else {
		config.ProxyEnable = 0
	}
	config.ProxyServer, _, err = internetSettings.GetStringValue("ProxyServer")
	if err != nil {
		config.ProxyServer = notExist
	}
	config.ProxyOverride, _, err = internetSettings.GetStringValue("ProxyOverride")
	if err != nil {
		config.ProxyOverride = notExist
	}

	internetSettings.SetDWordValue("ProxyEnable", 1)
	internetSettings.SetStringValue("ProxyServer", fmt.Sprintf("socks=127.0.0.1:%d", port))
	internetSettings.SetStringValue("ProxyOverride", "<local>")

	return nil
}

func CleanGlobalProxy(config *ProxyConfig) {
	internetSettings, err := registry.OpenKey(registry.CURRENT_USER, InternetSettings, registry.ALL_ACCESS)
	if err == nil {
		defer internetSettings.Close()
		internetSettings.SetDWordValue("ProxyEnable", config.ProxyEnable)
		if config.ProxyServer != notExist {
			internetSettings.SetStringValue("ProxyServer", config.ProxyServer)
		} else {
			internetSettings.DeleteValue("ProxyServer")
		}
		if config.ProxyOverride != notExist {
			internetSettings.SetStringValue("ProxyOverride", config.ProxyOverride)
		} else {
			internetSettings.DeleteValue("ProxyOverride")
		}
	}
}
