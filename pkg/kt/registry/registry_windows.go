package registry

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
)

const InternetSettings = "SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Internet Settings"

func SetGlobalProxy(port int) error {
	internetSettings, err := registry.OpenKey(registry.CURRENT_USER, InternetSettings, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer internetSettings.Close()

	internetSettings.SetDWordValue("ProxyEnable", 1)
	internetSettings.SetStringValue("ProxyServer", fmt.Sprintf("socks=127.0.0.1:%d", port))

	return nil
}

func CleanGlobalProxy() {
	internetSettings, err := registry.OpenKey(registry.CURRENT_USER, InternetSettings, registry.ALL_ACCESS)
	if err == nil {
		defer internetSettings.Close()
		internetSettings.SetDWordValue("ProxyEnable", 0)
		internetSettings.SetStringValue("ProxyServer", "")
	}
}
