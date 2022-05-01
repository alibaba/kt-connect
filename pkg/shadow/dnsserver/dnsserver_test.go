package dnsserver

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/miekg/dns"
)

func TestShouldRewriteARecord(t *testing.T) {
	s := &DnsServer{}
	r1, _ := dns.NewRR("tomcat.default.svc.cluster.local.    5    IN    A    172.21.4.129")
	rr := []dns.RR{r1}
	result := s.convertAnswer("tomcat.", rr)
	require.Equal(t, "tomcat.\t5\tIN\tA\t172.21.4.129", result[0].String())
}

func TestShouldNotRewriteCnameRecord(t *testing.T) {
	s := &DnsServer{}
	r1, _ := dns.NewRR("tomcat.com.     465     IN      CNAME   www.tomcat.com.")
	r2, _ := dns.NewRR("www.tomcat.com.     346     IN      A   10.12.4.6")
	rr := []dns.RR{r1, r2}
	result := s.convertAnswer("tomcat.", rr)
	require.Equal(t, "tomcat.\t465\tIN\tCNAME\twww.tomcat.com.", result[0].String())
	require.Equal(t, "www.tomcat.com.\t346\tIN\tA\t10.12.4.6", result[1].String())
}
