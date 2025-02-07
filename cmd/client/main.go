package main

import (
	"bytes"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"

	capi "github.com/hashicorp/consul/api"
)

func main() {
	// Get a new client
	// client, err := capi.NewClient(&capi.Config{Address: "127.0.0.1:8300"})
	client, err := capi.NewClient(capi.DefaultConfig())
	if err != nil {
		panic(err)
	}
	entries, meta, err := client.Health().Service("my-cool-service", "", true, &capi.QueryOptions{})
	if err != nil {
		panic(err)
	}
	service := entries[rand.Intn(len(entries))]
	address, err := url.Parse("http://" + service.Service.Address + "/check")
	if err != nil {
		panic(err)
	}
	slog.Info("my-cool-service", slog.Any("service address", service.Service.Address), slog.Any("address url", address), slog.Any("meta", meta))

	resp, err := http.Get(address.String())
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var b bytes.Buffer
	_, err = io.Copy(&b, resp.Body)
	if err != nil {
		panic(err)
	}
	slog.Info("response", slog.String("content", b.String()))
}
