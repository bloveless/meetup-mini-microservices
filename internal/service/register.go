package service

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	capi "github.com/hashicorp/consul/api"
)

func Register(client *capi.Client, serviceName, host string, port int64) (func(), string, error) {
	slog.Info("service", slog.String("host", host), slog.Int64("port", port))
	serviceID := fmt.Sprintf("%s-%v", serviceName, uuid.New())
	err := client.Agent().ServiceRegister(&capi.AgentServiceRegistration{
		Kind:    capi.ServiceKind("microservice"),
		ID:      serviceID,
		Name:    serviceName,
		Address: host,
		Port:    int(port),
		Check: &capi.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/check", host, port),
			Interval: "10s",
			Timeout:  "30s",
		},
	})
	if err != nil {
		return nil, "", err
	}
	return func() {
		err := client.Agent().ServiceDeregister(serviceID)
		if err != nil {
			panic(err)
		}
	}, serviceID, nil
}
