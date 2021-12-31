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

// Start setup dns server
func Start() {
	srv := NewDNSServerDefault()
	err := srv.ListenAndServe()
	if err != nil {
		log.Error().Err(err).Msgf("Failed to serve")
		panic(err.Error())
	}
}

// NewDNSServerDefault create default dns server
func NewDNSServerDefault() (srv *dns.Server) {
	srv = &dns.Server{Addr: ":" + strconv.Itoa(53), Net: "udp"}
	config, _ := dns.ClientConfigFromFile(resolvFile)

	srv.Handler = &server{config}

	log.Info().Msgf("Successful load local " + resolvFile)
	for _, server := range config.Servers {
		log.Info().Msgf("Success load nameserver %s", server)
	}
	for _, domain := range config.Search {
		log.Info().Msgf("Success load search %s", domain)
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
		log.Info().Msgf("Answer: %v", a)
		msg.Answer = append(msg.Answer, a)
	}

	_ = w.WriteMsg(&msg)
}

// Simulate kubernetes-like dns look up logic
func (s *server) query(req *dns.Msg) (rr []dns.RR) {
	if len(req.Question) <= 0 {
		log.Error().Msgf("No dns Msg question available")
		return
	}

	qtype := req.Question[0].Qtype
	name := req.Question[0].Name
	if !strings.HasSuffix(name, ".") {
		// This should never happen, just in case
		name = name + "."
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
	log.Info().Msgf("Looking up %s", name)

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
		log.Warn().Msgf("Received invalid domain query: " + name)
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
		return "", errors.New("error: no dns server available")
	}

	return net.JoinHostPort(s.config.Servers[0], s.config.Port), nil
}

// Look for domain record from upstream dns server
func (s *server) exchange(domain string, qtype uint16, name string) (rr []dns.RR, err error) {
	address, err := s.getResolveServer()
	if err != nil {
		log.Error().Err(err).Msgf("Failed to fetch upstream dns")
		return
	}
	log.Info().Msgf("Resolving domain %s via upstream %s", domain, address)

	c := new(dns.Client)
	msg := new(dns.Msg)
	msg.RecursionDesired = true
	msg.SetQuestion(domain, qtype)
	res, _, err := c.Exchange(msg, address)

	if res == nil {
		if err != nil {
			log.Error().Err(err).Msgf("Failed to resolve")
		} else {
			log.Error().Msgf("Failed to resolve")
		}
		return
	}

	if res.Rcode == dns.RcodeNameError {
		err = DomainNotExistError{domain}
		return
	} else if res.Rcode != dns.RcodeSuccess {
		log.Error().Msgf("Failed to answer name %s after %d query for %s", name, qtype, domain)
		return
	}

	for _, item := range res.Answer {
		log.Info().Msgf("Response: %s", item.String())
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
