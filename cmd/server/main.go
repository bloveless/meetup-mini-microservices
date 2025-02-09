package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	capi "github.com/hashicorp/consul/api"
)

const (
	myIPAddress = "192.168.0.182"
)

func main() {
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

	// Get a new client
	client, err := capi.NewClient(&capi.Config{
		Address: fmt.Sprintf("%s:8500", myIPAddress),
	})
	if err != nil {
		panic(err)
	}
	l, err := net.Listen("tcp", fmt.Sprintf("%s:0", address))
	if err != nil {
		panic(err)
	}
	defer l.Close()
	host, portRaw, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		panic(err)
	}
	port, err := strconv.ParseInt(portRaw, 10, 32)
	if err != nil {
		panic(err)
	}
	slog.Info("service", slog.String("host", host), slog.Int64("port", port))
	serviceID := fmt.Sprintf("my-cool-service-%v", uuid.New())
	err = client.Agent().ServiceRegister(&capi.AgentServiceRegistration{
		Kind:    capi.ServiceKind("microservice"),
		ID:      serviceID,
		Name:    "my-cool-service",
		Address: address,
		Port:    int(port),
		Check: &capi.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/check", host, port),
			Interval: "10s",
			Timeout:  "30s",
		},
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		err := client.Agent().ServiceDeregister(serviceID)
		if err != nil {
			panic(err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /check", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok " + serviceID))
	})

	slog.Info("starting server", slog.String("serviceID", serviceID))
	if err := http.Serve(l, mux); err != nil {
		panic(err)
	}
}
