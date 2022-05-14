package dns

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type DnsServer struct {
	dnsAddresses []string
}

func SetupLocalDns(remoteDnsPort, localDnsPort int, dnsOrder []string) error {
	var res = make(chan error)
	go func() {
		upstreamDnsAddresses := getDnsAddresses(dnsOrder, GetNameServer(), remoteDnsPort)
		log.Info().Msgf("Setup local DNS with upstream %v", upstreamDnsAddresses)
		res <-common.SetupDnsServer(&DnsServer{upstreamDnsAddresses}, localDnsPort, "udp")
	}()
	select {
	case err := <-res:
		return err
	case <-time.After(1 * time.Second):
		return nil
	}
}

func getDnsAddresses(dnsOrder []string, upstreamDns string, clusterDnsPort int) []string {
	upstreamPattern := fmt.Sprintf("^([cdptu]{3}:)?%s(:[0-9]+)?$", util.DnsOrderUpstream)
	var dnsAddresses []string
	for _, dnsAddr := range dnsOrder {
		if dnsAddr == util.DnsOrderCluster {
			dnsAddresses = append(dnsAddresses, fmt.Sprintf("tcp:%s:%d", common.Localhost, clusterDnsPort))
		} else if ok, err := regexp.MatchString(upstreamPattern, dnsAddr); err == nil && ok {
			upstreamParts := strings.Split(dnsAddr, ":")
			if upstreamDns != "" {
				switch strings.Count(dnsAddr, ":") {
				case 0:
					dnsAddresses = append(dnsAddresses, fmt.Sprintf("udp:%s:%d", upstreamDns, common.StandardDnsPort))
				case 1:
					if _, err = strconv.Atoi(upstreamParts[1]); err == nil {
						dnsAddresses = append(dnsAddresses, fmt.Sprintf("udp:%s:%s", upstreamDns, upstreamParts[1]))
					} else {
						dnsAddresses = append(dnsAddresses, fmt.Sprintf("%s:%s:%d", upstreamParts[0], upstreamDns, common.StandardDnsPort))
					}
				case 2:
					dnsAddresses = append(dnsAddresses, fmt.Sprintf("%s:%s:%s", upstreamParts[0], upstreamDns, upstreamParts[2]))
				default:
					log.Warn().Msgf("Skip invalid upstream dns server %s", dnsAddr)
				}
			}
		} else {
			switch strings.Count(dnsAddr, ":") {
			case 0:
				dnsAddresses = append(dnsAddresses, fmt.Sprintf("udp:%s:%d", dnsAddr, common.StandardDnsPort))
			case 1:
				if _, err = strconv.Atoi(strings.Split(dnsAddr, ":")[1]); err == nil {
					dnsAddresses = append(dnsAddresses, fmt.Sprintf("udp:%s", dnsAddr))
				} else {
					dnsAddresses = append(dnsAddresses, fmt.Sprintf("%s:%d", dnsAddr, common.StandardDnsPort))
				}
			case 2:
				dnsAddresses = append(dnsAddresses, dnsAddr)
			default:
				log.Warn().Msgf("Skip invalid dns server %s", dnsAddr)
			}
		}
	}
	return dnsAddresses
}

// ServeDNS query DNS record
func (s *DnsServer) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	msg := (&dns.Msg{}).SetReply(req)
	msg.Authoritative = true
	msg.Answer = query(req, s.dnsAddresses)
	if err := w.WriteMsg(msg); err != nil {
		log.Warn().Err(err).Msgf("Failed to reply dns request")
	}
}

func query(req *dns.Msg, dnsAddresses []string) []dns.RR {
	domain := req.Question[0].Name
	qtype := req.Question[0].Qtype

	answer := common.ReadCache(domain, qtype, int64(opt.Get().Connect.DnsCacheTtl))
	if answer != nil {
		log.Debug().Msgf("Found domain %s (%d) in cache", domain, qtype)
		return answer
	}

	for _, dnsAddr := range dnsAddresses {
		dnsParts := strings.SplitN(dnsAddr, ":", 3)
		protocol := dnsParts[0]
		ip := dnsParts[1]
		port, err := strconv.Atoi(dnsParts[2])
		if ip == "" || err != nil || (protocol != "tcp" && protocol != "udp") {
			// skip invalid dns address
			continue
		}
		res, err := common.NsLookup(domain, qtype, protocol, fmt.Sprintf("%s:%d", ip, port))
		if res != nil && len(res.Answer) > 0 {
			// only record none-empty result of cluster dns
			log.Debug().Msgf("Found domain %s (%d) in dns (%s:%d)", domain, qtype, ip, port)
			common.WriteCache(domain, qtype, res.Answer, time.Now().Unix())
			return res.Answer
		} else if err != nil && !common.IsDomainNotExist(err) {
			// usually io timeout error
			log.Warn().Err(err).Msgf("Failed to lookup %s (%d) in dns (%s:%d)", domain, qtype, ip, port)
		}
	}
	log.Debug().Msgf("Empty answer for domain lookup %s (%d)", domain, qtype)
	common.WriteCache(domain, qtype, []dns.RR{}, time.Now().Unix() - int64(opt.Get().Connect.DnsCacheTtl) / 2)
	return []dns.RR{}
}
