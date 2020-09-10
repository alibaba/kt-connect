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
		t.Errorf("error, get result: " + r.String())
	}
}

func TestGetFirstPart(t *testing.T) {
	s := &server{}
	r := s.getFirstPart("abc.def.com.")
	if r != "abc." {
		t.Errorf("error, get result: " + r)
	}
}

func TestGetFirst2Parts(t *testing.T) {
	s := &server{}
	r := s.getFirst2Parts("abc.def.com.")
	if r != "abc.def." {
		t.Errorf("error, get result:" + r)
	}
}
