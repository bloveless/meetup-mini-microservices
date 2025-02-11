package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/api"

	"github.com/bloveless/meetup-mini-microservices/internal/httpclient"
	"github.com/bloveless/meetup-mini-microservices/internal/ip"
	"github.com/bloveless/meetup-mini-microservices/internal/service"
)

const (
	consulAddress = "192.168.0.182:8500"
)

func main() {
	var algorithm string
	flag.StringVar(&algorithm, "algorithm", "one", `algorithm "one" or "two"`)
	flag.Parse()
	slog.Info("starting server", slog.String("algorithm", algorithm))
	// Get a new client
	client, err := api.NewClient(&api.Config{
		Address: consulAddress,
	})
	if err != nil {
		panic(err)
	}
	address := ip.Address()
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
	deregister, serviceID, err := service.Register(client, fmt.Sprintf("example-%s", algorithm), host, port)
	if err != nil {
		panic(err)
	}
	defer deregister()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /check", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok " + serviceID))
	})
	mux.HandleFunc("POST /ring", func(w http.ResponseWriter, r *http.Request) {
		// Receive something alter it and send it to the next
		var b bytes.Buffer
		_, err := io.Copy(&b, r.Body)
		if err != nil {
			panic(err)
		}
		slog.Info("received", slog.String("body", b.String()))
		nextService := "example-two"
		var output []string
		if algorithm == "one" {
			output = strings.Split(b.String(), " ")
			slices.Reverse(output)
		}

		if algorithm == "two" {
			nextService = "instigator"
			inputParts := strings.Split(b.String(), " ")
			for _, word := range inputParts {
				runes := []rune(word)
				slices.Reverse(runes)
				output = append(output, string(runes))
			}
		}
		slog.Info("sending", slog.String("nextService", nextService), slog.String("output", strings.Join(output, " ")))
		cclient := httpclient.NewConsul(consulAddress)
		resp, err := cclient.Post(fmt.Sprintf("http://%s/ring", nextService), "text/html", strings.NewReader(strings.Join(output, " ")))
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		var respB bytes.Buffer
		_, err = io.Copy(&respB, resp.Body)
		if err != nil {
			panic(err)
		}
		slog.Info("received", slog.Int("statusCode", resp.StatusCode), slog.String("body", b.String()))
	})

	slog.Info("starting server", slog.String("serviceID", serviceID))
	if err := http.Serve(l, mux); err != nil {
		panic(err)
	}
}
