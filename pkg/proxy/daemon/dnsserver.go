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

type dnsHandler struct{}

// getDomain get internal service dns address
func (h *dnsHandler) getDomain(origin string) string {
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

//ServeDNS query DNS rescord
func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {

	msg := dns.Msg{}
	msg.SetReply(r)
	msg.Authoritative = true

	origin := r.Question[0].Name

	domain := h.getDomain(origin)
	log.Info().Msgf("Received DNS query for %s: \n", domain)

	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")

	c := new(dns.Client)

	m := new(dns.Msg)
	m.SetQuestion(domain, r.Question[0].Qtype)
	m.RecursionDesired = true

	server := config.Servers[0]
	port := config.Port

	log.Info().Msgf("Exchange message for domain %s to dns server %s:%s\n", domain, server, port)

	res, _, err := c.Exchange(m, net.JoinHostPort(server, port))

	if res == nil {
		log.Error().Msgf("*** error: %s\n", err.Error())
	}

	if res.Rcode != dns.RcodeSuccess {
		log.Error().Msgf(" *** invalid answer name %s after %d query for %s\n", domain, r.Question[0].Qtype, domain)
	}

	// Stuff must be in the answer section
	for _, a := range res.Answer {
		log.Info().Msgf("%v\n", a)
		msg.Answer = append(msg.Answer, a)
	}

	w.WriteMsg(&msg)
}
