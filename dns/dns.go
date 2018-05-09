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

// Server is the top-level structure containing a reference to the
// dns server that was constructed
type Server struct {
	dns      *dns.Server
	protocol string
	port     int
}

type QueryFunc func(q dns.Question, prefix *config.Domain) (int, []string)

func getNameForIPv6(name string, prefix *config.Domain) string {
	p := IPv6ToNibble(net.ParseIP(prefix.Prefix), prefix.Mask)
	digits := strings.TrimSuffix(name, "."+p)
	strippedDigits := reverse(strings.Join(strings.Split(digits, "."), ""))
	return strippedDigits + "." + prefix.ReverseDomain + "."
}

func getIPv6ForName(name string, prefix *config.Domain) string {
	p := IPv6ToNibble(net.ParseIP(prefix.Prefix), prefix.Mask)

	digits := strings.TrimSuffix(name, "."+prefix.Domain+".")
	joinedDigits := reverse(strings.Join(strings.Split(digits, ""), ".")) + "." + p
	return NibbleToIPv6(joinedDigits).String()
}

func dynamicResponse(q dns.Question, prefix *config.Domain) (int, []string) {
	log.Debugf("DynamicResponse handler for: %s", q.Name)
	soa := config.GetConfig().DNS.Soa

	rcode := dns.RcodeSuccess
	answer := []string{}
	switch q.Qtype {
	case dns.TypeANY, dns.TypeAAAA:
		// manage the forward lookup, respond for ANY as well
		address := getIPv6ForName(q.Name, prefix)
		answer = append(answer, fmt.Sprintf("%s AAAA %s", q.Name, address))

	case dns.TypePTR:
		// manage the reverse lookup
		domain := getNameForIPv6(q.Name, prefix)
		answer = append(answer, fmt.Sprintf("%s PTR %s", q.Name, domain))

	default:
		answer = append(answer, soa.String(config.GetConfig().DNS.Domain.Domain))
		rcode = dns.RcodeNameError
	}
	return rcode, answer
}

func allNxErrorResponse(q dns.Question, prefix *config.Domain) (int, []string) {
	log.Debugf("AllNxErrorResponse handler for: %s", q.Name)
	soa := config.GetConfig().DNS.Soa

	rcode := dns.RcodeNameError
	answer := []string{}
	switch q.Qtype {

	default:
		answer = append(answer, soa.String(config.GetConfig().DNS.Domain.Domain))
	}
	return rcode, answer
}

var responseFunctions = map[string]QueryFunc{
	"Dynamic": dynamicResponse,
	"NxError": allNxErrorResponse,
}

func getResponseFunction(responseType string) QueryFunc {
	responseFunc, ok := responseFunctions[responseType]
	if ok {
		return responseFunc
	}
	return allNxErrorResponse
}

func (d *Server) appendAnswer(m *dns.Msg, answer string) {
	//send the answer
	rr, err := dns.NewRR(answer)
	if err != nil {
		log.Error("could not construct RR record: " + err.Error())
	}
	m.Answer = append(m.Answer, rr)
}

func (d *Server) parseQuery(m *dns.Msg, prefix *config.Domain, queryfunc QueryFunc) int {
	rcode := dns.RcodeSuccess
	soa := config.GetConfig().DNS.Soa
	ns := config.GetConfig().DNS.Ns

	for _, q := range m.Question {
		log.Debugf("query (%d): '%s'", m.Id, q.String())
		switch q.Qtype {
		case dns.TypeSOA:
			// manage SOA requests
			d.appendAnswer(m, soa.String(config.GetConfig().DNS.Domain.Domain))

		case dns.TypeNS:
			//manage NS requests
			if q.Name != (prefix.Domain + ".") {
				// if we are querying the NS for a host then we know nothing,
				// return the SOA record for the delegated domain.
				answer := soa.String(config.GetConfig().DNS.Domain.Domain)
				d.appendAnswer(m, answer)
				rcode = dns.RcodeNameError
			} else {
				// return the nameservers for this domain
				for _, ns := range ns.Servers {
					answer := fmt.Sprintf("%s NS %s", q.Name, ns)
					d.appendAnswer(m, answer)
				}
			}

		default:
			//fall through return the SOA with NXERROR
			code, answers := queryfunc(q, prefix)
			for _, answer := range answers {
				d.appendAnswer(m, answer)
			}
			rcode = code
		}
	}

	//show the answers
	for _, answer := range m.Answer {
		log.Debugf("answer (%d) [%s]: '%s'", m.Id, dns.RcodeToString[rcode], answer.String())
	}
	return rcode
}

func (d *Server) generateHandleDNSRequest(prefix *config.Domain) (string, func(w dns.ResponseWriter, r *dns.Msg)) {
	log.Infof("creating handler for %s (%s)", prefix.Domain, prefix.ResponseType)
	return prefix.Domain, func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Compress = false
		rcode := dns.RcodeSuccess

		switch r.Opcode {
		case dns.OpcodeQuery:
			rcode = d.parseQuery(m, prefix, getResponseFunction(prefix.ResponseType))
		}
		m.SetRcode(r, rcode)

		w.WriteMsg(m)
	}
}

// Start will perform the start of the DNS server and the setup of the handlers based on the
// configuration values that have been supplied.
func (d *Server) Start(quit chan struct{}) error {
	// start the DNS sever
	log.Infof("starting dns server on :%d (%s)", d.port, d.protocol)
	d.dns = &dns.Server{Addr: ":" + strconv.Itoa(d.port), Net: d.protocol}

	domains := config.GetConfig().SubDomains

	for domain, domainPrefix := range domains {
		reverse := IPv6ToNibble(net.ParseIP(domainPrefix.Prefix), domainPrefix.Mask)
		forwardDomainPrefix := domainPrefix
		reverseDomainPrefix := domainPrefix

		forwardDomainPrefix.Domain = domain
		forwardDomainPrefix.ReverseDomain = reverse
		reverseDomainPrefix.Domain = reverse
		reverseDomainPrefix.ReverseDomain = domain

		dns.HandleFunc(d.generateHandleDNSRequest(&forwardDomainPrefix))
		dns.HandleFunc(d.generateHandleDNSRequest(&reverseDomainPrefix))
	}

	// create fall through for other domains
	topLevelPrefix := config.GetConfig().DNS.Domain
	topLevelReverse := config.GetConfig().DNS.Domain
	reverse := IPv6ToNibble(net.ParseIP(topLevelPrefix.Prefix), topLevelPrefix.Mask)

	topLevelPrefix.ReverseDomain = reverse
	topLevelReverse.Domain = reverse
	topLevelReverse.ReverseDomain = topLevelPrefix.Domain

	dns.HandleFunc(d.generateHandleDNSRequest(&topLevelPrefix))
	dns.HandleFunc(d.generateHandleDNSRequest(&topLevelReverse))

	err := d.dns.ListenAndServe()
	if err != nil {
		log.Errorf("error starting dns server: %s", err.Error())
		quit <- struct{}{}
		return err
	}
	return nil
}

// Close shutdowns down the DNS server in response to a the main process shutting down
func (d *Server) Close() {
	d.dns.Shutdown()
}

// StartServer starts the DNS server in a go routine, returnings a reference to the server
func StartServer(quit chan struct{}) *Server {
	c := config.GetConfig()

	port := c.DNS.Port
	protocol := c.DNS.Protocol
	srv := &Server{
		port:     port,
		protocol: protocol,
	}
	go srv.Start(quit)

	return srv
}

//TODO: Fall through for top level domain to return NXERROR
//TODO: SOA and NS records should work for top level domain
