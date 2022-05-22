package dnsserver

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/miekg/dns"
)

func TestShouldRewriteARecord(t *testing.T) {
	s := DnsServer{}
	r1, _ := dns.NewRR("tomcat.default.svc.cluster.local.    5    IN    A    172.21.4.129")
	rr := []dns.RR{r1}
	result := s.convertAnswer("tomcat.", rr)
	require.Equal(t, "tomcat.\t5\tIN\tA\t172.21.4.129", result[0].String())
}

func TestShouldNotRewriteCnameRecord(t *testing.T) {
	s := DnsServer{}
	r1, _ := dns.NewRR("tomcat.com.     465     IN      CNAME   www.tomcat.com.")
	r2, _ := dns.NewRR("www.tomcat.com.     346     IN      A   10.12.4.6")
	rr := []dns.RR{r1, r2}
	result := s.convertAnswer("tomcat.", rr)
	require.Equal(t, "tomcat.\t465\tIN\tCNAME\twww.tomcat.com.", result[0].String())
	require.Equal(t, "www.tomcat.com.\t346\tIN\tA\t10.12.4.6", result[1].String())
}

func TestFetchAllPossibleDomains(t *testing.T) {
	s := DnsServer{}
	s.config = &dns.ClientConfig{}
	s.config.Search = []string{"default.svc.cluster.local", "svc.cluster.local", "cluster.local"}
	domains := s.fetchAllPossibleDomains("alibaba")
	require.Equal(t, 0, len(domains))
	domains = s.fetchAllPossibleDomains("alibaba.")
	require.Equal(t, 2, len(domains))
	require.Equal(t, "alibaba.default.svc.cluster.local.", domains[0])
	require.Equal(t, "alibaba.", domains[1])
	domains = s.fetchAllPossibleDomains("alibaba.ci.")
	require.Equal(t, 3, len(domains))
	require.Equal(t, "alibaba.ci.svc.cluster.local.", domains[0])
	require.Equal(t, "alibaba.ci.default.svc.cluster.local.", domains[1])
	require.Equal(t, "alibaba.ci.", domains[2])
	domains = s.fetchAllPossibleDomains("pod-0.alibaba.ci.")
	require.Equal(t, 3, len(domains))
	require.Equal(t, "pod-0.alibaba.ci.", domains[0])
	require.Equal(t, "pod-0.alibaba.ci.svc.cluster.local.", domains[1])
	require.Equal(t, "pod-0.alibaba.ci.cluster.local.", domains[2])
	domains = s.fetchAllPossibleDomains("pod-0.alibaba.ci.svc.")
	require.Equal(t, 2, len(domains))
	require.Equal(t, "pod-0.alibaba.ci.svc.", domains[0])
	require.Equal(t, "pod-0.alibaba.ci.svc.cluster.local.", domains[1])
	domains = s.fetchAllPossibleDomains("pod-0.alibaba.ci.svc.cluster.local.")
	require.Equal(t, 1, len(domains))
	require.Equal(t, "pod-0.alibaba.ci.svc.cluster.local.", domains[0])
}