package dns

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	"time"
)

type DnsServer struct {
	upstreamAddr string
}

func SetupLocalDns(upstreamIp string, dnsPort int) error {
	var success = make(chan error)
	go func() {
		time.Sleep(1 * time.Second)
		success <-nil
	}()
	go func() {
		success <-common.SetupDnsServer(&DnsServer{
			upstreamAddr: fmt.Sprintf("%s:%d", upstreamIp, common.RemoteDnsPort),
		}, dnsPort, "udp")
	}()
	return <-success
}

// ServeDNS query DNS record
func (s *DnsServer) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	msg := (&dns.Msg{}).SetReply(req)
	msg.Authoritative = true
	res, err := common.NsLookup(req.Question[0].Name, req.Question[0].Qtype, "tcp", s.upstreamAddr)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to lookup  %s", req.Question[0].Name)
		return
	}
	msg.Answer = res.Answer
	if err = w.WriteMsg(msg); err != nil {
		log.Warn().Err(err).Msgf("Failed to reply dns request")
	}
}
