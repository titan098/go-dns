package dns

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"bitbucket.org/titan098/go-dns/config"
	"bitbucket.org/titan098/go-dns/logging"
	"github.com/miekg/dns"
)

var log = logging.SetupLogging("dns")

type DNSServer struct {
	dns      *dns.Server
	protocol string
	port     int
}

func (d *DNSServer) getNameForIPv6(name string, prefix *config.Domain) string {
	p := IPv6ToNibble(net.ParseIP(prefix.Prefix), prefix.Mask)
	digits := strings.TrimSuffix(name, "."+p)
	strippedDigits := reverse(strings.Join(strings.Split(digits, "."), ""))
	return strippedDigits + "." + prefix.Domain + "."
}

func (d *DNSServer) getIPv6ForName(name string, prefix *config.Domain) string {
	p := IPv6ToNibble(net.ParseIP(prefix.Prefix), prefix.Mask)

	digits := strings.TrimSuffix(name, "."+prefix.Domain+".")
	joinedDigits := reverse(strings.Join(strings.Split(digits, ""), ".")) + "." + p
	return NibbleToIPv6(joinedDigits).String()
}

func (d *DNSServer) appendAnswer(m *dns.Msg, answer string) {
	//send the answer
	rr, err := dns.NewRR(answer)
	if err != nil {
		log.Error("could not construct RR record: " + err.Error())
	}
	m.Answer = append(m.Answer, rr)
}

func (d *DNSServer) parseQuery(m *dns.Msg, prefix *config.Domain, soa *config.Soa, ns *config.Ns) {
	for _, q := range m.Question {
		log.Debugf("query (%d): '%s'", m.Id, q.String())
		switch q.Qtype {
		case dns.TypeSOA:
			// manage SOA requests
			d.appendAnswer(m, soa.String(prefix.Domain))

		case dns.TypeNS:
			//manage NS requests
			if q.Name != prefix.Domain {
				// if we are querying the NS for a host then we know nothing,
				// return the SOA record for the delegated domain.
				answer := soa.String(prefix.Domain)
				d.appendAnswer(m, answer)
			} else {
				// return the nameservers for this domain
				for _, ns := range ns.Servers {
					answer := fmt.Sprintf("%s NS %s", q.Name, ns)
					d.appendAnswer(m, answer)
				}
			}

		case dns.TypeAAAA:
			// manage the forward lookup
			address := d.getIPv6ForName(q.Name, prefix)
			answer := fmt.Sprintf("%s AAAA %s", q.Name, address)
			d.appendAnswer(m, answer)

		case dns.TypePTR:
			// manage the reverse lookup
			domain := d.getNameForIPv6(q.Name, prefix)
			answer := fmt.Sprintf("%s PTR %s", q.Name, domain)
			d.appendAnswer(m, answer)
		}
	}

	//show the answers
	for _, answer := range m.Answer {
		log.Debugf("answer (%d): '%s'", m.Id, answer.String())
	}
}

func (d *DNSServer) generateHandleDNSRequest(domain string, prefix *config.Domain, soa *config.Soa, ns *config.Ns) (string, func(w dns.ResponseWriter, r *dns.Msg)) {
	log.Infof("creating handler for %s", domain)
	return domain, func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Compress = false

		switch r.Opcode {
		case dns.OpcodeQuery:
			d.parseQuery(m, prefix, soa, ns)
		}

		w.WriteMsg(m)
	}
}

func (d *DNSServer) Start(quit chan struct{}) error {
	// start the DNS sever
	log.Infof("starting dns server on :%d (%s)", d.port, d.protocol)
	d.dns = &dns.Server{Addr: ":" + strconv.Itoa(d.port), Net: d.protocol}

	dnsConfig := config.GetConfig().DNS
	domains := config.GetConfig().Domains

	for domain, domainPrefix := range domains {
		reverse := IPv6ToNibble(net.ParseIP(domainPrefix.Prefix), domainPrefix.Mask)
		forwardDomainPrefix := domainPrefix
		reverseDomainPrefix := domainPrefix

		forwardDomainPrefix.Domain = domain
		reverseDomainPrefix.Domain = reverse

		dns.HandleFunc(d.generateHandleDNSRequest(domain, &forwardDomainPrefix, &dnsConfig.Soa, &dnsConfig.Ns))
		dns.HandleFunc(d.generateHandleDNSRequest(reverse, &reverseDomainPrefix, &dnsConfig.Soa, &dnsConfig.Ns))
	}

	err := d.dns.ListenAndServe()
	if err != nil {
		log.Errorf("error starting dns server: %s", err.Error())
		quit <- struct{}{}
		return err
	}
	return nil
}

func (d *DNSServer) Close() {
	d.dns.Shutdown()
}

func StartServer(quit chan struct{}) *DNSServer {
	c := config.GetConfig()

	port := c.DNS.Port
	protocol := c.DNS.Protocol
	srv := &DNSServer{
		port:     port,
		protocol: protocol,
	}
	go srv.Start(quit)

	return srv
}
