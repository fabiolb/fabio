package config

import (
	"log"
	"net"
)

// LocalIP tries to determine a non-loopback address for the local machine
func LocalIP() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
			if ipnet.IP.To4() != nil || ipnet.IP.To16() != nil {
				return ipnet.IP, nil
			}
		}
	}
	return nil, nil
}

func LocalIPString() string {
	ip, err := LocalIP()
	if err != nil {
		log.Print("[WARN] Error determining local ip address. ", err)
		return ""
	}
	if ip == nil {
		log.Print("[WARN] Could not determine local ip address")
		return ""
	}
	return ip.String()
}
