package dns

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"

	"bitbucket.org/titan098/go-dns/logging"
	"github.com/miekg/dns"
)

var log = logging.SetupLogging("dns")
var prefix = "2001:470:1f23:ff::"
var mask = 64
var domain = "ipv6.ellefsen.za.net."

type DNSServer struct {
	dns  *dns.Server
	port int
}

func (d *DNSServer) getNameForIPv6(name string) string {
	p := IPv6ToNibble(net.ParseIP(prefix), mask)
	digits := strings.TrimSuffix(name, "."+p)
	strippedDigits := reverse(strings.Join(strings.Split(digits, "."), ""))
	return strippedDigits + "." + domain
}

func (d *DNSServer) getIPv6ForName(name string) string {
	p := IPv6ToNibble(net.ParseIP(prefix), mask)

	digits := strings.TrimSuffix(name, "."+domain)
	joinedDigits := reverse(strings.Join(strings.Split(digits, ""), ".")) + p
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

func (d *DNSServer) parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		log.Debugf("query %d %s", q.Qtype, q.Name)
		switch q.Qtype {
		case dns.TypeAAAA:
			// manage the forward lookup
			address := d.getIPv6ForName(q.Name)
			answer := fmt.Sprintf("%s AAAA %s", q.Name, address)
			log.Debugf("answer %s", answer)
			d.appendAnswer(m, answer)

		case dns.TypePTR:
			// manage the reverse lookup
			domain := d.getNameForIPv6(q.Name)
			answer := fmt.Sprintf("%s PTR %s", q.Name, domain)
			log.Debugf("answer %s", answer)
			d.appendAnswer(m, answer)
		}
	}
}

func (d *DNSServer) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		d.parseQuery(m)
	}

	w.WriteMsg(m)
}

func (d *DNSServer) Start(quit chan struct{}) error {
	// start the DNS sever
	log.Infof("starting dns server on :%d", d.port)

	dns.HandleFunc("ipv6.ellefsen.za.net.", d.handleDNSRequest)
	dns.HandleFunc("f.f.0.0.3.2.f.1.0.7.4.0.1.0.0.2.ip6.arpa.", d.handleDNSRequest)

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

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func NibbleToIPv6(nibble string) net.IP {
	padded := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	stripped := strings.TrimSuffix(nibble, ".ip6.arpa.")
	b := reverse(strings.Join(strings.Split(stripped, "."), ""))
	decodedBytes, _ := hex.DecodeString(b)

	for i, b := range decodedBytes {
		padded[i] = b
	}

	ip := net.IP(padded)
	return ip
}

func IPv6ToNibble(ip net.IP, prefix int) string {
	encodedBytes := hex.EncodeToString(ip)
	encodedBytes = encodedBytes[0:((128 - prefix) / 4)]

	b := reverse(strings.Join(strings.Split(encodedBytes, ""), "."))
	b += ".ip6.arpa."
	return b
}

func SplitPrefix(prefix string) (string, int) {
	splitAddress := strings.Split(prefix, "/")
	mask, _ := strconv.Atoi(splitAddress[1])
	return splitAddress[0], mask
}

func StartServer(quit chan struct{}) *DNSServer {
	port := 15353
	proto := "udp"

	p := net.ParseIP(prefix)
	log.Info(NibbleToIPv6("f.f.0.0.3.2.f.1.0.7.4.0.1.0.0.2.ip6.arpa."))
	log.Info(IPv6ToNibble(p, mask))

	srv := &DNSServer{
		port: port,
		dns:  &dns.Server{Addr: ":" + strconv.Itoa(port), Net: proto}}
	go srv.Start(quit)

	return srv
}
