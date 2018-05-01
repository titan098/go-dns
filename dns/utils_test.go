package dns

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitPrefix(t *testing.T) {
	prefix := "2003::/8"
	address, mask := SplitPrefix(prefix)
	assert.Equal(t, "2003::", address)
	assert.Equal(t, 8, mask)
}

func TestIPv6ToNibble(t *testing.T) {
	address := "2001:db8::1"
	result := IPv6ToNibble(net.ParseIP(address), 128)
	result64 := IPv6ToNibble(net.ParseIP(address), 64)
	result80 := IPv6ToNibble(net.ParseIP(address), 80)
	assert.Equal(t, "1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa.", result)
	assert.Equal(t, "0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa.", result64)
	assert.Equal(t, "0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa.", result80)
}

func TestNibbleToIPv6(t *testing.T) {
	result128 := NibbleToIPv6("1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa.")
	result60 := NibbleToIPv6("0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa.")
	assert.Equal(t, "2001:db8::1", result128.String())
	assert.Equal(t, "2001:db8::", result60.String())
}
