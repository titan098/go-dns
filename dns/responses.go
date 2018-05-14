package dns

import (
	"fmt"
	"strings"

	"bitbucket.org/titan098/go-dns/config"
	"github.com/miekg/dns"
)

// dynamicResponse will create a DNS entry for an IPv6 address, and
// the corresponding IPv6 address for a DNS entry
func dynamicResponse(q dns.Question, prefix *config.Domain) (int, []string) {
	log.Debugf("DynamicResponse handler for: %s", q.Name)
	soa := config.GetConfig().DNS.Soa

	rcode := dns.RcodeSuccess
	answer := []string{}
	name := strings.ToLower(q.Name)
	switch q.Qtype {
	case dns.TypeANY, dns.TypeAAAA:
		if strings.HasSuffix(name, ".ip6.arpa.") {
			// this is a reverse lookup, return the SOA record
			answer = append(answer, soa.String(prefix.Domain))
			rcode = dns.RcodeNameError
		} else {
			// manage the forward lookup, respond for ANY as well
			address := GetIPv6ForName(name, prefix)
			answer = append(answer, fmt.Sprintf("%s AAAA %s", name, address))
		}

	case dns.TypePTR:
		// manage the reverse lookup
		if !strings.HasSuffix(name, ".ip6.arpa.") {
			// this is a reverse lookup, return the SOA record
			answer = append(answer, soa.String(prefix.Domain))
		} else {
			domain := GetNameForIPv6(name, prefix)
			answer = append(answer, fmt.Sprintf("%s PTR %s", name, domain))
		}

	default:
		answer = append(answer, soa.String(config.GetConfig().DNS.Domain.Domain))
		rcode = dns.RcodeNameError
	}
	return rcode, answer
}

// staticResponse handles requests which have been declared as having a fixed
// mapping between name and ip address.
func staticResponse(q dns.Question, prefix *config.Domain) (int, []string) {
	log.Debugf("StaticResponse handler for: %s", q.Name)
	soa := config.GetConfig().DNS.Soa

	rcode := dns.RcodeSuccess
	answer := []string{}
	name := strings.ToLower(q.Name)
	switch q.Qtype {
	case dns.TypeANY, dns.TypeAAAA:
		if strings.HasSuffix(name, ".ip6.arpa.") {
			// this is a reverse lookup, return the SOA record
			answer = append(answer, soa.String(config.GetConfig().DNS.Domain.Domain))
			rcode = dns.RcodeNameError
		} else {
			// manage the forward lookup, respond for ANY as well
			answer = append(answer, fmt.Sprintf("%s AAAA %s", name, prefix.Prefix))
		}

	case dns.TypePTR:
		// manage the reverse lookup
		if !strings.HasSuffix(name, ".ip6.arpa.") {
			// this is a forward lookup, return an SOA record
			answer = append(answer, soa.String(config.GetConfig().DNS.Domain.Domain))
		} else {
			answer = append(answer, fmt.Sprintf("%s PTR %s", name, prefix.ReverseDomain))
		}

	default:
		answer = append(answer, soa.String(config.GetConfig().DNS.Domain.Domain))
		rcode = dns.RcodeNameError
	}
	return rcode, answer
}

// allNxErrorResponse will return NXERROR for every response, this can be used
// as a top level fall through for unknown responses
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

// ResponseFunctions contains the mapptings between the response type config options
// and the response function
var ResponseFunctions = map[string]QueryFunc{
	"Dynamic": dynamicResponse,
	"Static":  staticResponse,
	"NxError": allNxErrorResponse,
}
