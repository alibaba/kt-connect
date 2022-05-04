package dnsserver

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	"net"
	"strings"
)

// DnsServer nds server
type DnsServer struct {
	localDomain string
	config *dns.ClientConfig
}

// Start setup dns server
func Start(dnsPort int, dnsProtocol string, localDomain string) {
	config, _ := dns.ClientConfigFromFile(util.ResolvConf)
	for _, server := range config.Servers {
		log.Info().Msgf("Load nameserver %s", server)
	}
	for _, domain := range config.Search {
		log.Info().Msgf("Load search %s", domain)
	}
	err := common.SetupDnsServer(&DnsServer{localDomain, config}, dnsPort, dnsProtocol)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to start dns server")
	}
}

// ServeDNS query DNS record
func (s *DnsServer) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	msg := (&dns.Msg{}).SetReply(req)
	msg.Authoritative = true
	msg.Answer = s.query(req)
	log.Info().Msgf("Answer: %v", msg.Answer)

	if err := w.WriteMsg(msg); err != nil {
		log.Error().Err(err).Msgf("Failed to response")
	}
}

// Simulate kubernetes-like dns look up logic
func (s *DnsServer) query(req *dns.Msg) (rr []dns.RR) {
	if len(req.Question) <= 0 {
		log.Error().Msgf("No dns Msg question available")
		return
	}

	name := req.Question[0].Name
	qtype := req.Question[0].Qtype
	answer := common.ReadCache(name, qtype, 60)
	if answer != nil {
		log.Debug().Msgf("Found domain %s (%d) in cache", name, qtype)
		return answer
	}

	if s.localDomain != "" {
		for _, d := range strings.Split(s.localDomain, ",") {
			if strings.HasSuffix(name, d+".") {
				name = name[0:(len(name) - len(d) - 1)]
				break
			}
		}
	}
	log.Info().Msgf("Looking up %s (%d)", name, qtype)

	rr = make([]dns.RR, 0)
	domainsToLookup := s.fetchAllPossibleDomains(name)
	for _, domain := range domainsToLookup {
		r, err := s.lookup(domain, qtype, name)
		if err == nil {
			rr = r
			break
		}
	}
	common.WriteCache(name, qtype, rr)
	return
}

// get all domains need to lookup
func (s *DnsServer) fetchAllPossibleDomains(name string) []string {
	count := strings.Count(name, ".")
	domainSuffixes := s.getSuffixes()
	var namesToLookup []string
	switch count {
	case 0:
		// invalid domain, dns name always ends with a '.'
		log.Warn().Msgf("Received invalid domain query %s", name)
	case 1:
		if len(domainSuffixes) > 0 {
			// service name
			namesToLookup = append(namesToLookup, name+domainSuffixes[0])
		}
		// raw domain
		namesToLookup = append(namesToLookup, name)
	case 2:
		if len(domainSuffixes) > 1 {
			// service.namespace name
			namesToLookup = append(namesToLookup, name+domainSuffixes[1])
			// stateful-set-pod.service name
			namesToLookup = append(namesToLookup, name+domainSuffixes[0])
		}
		// raw domain
		namesToLookup = append(namesToLookup, name)
	case 3:
		// raw domain
		namesToLookup = append(namesToLookup, name)
		if len(domainSuffixes) > 1 {
			// stateful-set-pod.service.namespace name
			namesToLookup = append(namesToLookup, name+domainSuffixes[1])
		}
		if len(domainSuffixes) > 2 {
			// service.namespace.svc name
			namesToLookup = append(namesToLookup, name+domainSuffixes[2])
		}
	case 4:
		// raw domain
		namesToLookup = append(namesToLookup, name)
		if len(domainSuffixes) > 2 {
			// stateful-set-pod.service.namespace.svc name
			namesToLookup = append(namesToLookup, name+domainSuffixes[2])
		}
	default:
		// raw domain
		namesToLookup = append(namesToLookup, name)
	}
	return namesToLookup
}

// Convert short domain to fully qualified domain name
func (s *DnsServer) getSuffixes() (suffixes []string) {
	for _, s := range s.config.Search {
		// @see https://github.com/alibaba/kt-connect/issues/153
		if strings.HasSuffix(s, ".") {
			suffixes = append(suffixes, s)
		} else {
			suffixes = append(suffixes, s+".")
		}
	}
	return
}

// Get upstream dns server address
func (s *DnsServer) getResolveServer() (address string, err error) {
	if len(s.config.Servers) <= 0 {
		return "", fmt.Errorf("error: no dns server available")
	}

	return net.JoinHostPort(s.config.Servers[0], s.config.Port), nil
}

// Look for domain record from upstream dns server
func (s *DnsServer) lookup(domain string, qtype uint16, name string) ([]dns.RR, error) {
	address, err := s.getResolveServer()
	if err != nil {
		log.Error().Err(err).Msgf("Failed to fetch upstream dns")
		return []dns.RR{}, err
	}
	log.Debug().Msgf("Resolving domain %s (%d) via upstream %s", domain, qtype, address)

	res, err := common.NsLookup(domain, qtype, "udp", address)
	if err != nil {
		if common.IsDomainNotExist(err) {
			log.Debug().Msgf(err.Error())
		} else {
			log.Warn().Err(err).Msgf("Failed to answer name %s (%d) query for %s", name, qtype, domain)
		}
		return []dns.RR{}, err
	}

	if len(res.Answer) == 0 {
		log.Debug().Msgf("Empty answer")
	}
	return s.convertAnswer(name, res.Answer), nil
}

// Replace fully qualified domain name with short domain name in dns answer
func (s *DnsServer) convertAnswer(name string, answer []dns.RR) []dns.RR {
	cnames := []string{name}
	for _, item := range answer {
		log.Debug().Msgf("Response: %s", item.String())
		if item.Header().Rrtype == dns.TypeCNAME {
			cnames = append(cnames, item.(*dns.CNAME).Target)
		}
	}
	for _, item := range answer {
		if !util.Contains(item.Header().Name, cnames) {
			item.Header().Name = name
		}
	}
	return answer
}
