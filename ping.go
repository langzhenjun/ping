package pingo

import (
	"fmt"
	"time"

	"os"

	"errors"
	"net"

	"log"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	ProtocolICMP     = 1
	ProtocolIPv6ICMP = 58
)

func addr(c *icmp.PacketConn, protocol int, host string) (net.Addr, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	netaddr := func(ip net.IP) (net.Addr, error) {
		switch c.LocalAddr().(type) {
		case *net.UDPAddr:
			return &net.UDPAddr{IP: ip}, nil
		case *net.IPAddr:
			return &net.IPAddr{IP: ip}, nil
		default:
			return nil, errors.New("neither UDPAddr nor IPAddr")
		}
	}
	for _, ip := range ips {
		switch protocol {
		case ProtocolICMP:
			if ip.To4() != nil {
				return netaddr(ip)
			}
		case ProtocolIPv6ICMP:
			if ip.To16() != nil && ip.To4() == nil {
				return netaddr(ip)
			}
		}
	}
	return nil, errors.New("no A or AAAA record")
}

func listenPacket(network string) (*icmp.PacketConn, error) {
	switch network {
	case "udp4":
		return icmp.ListenPacket(network, "0.0.0.0")
	case "ip4:icmp":
		return icmp.ListenPacket(network, "0.0.0.0")
	case "udp6":
		return icmp.ListenPacket(network, "::")
	case "ip6:ipv6-icmp":
		return icmp.ListenPacket(network, "::1")
	default:
		return nil, errors.New("not supported network: " + network)
	}
}

func Ping(network string, address string, timeout int, icmpType icmp.Type, protocol int, seq int, message string) (time.Duration, error) {
	c, err := listenPacket(network)
	if err != nil {
		return 0, err
	}
	defer c.Close()

	// change address -> IP
	dst, err := addr(c, protocol, address)
	if err != nil {
		return 0, err
	}

	if network != "udp6" && protocol == ProtocolIPv6ICMP {
		var f ipv6.ICMPFilter
		f.SetAll(true)
		f.Accept(ipv6.ICMPTypeDestinationUnreachable)
		f.Accept(ipv6.ICMPTypePacketTooBig)
		f.Accept(ipv6.ICMPTypeTimeExceeded)
		f.Accept(ipv6.ICMPTypeParameterProblem)
		f.Accept(ipv6.ICMPTypeEchoReply)
		if err := c.IPv6PacketConn().SetICMPFilter(&f); err != nil {
			return 0, err
		}
	}

	wm := icmp.Message{
		Type: icmpType, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1 << uint(seq),
			Data: []byte(message),
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		return 0, err
	}

	start := time.Now()

	if n, err := c.WriteTo(wb, dst); err != nil {
		return 0, err
	} else if n != len(wb) {
		return 0, fmt.Errorf("got %v; want %v", n, len(wb))
	}

	for {
		rb := make([]byte, 1024)
		if err := c.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond)); err != nil {
			return 0, err
		}

		n, peer, err := c.ReadFrom(rb)
		if err != nil {
			return 0, err
		}

		end := time.Now()

		rm, err := icmp.ParseMessage(protocol, rb[:n])
		if err != nil {
			return 0, err
		}
		switch rm.Type {
		case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
			return end.Sub(start), nil
		default:
			log.Printf("got %+v from %v; want echo reply\r\n", rm, peer)
			continue
		}
	}
}
