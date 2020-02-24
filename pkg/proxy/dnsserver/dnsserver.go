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

	log.Info().Msgf("Successful load local " + resolvFile)
	for _, server := range config.Servers {
		log.Info().Msgf("Success load nameserver %s\n", server)
	}
	for _, domain := range config.Search {
		log.Info().Msgf("Success load search %s\n", domain)
	}
	return
}

//ServeDNS query DNS rescord
func (s *server) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(req)
	msg.Authoritative = true
	// Stuff must be in the answer section
	for _, a := range s.query(req) {
		log.Info().Msgf("%v\n", a)
		msg.Answer = append(msg.Answer, a)
	}

	_ = w.WriteMsg(&msg)
}

func (s *server) getFirst2Parts(origin string) string {
	firstPart := s.getFirstPart(origin)
	return origin[:len(firstPart)] + s.getFirstPart(origin[len(firstPart):])
}

func (s *server) getFirstPart(origin string) string {
	dotIndex := strings.Index(origin, ".") + 1
	return origin[:dotIndex]
}

func (s *server) getDomainWithClusterPostfix(origin string) (domain string) {
	postfix := s.config.Search[0]
	domain = origin + postfix + "."
	log.Info().Msgf("Format domain %s to %s\n", origin, domain)
	return
}

func (s *server) query(req *dns.Msg) (rr []dns.RR) {
	if len(req.Question) <= 0 {
		log.Error().Msgf("*** error: dns Msg question length is 0")
		return
	}

	qtype := req.Question[0].Qtype
	name := req.Question[0].Name

	count := strings.Count(name, ".")
	var err error
	switch count {
	case 0:
		// invalid domain
		log.Warn().Msgf("received invalid domain query: " + name)
		rr = make([]dns.RR, 0)
	case 1:
		// it's service
		rr, err = s.exchange(s.getDomainWithClusterPostfix(name), qtype, name)
		if IsDomainNotExist(err) {
			// it's raw domain
			rr, _ = s.exchange(name, qtype, name)
		}
	case 2:
		// it's raw domain
		rr, err = s.exchange(name, qtype, name)
		if IsDomainNotExist(err) {
			// it's service.namespace
			rr, _ = s.exchange(s.getDomainWithClusterPostfix(name), qtype, name)
		}
	default:
		// it's raw domain
		rr, err = s.exchange(s.getDomainWithClusterPostfix(name), qtype, name)
		if IsDomainNotExist(err) {
			// it's service with custom local domain postfix
			rr, err = s.exchange(s.getDomainWithClusterPostfix(s.getFirstPart(name)), qtype, name)
			if IsDomainNotExist(err) {
				// it's service.namespace with custom local domain postfix
				rr, _ = s.exchange(s.getDomainWithClusterPostfix(s.getFirst2Parts(name)), qtype, name)
			}
		}
	}
	return
}

func (s *server) getResolvServer() (address string, err error) {
	if len(s.config.Servers) <= 0 {
		err = errors.New("*** error: dns server is 0")
		return
	}

	server := s.config.Servers[0]
	port := s.config.Port

	address = net.JoinHostPort(server, port)
	return
}

func (s *server) exchange(domain string, qtype uint16, name string) (rr []dns.RR, err error) {
	log.Info().Msgf("Received DNS query for %s: \n", domain)
	address, err := s.getResolvServer()
	if err != nil {
		log.Error().Msgf(err.Error())
		return
	}
	log.Info().Msgf("Exchange message for domain %s to dns server %s\n", domain, address)

	c := new(dns.Client)
	msg := new(dns.Msg)
	msg.RecursionDesired = true
	msg.SetQuestion(domain, qtype)
	res, _, err := c.Exchange(msg, address)

	if res == nil {
		if err != nil {
			log.Error().Msgf("*** error: %s\n", err.Error())
		} else {
			log.Error().Msgf("*** error: unknown\n")
		}
		return
	}

	if res.Rcode == dns.RcodeNameError {
		err = DomainNotExistError{domain}
		return
	} else if res.Rcode != dns.RcodeSuccess {
		log.Error().Msgf(" *** failed to answer name %s after %d query for %s\n", name, qtype, domain)
		return
	}

	for _, item := range res.Answer {
		log.Info().Msgf("response: %s", item.String())
		r, errInLoop := s.getAnswer(name, domain, item)
		if errInLoop != nil {
			err = errInLoop
			return
		}
		rr = append(rr, r)
	}

	return
}

func (s *server) getAnswer(name string, inClusterName string, acutal dns.RR) (tmp dns.RR, err error) {
	if name != inClusterName {
		log.Info().Msgf("origin %s query name is not same %s", inClusterName, name)
		log.Info().Msgf("origin answer rr to %s", acutal.String())

		var parts []string
		parts = append(parts, name)
		answer := strings.Split(acutal.String(), "\t")
		parts = append(parts, answer[1:]...)

		rrStr := strings.Join(parts, " ")
		log.Info().Msgf("rewrite rr to %s", rrStr)
		tmp, err = dns.NewRR(rrStr)
		if err != nil {
			return
		}
	} else {
		tmp = acutal
	}
	return
}
