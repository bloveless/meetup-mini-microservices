package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConsulResolver implements a custom http.RoundTripper that resolves
// service addresses via Consul.
type ConsulResolver struct {
	client    *api.Client
	transport http.RoundTripper // Underlying transport
}

func NewConsulResolver(consulAddress string) (*ConsulResolver, error) {
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
	return &ConsulResolver{
		client:    client,
		transport: http.DefaultTransport,
	}, nil
}

func (r *ConsulResolver) RoundTrip(req *http.Request) (*http.Response, error) {
	serviceName := req.URL.Hostname() // Extract service name from URL

	// Query Consul for the service
	services, _, err := r.client.Health().Service(serviceName, "", true, &api.QueryOptions{})
	if err != nil {
		return nil, fmt.Errorf("error querying Consul: %w", err)
	}

	if len(services) == 0 {
		// Fallback to regular DNS resolution if the service isn't in Consul
		// This is CRUCIAL for handling external services or situations
		// where Consul might be temporarily unavailable.
		fmt.Println("Service", serviceName, "not found in Consul, falling back to default transport.")

		return r.transport.RoundTrip(req)
	}

	service := services[rand.IntN(len(services))]

	address := service.Service.Address

	// Modify the request to point to the resolved address and port.
	req.URL.Host = fmt.Sprintf("%s", address)

	// Important:  If you're using HTTPS, you'll likely need to modify
	// the req.URL.Scheme to "https" if it isn't already.  You might also
	// need to handle TLS configuration.

	return r.transport.RoundTrip(req)
}

func main() {
	// Get a new client
	consulAddress := "localhost:8500"
	resolver, err := NewConsulResolver(consulAddress)
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Transport: resolver,
		Timeout:   10 * time.Second, // Set a timeout!
	}

	// Example usage:
	req, err := http.NewRequestWithContext(context.Background(), "GET", "http://my-cool-service/check", nil)
	if err != nil {
		panic(err)
	}
	//
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var b bytes.Buffer
	_, err = io.Copy(&b, resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Status Code:", resp.StatusCode)
	fmt.Println("Response: ", b.String())
}
