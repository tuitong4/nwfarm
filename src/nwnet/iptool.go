package iptool

import (
	"net"
	"strings"
)

type V4Prefix net.IPNet
type V6Prefix net.IPNet

func (ipv4 *V4Prefix) Numeric() uint32 {

	t := uint32(ipv4.IP[12]) * 0x1000000
	t += uint32(ipv4.IP[13]) * 0x10000
	t += uint32(ipv4.IP[14]) * 0x100
	t += uint32(ipv4.IP[15])

	return t
}

func (ipv4 *V4Prefix) NetAddress() net.IP {
	return ipv4.IP.Mask(ipv4.Mask)
}

func IPv4(ipv4 string) (*V4Prefix, error) {

	if !strings.Contains(ipv4, "/") {
		ipv4 = ipv4 + "/32"
	}

	ip, ipnet, err := net.ParseCIDR(ipv4)
	if err != nil {
		return nil, err
	}

	return &V4Prefix{ip, ipnet.Mask}, nil
}

func (ipv6 *V6Prefix) NetAddress() net.IP {
	return ipv6.IP.Mask(ipv6.Mask)
}

func IPv6(ipv6 string) (*V6Prefix, error) {

	if !strings.Contains(ipv6, "/") {
		ipv6 = ipv6 + "/128"
	}

	ip, ipnet, err := net.ParseCIDR(ipv6)
	if err != nil {
		return nil, err
	}

	return &V6Prefix{ip, ipnet.Mask}, nil
}
