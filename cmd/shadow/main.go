package main

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/shadow/dnsserver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

const (
	// ArgLocalDomains application argument for local domain config
	ArgLocalDomains = "--local-domain"
	// ArgDnsProtocol application argument for shadow pod dns protocol
	ArgDnsProtocol = "--protocol"
	// ArgLogLevel application argument for shadow pod log level
	ArgLogLevel = "--log-level"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	logLevel := getParameter(common.EnvVarLogLevel, ArgLogLevel, "info")
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to parse log level")
	}
	zerolog.SetGlobalLevel(level)
	dnsPort := common.StandardDnsPort
	dnsProtocol := getParameter(common.EnvVarDnsProtocol, ArgDnsProtocol, "udp")
	localDomain := getParameter(common.EnvVarLocalDomains, ArgLocalDomains, "")
	log.Info().Msgf("Shadow DNS on %s port %d, log level %s", dnsProtocol, dnsPort, logLevel)
	if localDomain != "" {
		log.Info().Msgf("Using local domain %s", localDomain)
	}
	dnsserver.Start(dnsPort, dnsProtocol, localDomain)
}

func getParameter(envVar string, argVar string, defaultValue string) string {
	if os.Getenv(envVar) != "" {
		return os.Getenv(envVar)
	}
	for _, arg := range os.Args {
		kv := strings.SplitN(arg, "=", 2)
		if len(kv) > 1 && kv[0] == argVar && kv[1] != "" {
			return kv[1]
		}
	}
	return defaultValue
}
