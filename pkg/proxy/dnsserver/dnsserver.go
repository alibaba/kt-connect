package dnsserver

import (
	"errors"
	"net"
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
	log.Info().Msgf("looking up %s", name)
	rr, err := s.exchange(name, qtype, name)
	if IsDomainNotExist(err) {
		for _, suffix := range s.getSuffixes() {
			rr, err = s.exchange(name+suffix, qtype, name)
			if err == nil {
				break
			}
		}
	}
	return
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
	log.Info().Msgf("resolve domain %s via server %s", domain, address)

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
