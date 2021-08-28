package registry

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows/registry"
	"strings"
	"syscall"
	"unsafe"
)

const (
	HWND_BROADCAST      = uintptr(0xffff)
	WM_SETTINGCHANGE    = uintptr(0x001A)
	InternetSettings    = "Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings"
	EnvironmentSettings = "Environment"
	notExist            = "<NotExist>"
	RegKeyProxyEnable   = "ProxyEnable"
	RegKeyProxyServer   = "ProxyServer"
	RegKeyProxyOverride = "ProxyOverride"
	RegKeyHttpProxy     = "HTTP_PROXY"
	User32Dll           = "user32.dll"
	ApiSendMessage      = "SendMessageW"
	IntSocksLocalhost   = "socks=127.0.0.1:"
	EnvSocksLocalhost   = "socks://127.0.0.1:"
)

func SetGlobalProxy(port int, config *ProxyConfig) error {
	internetSettings, err := registry.OpenKey(registry.CURRENT_USER, InternetSettings, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer internetSettings.Close()

	val, _, err := internetSettings.GetIntegerValue(RegKeyProxyEnable)
	if err == nil {
		config.ProxyEnable = uint32(val)
	} else {
		config.ProxyEnable = 0
	}
	config.ProxyServer, _, err = internetSettings.GetStringValue(RegKeyProxyServer)
	if err != nil {
		config.ProxyServer = notExist
	}
	config.ProxyOverride, _, err = internetSettings.GetStringValue(RegKeyProxyOverride)
	if err != nil {
		config.ProxyOverride = notExist
	}

	err = internetSettings.SetDWordValue(RegKeyProxyEnable, 1)
	if err != nil {
		return err
	}
	internetSettings.SetStringValue(RegKeyProxyServer, fmt.Sprintf("%s%d", IntSocksLocalhost, port))
	internetSettings.SetStringValue(RegKeyProxyOverride, "<local>")
	return nil
}

func CleanGlobalProxy(config *ProxyConfig) {
	internetSettings, err := registry.OpenKey(registry.CURRENT_USER, InternetSettings, registry.ALL_ACCESS)
	if err == nil {
		defer internetSettings.Close()
		internetSettings.SetDWordValue(RegKeyProxyEnable, config.ProxyEnable)
		if config.ProxyServer != notExist {
			internetSettings.SetStringValue(RegKeyProxyServer, config.ProxyServer)
		} else {
			internetSettings.DeleteValue(RegKeyProxyServer)
		}
		if config.ProxyOverride != notExist {
			internetSettings.SetStringValue(RegKeyProxyOverride, config.ProxyOverride)
		} else {
			internetSettings.DeleteValue(RegKeyProxyOverride)
		}
	} else {
		log.Error().Msgf("Failed to reset global proxy configuration: %s", err)
	}
}

func SetHttpProxyEnvironmentVariable(port int, config *ProxyConfig) error {
	internetSettings, err := registry.OpenKey(registry.CURRENT_USER, EnvironmentSettings, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer internetSettings.Close()

	config.HttpProxyVar, _, err = internetSettings.GetStringValue(RegKeyHttpProxy)
	if err != nil {
		config.HttpProxyVar = notExist
	}

	internetSettings.SetStringValue(RegKeyHttpProxy, fmt.Sprintf("%s%d", EnvSocksLocalhost, port))
	refreshEnvironmentVariable()
	return nil
}

func CleanHttpProxyEnvironmentVariable(config *ProxyConfig) {
	internetSettings, err := registry.OpenKey(registry.CURRENT_USER, EnvironmentSettings, registry.ALL_ACCESS)
	if err == nil {
		defer internetSettings.Close()
		if config.HttpProxyVar != notExist {
			internetSettings.SetStringValue(RegKeyHttpProxy, config.HttpProxyVar)
		} else {
			internetSettings.DeleteValue(RegKeyHttpProxy)
		}
		refreshEnvironmentVariable()
	} else {
		log.Error().Msgf("Failed to reset global proxy environment variable: %s", err)
	}
}

func ResetGlobalProxyAndEnvironmentVariable() {
	resetGlobalProxy()
	resetHttpProxyEnvironmentVariable()
}

func resetGlobalProxy() {
	internetSettings, err := registry.OpenKey(registry.CURRENT_USER, InternetSettings, registry.ALL_ACCESS)
	if err == nil {
		defer internetSettings.Close()
		val, _, err := internetSettings.GetStringValue(RegKeyProxyServer)
		if err == nil && strings.HasPrefix(val, IntSocksLocalhost) {
			internetSettings.SetDWordValue(RegKeyProxyEnable, 0)
			internetSettings.DeleteValue(RegKeyProxyServer)
			internetSettings.DeleteValue(RegKeyProxyOverride)
		}
	}
}

func resetHttpProxyEnvironmentVariable() {
	internetSettings, err := registry.OpenKey(registry.CURRENT_USER, EnvironmentSettings, registry.ALL_ACCESS)
	if err == nil {
		defer internetSettings.Close()
		val, _, err := internetSettings.GetStringValue(RegKeyHttpProxy)
		if err == nil && strings.HasPrefix(val, EnvSocksLocalhost) {
			internetSettings.DeleteValue(RegKeyHttpProxy)
		}
		refreshEnvironmentVariable()
	}
}

func refreshEnvironmentVariable() {
	syscall.NewLazyDLL(User32Dll).NewProc(ApiSendMessage).Call(
		HWND_BROADCAST, WM_SETTINGCHANGE, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Environment"))))
}
