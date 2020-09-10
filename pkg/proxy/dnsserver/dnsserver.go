package dnsserver

import (
	"errors"
	"github.com/alibaba/kt-connect/pkg/common"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

// dns server
type server struct {
	config *dns.ClientConfig
}

// constants
const resolvFile = "/etc/resolv.conf"

// NewDNSServerDefault create default dns server
func NewDNSServerDefault() (srv *dns.Server) {
	srv = &dns.Server{Addr: ":" + strconv.Itoa(53), Net: "udp"}
	config, _ := dns.ClientConfigFromFile(resolvFile)

	srv.Handler = &server{config}

	log.Info().Msgf("successful load local " + resolvFile)
	for _, server := range config.Servers {
		log.Info().Msgf("success load nameserver %s", server)
	}
	for _, domain := range config.Search {
		log.Info().Msgf("success load search %s", domain)
	}
	return
}

// ServeDNS query DNS record
func (s *server) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(req)
	msg.Authoritative = true
	// Stuff must be in the answer section
	for _, a := range s.query(req) {
		log.Info().Msgf("answer: %v", a)
		msg.Answer = append(msg.Answer, a)
	}

	_ = w.WriteMsg(&msg)
}

// Simulate kubernetes-like dns look up logic
func (s *server) query(req *dns.Msg) (rr []dns.RR) {
	if len(req.Question) <= 0 {
		log.Error().Msgf("error: no dns Msg question available")
		return
	}

	qtype := req.Question[0].Qtype
	name := req.Question[0].Name
	if !strings.HasSuffix(name, ".") {
		// This should never happen, just in case
		name = name + "."
	}
	localDomain := os.Getenv(common.EnvVarLocalDomain)
	if localDomain != "" && strings.HasSuffix(name, localDomain+".") {
		name = name[0:(len(name) - len(localDomain) - 1)]
	}
	log.Info().Msgf("looking up %s", name)

	rr = make([]dns.RR, 0)
	domainsToLookup := s.fetchAllPossibleDomains(name)
	for _, domain := range domainsToLookup {
		r, err := s.exchange(domain, qtype, name)
		if err == nil {
			rr = r
			break
		}
	}
	return
}

// get all domains need to lookup
func (s *server) fetchAllPossibleDomains(name string) []string {
	count := strings.Count(name, ".")
	domainSuffixes := s.getSuffixes()
	var namesToLookup []string
	switch count {
	case 0:
		// invalid domain, dns name always ends with a '.'
		log.Warn().Msgf("received invalid domain query: " + name)
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
	default:
		// raw domain
		namesToLookup = append(namesToLookup, name)
	}
	return namesToLookup
}

// Convert short domain to fully qualified domain name
func (s *server) getSuffixes() (suffixes []string) {
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
func (s *server) getResolveServer() (address string, err error) {
	if len(s.config.Servers) <= 0 {
		err = errors.New("error: no dns server available")
		return
	}

	server := s.config.Servers[0]
	port := s.config.Port

	address = net.JoinHostPort(server, port)
	return
}

// Look for domain record from upstream dns server
func (s *server) exchange(domain string, qtype uint16, name string) (rr []dns.RR, err error) {
	address, err := s.getResolveServer()
	if err != nil {
		log.Error().Msgf("error: fail to fetch upstream dns: %s", err.Error())
		return
	}
	log.Info().Msgf("resolving domain %s via upstream %s", domain, address)

	c := new(dns.Client)
	msg := new(dns.Msg)
	msg.RecursionDesired = true
	msg.SetQuestion(domain, qtype)
	res, _, err := c.Exchange(msg, address)

	if res == nil {
		if err != nil {
			log.Error().Msgf("error: fail to resolve: %s", err.Error())
		} else {
			log.Error().Msgf("error: fail to resolve")
		}
		return
	}

	if res.Rcode == dns.RcodeNameError {
		err = DomainNotExistError{domain}
		return
	} else if res.Rcode != dns.RcodeSuccess {
		log.Error().Msgf("error: failed to answer name %s after %d query for %s", name, qtype, domain)
		return
	}

	for _, item := range res.Answer {
		log.Info().Msgf("response: %s", item.String())
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
func (s *server) convertAnswer(name, inClusterName string, actual dns.RR) (rr dns.RR, err error) {
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
