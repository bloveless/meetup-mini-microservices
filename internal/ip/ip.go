package ip

import (
	"log/slog"
	"net"
	"strings"
)

func Address() string {
	var address string
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			panic(err)
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.To4() != nil && !strings.HasPrefix(ip.To4().String(), "127") {
				address = ip.String()
				slog.Info("found ip v4 address", slog.String("ip", ip.String()), slog.String("ipv4", string(ip.To4())), slog.String("ipv6", string(ip.To16())))
			}
		}
	}
	return address
}
