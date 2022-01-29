package dns

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type DnsServer struct {
	upstreamIp string
}

func SetupLocalDns(shadowPodIp string) {
	common.SetupDnsServer(&DnsServer{ upstreamIp: shadowPodIp }, 53, "udp")
}

// ServeDNS query DNS record
func (s *DnsServer) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	msg := (&dns.Msg{}).SetReply(req)
	msg.Authoritative = true
	res, err := common.NsLookup(req.Question[0].Name, req.Question[0].Qtype, "tcp", s.upstreamIp)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to lookup  %s", req.Question[0].Name)
		return
	}
	msg.Answer = res.Answer
	if err = w.WriteMsg(msg); err != nil {
		log.Warn().Err(err).Msgf("Failed to reply dns request")
	}
}
