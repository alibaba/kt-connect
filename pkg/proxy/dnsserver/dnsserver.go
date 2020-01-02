package dnsserver

import (
	"errors"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

// dns server
type server struct{}

// NewDNSServerDefault create default dns server
func NewDNSServerDefault() (srv *dns.Server) {
	srv = &dns.Server{Addr: ":" + strconv.Itoa(53), Net: "udp"}
	srv.Handler = &server{}

	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")

	log.Info().Msgf("Successful load local /etc/resolv.conf")
	for _, server := range config.Servers {
		log.Info().Msgf("Success load nameserver %s\n", server)
	}
	return
}

//ServeDNS query DNS rescord
func (h *server) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(req)
	msg.Authoritative = true
	// Stuff must be in the answer section
	for _, a := range query(w, req) {
		log.Info().Msgf("%v\n", a)
		msg.Answer = append(msg.Answer, a)
	}

	w.WriteMsg(&msg)
}

func getDomain(origin string) string {
	domain := origin

	namespace, find := os.LookupEnv("PROXY_NAMESPACE")
	if !find {
		namespace = "default"
	}

	if !strings.Contains(domain, ".") {
		domain = domain + namespace + ".svc.cluster.local."
		log.Info().Msgf("*** Use in cluster dns address %s\n", domain)
	}

	log.Info().Msgf("Format domain %s to %s\n", origin, domain)
	return domain
}

func query(w dns.ResponseWriter, req *dns.Msg) (rr []dns.RR) {
	if len(req.Question) <= 0 {
		log.Error().Msgf("*** error: dns Msg question length is 0")
		return
	}

	msg := new(dns.Msg)
	msg.RecursionDesired = true

	qtype := req.Question[0].Qtype
	name := req.Question[0].Name
	domain := getDomain(name)

	log.Info().Msgf("Received DNS query for %s: \n", domain)
	msg.SetQuestion(domain, qtype)
	rr = exchange(domain, qtype, msg, name)
	return
}

func getResolvServer() (address string, err error) {
	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")

	if len(config.Servers) <= 0 {
		err = errors.New("*** error: dns server is 0")
		return
	}

	server := config.Servers[0]
	port := config.Port

	address = net.JoinHostPort(server, port)
	return
}

func exchange(domain string, Qtype uint16, m *dns.Msg, name string) (rr []dns.RR) {
	address, err := getResolvServer()
	if err != nil {
		log.Error().Msgf(err.Error())
		return
	}
	log.Info().Msgf("Exchange message for domain %s to dns server %s\n", domain, address)

	c := new(dns.Client)
	res, _, err := c.Exchange(m, address)

	if res == nil {
		if err != nil {
			log.Error().Msgf("*** error: %s\n", err.Error())
		} else {
			log.Error().Msgf("*** error: unknown\n")
		}
		return
	}

	if res.Rcode != dns.RcodeSuccess {
		log.Error().Msgf(" *** invalid answer name %s after %d query for %s\n", name, Qtype, domain)
		return
	}

	for _, item := range res.Answer {
		log.Info().Msgf("response: %s", item.String())
		r, err := getAnswer(name, domain, item)
		if err != nil {
			return
		}
		rr = append(rr, r)
	}

	return
}

func getAnswer(name string, inClusterName string, acutal dns.RR) (tmp dns.RR, err error) {
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
