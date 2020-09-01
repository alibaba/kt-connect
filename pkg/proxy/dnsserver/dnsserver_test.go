package dnsserver

import (
	"testing"

	"github.com/miekg/dns"
)

func TestAnswerRewrite(t *testing.T) {
	s := &server{}
	actual, _ := dns.NewRR("tomcat.default.svc.cluster.local. 5 IN A 172.21.4.129")
	r, err := s.convertAnswer("tomcat.", "tomcat.default.svc.cluster.local.", actual)
	if err != nil {
		t.Errorf("error")
		return
	}
	if r.String() != "tomcat.	5	IN	A	172.21.4.129" {
		t.Errorf("error")
	}
}

func TestGetDomainWithClusterPostfix(t *testing.T) {
	s := &server{}
	s.config = &dns.ClientConfig{}
	s.config.Search = []string{"default.svc.cluster.local", "svc.cluster.local", "cluster.local"}
	res := s.getDomainWithClusterPostfix("app-svc.", 1)
	if "app-svc.default.svc.cluster.local." != res {
		t.Errorf("error: " + res)
	}
	res = s.getDomainWithClusterPostfix("app-svc.default.", 2)
	if "app-svc.default.svc.cluster.local." != res {
		t.Errorf("error: " + res)
	}
}
