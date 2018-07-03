package main

import (
	"net"
	"strings"
)

type IpProvider struct {
	addresses []net.IP
}

func (p *IpProvider) getIpAddresses(filter func([]net.IP) []net.IP) []net.IP {
	if len(p.addresses) == 0 {
		var ips []net.IP
		interfaces, _ := net.Interfaces()
		for _, i := range interfaces {
			addresses, _ := i.Addrs()
			for _, addr := range addresses {
				switch v := addr.(type) {
				case *net.IPNet:
					ips = append(ips, v.IP)
				case *net.IPAddr:
					ips = append(ips, v.IP)
				}
			}
		}
		p.addresses = ips
	}
	return filter(p.addresses)
}

func (p *IpProvider) filterIpv4(ips []net.IP) []net.IP {
	var filtered []net.IP
	for _, ip := range ips {
		if !strings.Contains(ip.String(), ":") {
			filtered = append(filtered, ip)
		}
	}
	return filtered
}
