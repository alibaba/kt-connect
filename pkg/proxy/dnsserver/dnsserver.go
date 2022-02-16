package dnsserver

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"strings"
)

// DnsServer nds server
type DnsServer struct {
	config *dns.ClientConfig
}

// Start setup dns server
func Start() {
	config, _ := dns.ClientConfigFromFile(util.ResolvConf)
	for _, server := range config.Servers {
		log.Info().Msgf("Load nameserver %s", server)
	}
	for _, domain := range config.Search {
		log.Info().Msgf("Load search %s", domain)
	}
	err := common.SetupDnsServer(&DnsServer{config}, common.StandardDnsPort, os.Getenv(common.EnvVarDnsProtocol))
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
	answer := common.ReadCache(name, qtype)
	if answer != nil {
		log.Debug().Msgf("Found domain %s (%d) in cache", name, qtype)
		return answer
	}

	localDomains := os.Getenv(common.EnvVarLocalDomains)
	if localDomains != "" {
		for _, d := range strings.Split(localDomains, ",") {
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
			// stateful-set-pod.service name
			namesToLookup = append(namesToLookup, name+domainSuffixes[0])
			// service.namespace name
			namesToLookup = append(namesToLookup, name+domainSuffixes[1])
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
func (s *DnsServer) lookup(domain string, qtype uint16, name string) (rr []dns.RR, err error) {
	address, err := s.getResolveServer()
	if err != nil {
		log.Error().Err(err).Msgf("Failed to fetch upstream dns")
		return
	}
	log.Debug().Msgf("Resolving domain %s (%d) via upstream %s", domain, qtype, address)

	res, err := common.NsLookup(domain, qtype, "udp", address)
	if err != nil {
		if common.IsDomainNotExist(err) {
			log.Debug().Msgf(err.Error())
		} else {
			log.Warn().Err(err).Msgf("Failed to answer name %s (%d) query for %s", name, qtype, domain)
		}
		return
	}

	if len(res.Answer) == 0 {
		log.Debug().Msgf("Empty answer")
	}
	for _, item := range res.Answer {
		log.Debug().Msgf("Response: %s", item.String())
		r, errInLoop := s.convertAnswer(name, domain, item)
		if errInLoop != nil {
			err = errInLoop
			return
		}
		rr = append(rr, r)
	}

	return
}

// Replace fully qualified domain name with short domain name in dns answer
func (s *DnsServer) convertAnswer(name, inClusterName string, actual dns.RR) (rr dns.RR, err error) {
	if name != inClusterName {
		var parts []string
		parts = append(parts, name)
		answer := strings.Split(actual.String(), "\t")
		parts = append(parts, answer[1:]...)
		rrStr := strings.Join(parts, " ")
		rr, err = dns.NewRR(rrStr)
		if err != nil {
			return
		}
	} else {
		rr = actual
	}
	rr.Header().Name = name
	return
}
