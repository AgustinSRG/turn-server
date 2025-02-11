// Utils for IP addresses

package main

import (
	"net"
)

func IsPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}

// Detects external IP address from network interfaces
func DetectExternalIPAddress() (*net.IP, error) {
	networkInterfaces, err := net.Interfaces()

	if err != nil {
		return nil, err
	}

	for _, i := range networkInterfaces {
		addrs, err := i.Addrs()

		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}

			if !IsPublicIP(ip) {
				continue
			}

			return &ip, nil
		}
	}

	return nil, nil
}
