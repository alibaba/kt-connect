package daemon

import (
	"testing"
	"github.com/miekg/dns"
)

func TestAnswerRewrite(t *testing.T) {
	acutal, _ := dns.NewRR("tomcat.default.svc.cluster.local. 5 IN A 172.21.4.129")
	r, err := getAnswer("tomcat.", "tomcat.default.svc.cluster.local.", acutal)
	if err != nil {
		t.Errorf("error")
		return
	}
	if r.String() != "tomcat.	5	IN	A	172.21.4.129" {
		t.Errorf("error")
	}
}