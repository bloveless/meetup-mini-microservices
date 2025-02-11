package main

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/hashicorp/consul/api"
	"golang.org/x/sync/errgroup"

	"github.com/bloveless/meetup-mini-microservices/internal/httpclient"
	"github.com/bloveless/meetup-mini-microservices/internal/ip"
	"github.com/bloveless/meetup-mini-microservices/internal/service"
)

const (
	consulAddress = "192.168.0.182:8500"
	firstService  = "example-one"
)

func main() {
	// Get a new consul client
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
	deregister, serviceID, err := service.Register(client, "instigator", host, port)
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
	})

	wg := errgroup.Group{}
	wg.Go(func() error {
		for {
			time.Sleep(5 * time.Second)
			s := gofakeit.HackerPhrase()
			slog.Info("fake", slog.Any("data", s))

			cclient := httpclient.NewConsul(consulAddress)
			resp, err := cclient.Post(fmt.Sprintf("http://%s/ring", firstService), "text/plain", strings.NewReader(s))
			if err != nil {
				slog.Error("unable to send instigation request", slog.String("err", err.Error()))
				continue
			}
			defer resp.Body.Close()
			slog.Info("response", slog.Int("code", resp.StatusCode))
		}
	})

	wg.Go(func() error {
		slog.Info("starting server", slog.String("serviceID", serviceID))
		return http.Serve(l, mux)
	})

	wg.Wait()
}
