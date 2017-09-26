package pingo

import (
	"testing"

	"fmt"

	"errors"

	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

type pingTest struct {
	network  string
	address  string
	icmpType icmp.Type
	timeout  int
	protocol int
	message  string
}

var tests = []pingTest{
	{"udp4", "google.com", ipv4.ICMPTypeEcho, 1000, ProtocolICMP,
		"hellohellohellohellohellohellohellohellohellohellohellohellohellohellohellohellohellohellohellohellohellohellohellohellohello"},
	{"ip4:icmp", "google.com", ipv4.ICMPTypeEcho, 1000, ProtocolICMP, "hello"},

	{"udp4", "127.0.0.1", ipv4.ICMPTypeEcho, 1000, ProtocolICMP, ""},
	{"ip4:icmp", "127.0.0.1", ipv4.ICMPTypeEcho, 1000, ProtocolICMP, ""},

	{"udp4", "121.58.225.10", ipv4.ICMPTypeEcho, 1000, ProtocolICMP, ""},
	{"ip4:icmp", "121.58.225.10", ipv4.ICMPTypeEcho, 1000, ProtocolICMP, ""},

	{"udp4", "aliyun.com", ipv4.ICMPTypeEcho, 1000, ProtocolICMP, ""},
	{"ip4:icmp", "aliyun.com", ipv4.ICMPTypeEcho, 1000, ProtocolICMP, ""},

	{"ip4:icmp", "baidu.com", ipv4.ICMPTypeEcho, 1000, ProtocolICMP, ""},

	{"udp6", "::1", ipv6.ICMPTypeEchoRequest, 1000, ProtocolIPv6ICMP, "world"},
	{"ip6:ipv6-icmp", "::1", ipv6.ICMPTypeEchoRequest, 1000, ProtocolIPv6ICMP, ""},
}

var notfound_tests = []pingTest{
	{"ip4:icmp", "notfound.notfound", ipv4.ICMPTypeEcho, 1000, ProtocolICMP, ""},
	{"udp4", "notfound.notfound", ipv4.ICMPTypeEcho, 1000, ProtocolICMP, ""},
}

func TestMe(t *testing.T) {

	for i, tt := range tests {
		delay, err := Ping(
			tt.network,
			tt.address,
			tt.timeout,
			tt.icmpType,
			tt.protocol,
			i,
			tt.message)

		if err != nil {
			t.Error(err.Error())
		} else {
			s := fmt.Sprintf("Seq: %-8d Network: %-16s Addr: %-16s Delay: %-16v",
				i, tt.network, tt.address, delay)
			t.Log(s)
		}
		time.Sleep(time.Millisecond * 100)
	}
}

var noSuchHostError = errors.New("lookup notfound.notfound: no such host")

func TestNotFound(t *testing.T) {
	for i, tt := range notfound_tests {
		_, err := Ping(
			tt.network,
			tt.address,
			tt.timeout,
			tt.icmpType,
			tt.protocol,
			i,
			tt.message)

		if err.Error() == noSuchHostError.Error() {
			t.Log("expected ping error: ", noSuchHostError)
		} else {
			t.Error("unexpected")
		}

		time.Sleep(time.Millisecond * 100)
	}
}
