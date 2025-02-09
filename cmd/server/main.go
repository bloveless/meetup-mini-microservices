package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	capi "github.com/hashicorp/consul/api"
)

func main() {
	// Get a new client
	// client, err := capi.NewClient(&capi.Config{Address: "127.0.0.1:8300"})
	client, err := capi.NewClient(capi.DefaultConfig())
	if err != nil {
		panic(err)
	}
	l, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	address := l.Addr().String()
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
			HTTP:     fmt.Sprintf("http://%s/check", address),
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
