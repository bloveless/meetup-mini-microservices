package httpclient

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/hashicorp/consul/api"
)

// consulResolver implements a custom http.RoundTripper that resolves
// service addresses via Consul.
type consulResolver struct {
	client    *api.Client
	transport http.RoundTripper // Underlying transport
}

func newConsulResolver(consulAddress string) (*consulResolver, error) {
	config := api.DefaultConfig()
	if consulAddress != "" {
		config.Address = consulAddress
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("error creating Consul client: %w", err)
	}

	// Use the system's default transport as a fallback.  This is important
	// for cases where you might need to connect to services *not* registered
	// in Consul.
	return &consulResolver{
		client:    client,
		transport: http.DefaultTransport,
	}, nil
}

func (r *consulResolver) RoundTrip(req *http.Request) (*http.Response, error) {
	serviceName := req.URL.Hostname() // Extract service name from URL

	var err error
	var services []*api.ServiceEntry
	for range 3 {
		// Query Consul for the service
		services, _, err = r.client.Health().Service(serviceName, "", true, &api.QueryOptions{})
		if err != nil {
			return nil, fmt.Errorf("error querying Consul: %w", err)
		}

		if len(services) > 0 {
			break
		}

		fmt.Println("Service", serviceName, "not found in Consul, waiting and trying again.")
		time.Sleep(5 * time.Second)
	}

	if len(services) == 0 {
		slog.Warn("service not found in Consul after waiting. proceeding with default transport", slog.String("name", serviceName))
		return r.transport.RoundTrip(req)
	}

	service := services[rand.IntN(len(services))]

	address := service.Service.Address
	port := service.Service.Port

	// Modify the request to point to the resolved address and port.
	req.URL.Host = fmt.Sprintf("%s:%d", address, port)

	return r.transport.RoundTrip(req)
}

func NewConsul(consulAddress string) *http.Client {
	resolver, err := newConsulResolver(consulAddress)
	if err != nil {
		panic(err)
	}
	return &http.Client{
		Transport: resolver,
		Timeout:   10 * time.Second,
	}
}
