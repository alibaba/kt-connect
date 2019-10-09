package daemon

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/miekg/dns"
)

// DNSRequestHandler
type DNSRequestHandler struct{}

// ServeDNS start dns server
func ServeDNS() (err error) {
	srv := &dns.Server{Addr: ":" + strconv.Itoa(53), Net: "udp"}
	srv.Handler = &DNSRequestHandler{}

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
func (h *DNSRequestHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
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

	rr = exchange(domain, qtype, msg)
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

func exchange(domain string, Qtype uint16, m *dns.Msg) (rr []dns.RR) {
	address, err := getResolvServer()
	if err != nil {
		log.Error().Msgf(err.Error())
		return
	}
	log.Info().Msgf("Exchange message for domain %s to dns server %s\n", domain, address)

	c := new(dns.Client)
	res, _, err := c.Exchange(m, address)

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
