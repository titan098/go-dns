package dns

import (
	"encoding/hex"
	"net"
	"strconv"
	"strings"

	"github.com/titan098/go-dns/config"
)

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// NibbleToIPv6 will convert a bind nibble format string to an net.IP object.
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

// IPv6ToNibble converts an IPv6 address to a bind nibble.
func IPv6ToNibble(ip net.IP, prefix int) string {
	encodedBytes := hex.EncodeToString(ip)
	if prefix <= 128 {
		mask := (128 - prefix) / 4
		encodedBytes = encodedBytes[0 : len(encodedBytes)-mask]
	}

	b := reverse(strings.Join(strings.Split(encodedBytes, ""), "."))
	b += ".ip6.arpa."
	return b
}

// SplitPrefix splits an IPv6 prefix into an address and a prefix mask.
func SplitPrefix(prefix string) (string, int) {
	splitAddress := strings.Split(prefix, "/")
	mask, _ := strconv.Atoi(splitAddress[1])
	return splitAddress[0], mask
}

// GetNameForIPv6 constructs a dns name for a passed IPv6 prefix
func GetNameForIPv6(name string, prefix *config.Domain) string {
	p := IPv6ToNibble(net.ParseIP(prefix.Prefix), prefix.Mask)
	digits := strings.TrimSuffix(name, "."+p)
	strippedDigits := reverse(strings.Join(strings.Split(digits, "."), ""))
	return strippedDigits + "." + prefix.ReverseDomain + "."
}

// GetIPv6ForName constructs an IPv6 name for a passed dns name
func GetIPv6ForName(name string, prefix *config.Domain) string {
	p := IPv6ToNibble(net.ParseIP(prefix.Prefix), prefix.Mask)

	digits := strings.TrimSuffix(name, "."+prefix.Domain+".")
	joinedDigits := reverse(strings.Join(strings.Split(digits, ""), ".")) + "." + p
	return NibbleToIPv6(joinedDigits).String()
}
