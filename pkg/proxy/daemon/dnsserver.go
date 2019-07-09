package daemon

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

type handler struct{}

func (this *handler) getDomain(origin string) string {
	domain := origin

	namespace, find := os.LookupEnv("PROXY_NAMESPACE")
	if !find {
		namespace = "default"
	}

	index := strings.Index(domain, ".")

	if index+1 == len(domain) {
		domain = domain + namespace + ".svc.cluster.local."
		fmt.Printf("*** Use in cluster dns address %s\n", domain)
	}
	fmt.Printf("Format domain %s to %s\n", origin, domain)

	return domain
}

//ServeDNS query DNS rescord
func (this *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {

	msg := dns.Msg{}
	msg.SetReply(r)
	msg.Authoritative = true

	origin := r.Question[0].Name

	domain := this.getDomain(origin)
	fmt.Printf("Received DNS query for %s: \n", domain)

	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")

	c := new(dns.Client)

	m := new(dns.Msg)
	m.SetQuestion(domain, r.Question[0].Qtype)
	m.RecursionDesired = true

	server := config.Servers[0]
	port := config.Port

	fmt.Printf("Exchange message for domain %s to dns server %s:%s\n", domain, server, port)

	res, _, err := c.Exchange(m, net.JoinHostPort(server, port))

	if res == nil {
		fmt.Printf("*** error: %s\n", err.Error())
	}

	if res.Rcode != dns.RcodeSuccess {
		fmt.Printf(" *** invalid answer name %s after %d query for %s\n", domain, r.Question[0].Qtype, domain)
	}

	// Stuff must be in the answer section
	for _, a := range res.Answer {
		fmt.Printf("%v\n", a)
		msg.Answer = append(msg.Answer, a)
	}

	w.WriteMsg(&msg)
}

// DNSServerStart start dns server
func DNSServerStart() {
	srv := &dns.Server{Addr: ":" + strconv.Itoa(53), Net: "udp"}
	srv.Handler = &handler{}

	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")

	fmt.Printf("Successful load local /etc/resolv.conf")
	for _, server := range config.Servers {
		fmt.Printf("Success load nameserver %s\n", server)
	}

	fmt.Printf("DNS Server Start At 53...\n")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}
