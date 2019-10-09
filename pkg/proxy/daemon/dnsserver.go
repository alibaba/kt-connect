package daemon

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/miekg/dns"
)

type dnsHandler struct{}

// StartDNSDaemon start dns server
func StartDNSDaemon() (err error) {
	srv := &dns.Server{Addr: ":" + strconv.Itoa(53), Net: "udp"}
	srv.Handler = &dnsHandler{}

	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")

	log.Info().Msgf("Successful load local /etc/resolv.conf")
	for _, server := range config.Servers {
		log.Info().Msgf("Success load nameserver %s\n", server)
	}

	fmt.Printf("DNS Server Start At 53...\n")
	err = srv.ListenAndServe()
	if err != nil {
		log.Error().Msgf("Failed to set udp listener %s\n", err.Error())
	}
	return
}

//ServeDNS query DNS rescord
func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	msg.Authoritative = true
	// Stuff must be in the answer section
	for _, a := range query(w, r) {
		log.Info().Msgf("%v\n", a)
		msg.Answer = append(msg.Answer, a)
	}

	w.WriteMsg(&msg)
}

// getDomain get internal service dns address
func getDomain(origin string) string {
	domain := origin

	namespace, find := os.LookupEnv("PROXY_NAMESPACE")
	if !find {
		namespace = "default"
	}

	index := strings.Index(domain, ".")

	if index+1 == len(domain) {
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

	domain := getDomain(req.Question[0].Name)
	qtype := req.Question[0].Qtype
	log.Info().Msgf("Received DNS query for %s: \n", domain)

	msg := new(dns.Msg)
	msg.SetQuestion(domain, qtype)
	msg.RecursionDesired = true

	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")

	if len(config.Servers) <= 0 {
		log.Error().Msgf("*** error: dns server is 0")
		return
	}

	rr = exchange(domain, config.Servers[0], config.Port, qtype, msg)
	return
}

func exchange(domain string, server string, port string, Qtype uint16, m *dns.Msg) (rr []dns.RR) {
	log.Info().Msgf("Exchange message for domain %s to dns server %s:%s\n", domain, server, port)

	c := new(dns.Client)
	res, _, err := c.Exchange(m, net.JoinHostPort(server, port))

	if res == nil {
		log.Error().Msgf("*** error: %s\n", err.Error())
		return
	}

	if res.Rcode != dns.RcodeSuccess {
		log.Error().Msgf(" *** invalid answer name %s after %d query for %s\n", domain, Qtype, domain)
		return
	}

	rr = res.Answer
	return
}
