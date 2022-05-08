package common

import (
	"fmt"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	"strconv"
	"sync"
	"time"
)

type NsEntry struct {
	answer []dns.RR
	timestamp int64
}

// domain to ip cache
var nsCache = sync.Map{}

// SetupDnsServer start dns server on specified port
func SetupDnsServer(dnsHandler dns.Handler, port int, net string) error {
	log.Info().Msgf("Creating %s dns on port %d", net, port)
	srv := &dns.Server{
		Addr: ":" + strconv.Itoa(port),
		Net: net,
		Handler: dnsHandler,
	}
	// process will hang at here
	return srv.ListenAndServe()
}

// NsLookup query domain record, dnsServerAddr use '<ip>:<port>' format
func NsLookup(domain string, qtype uint16, net, dnsServerAddr string) (*dns.Msg, error) {
	c := new(dns.Client)
	c.Net = net
	msg := new(dns.Msg)
	msg.RecursionDesired = true
	msg.SetQuestion(domain, qtype)
	res, _, err := c.Exchange(msg, dnsServerAddr)
	if err != nil {
		return nil, err
	}
	if res.Rcode == dns.RcodeNameError {
		return nil, DomainNotExistError{name: domain, qtype: qtype}
	} else if res.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("response code %d", res.Rcode)
	}
	return res, nil
}

// ReadCache fetch from cache
func ReadCache(domain string, qtype uint16, ttl int64) []dns.RR {
	if record, exists := nsCache.Load(getCacheKey(domain, qtype)); exists && notExpired(record.(NsEntry).timestamp, ttl) {
		return record.(NsEntry).answer
	}
	return nil
}

// WriteCache record to cache
func WriteCache(domain string, qtype uint16, answer []dns.RR, timestamp int64) {
	nsCache.Store(getCacheKey(domain, qtype), NsEntry{answer, timestamp})
}

func notExpired(timestamp int64, ttl int64) bool {
	return time.Now().Unix() < timestamp + ttl
}

func getCacheKey(domain string, qtype uint16) string {
	return fmt.Sprintf("%s:%d", domain, qtype)
}
