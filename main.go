package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strings"
)

type source struct {
	hostname string
	scheme   string
}

type target struct {
	address  string
	hostname string
	port     int
	scheme   string
}

type rule struct {
	source source
	target target
}

const (
	HTTP  = "http"
	HTTPS = "https"
)

var manager autocert.Manager

func findTarget(r *http.Request) target {
	hostArray := strings.Split(r.URL.Host, ".")
	address := fmt.Sprintf("nginx.%s.svc.cluster.local", hostArray[0])
	return target{address, "support.zd-dev.com", 443, HTTPS}
}

func director(r *http.Request) {
	t := findTarget(r)
	r.URL.Scheme = t.scheme
	if t.port == 80 || t.port == 443 {
		r.URL.Host = t.address
	} else {
		r.URL.Host = fmt.Sprintf("%s:%s", t.address, t.port)
	}
	r.Host = t.hostname
	if _, ok := r.Header["User-Agent"]; !ok {
		r.Header.Set("User-Agent", "")
	}
}

func hostPolicy(context.Context, string) error {
	// TODO only allow while listed domains and subdomains
	return nil
}

func init() {
	manager = autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: hostPolicy,
	}
}

func main() {
	proxy := httputil.ReverseProxy{Director: director}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	httpServer := http.Server{Addr: ":80", Handler: &proxy}
	go func() {
		log.Print(httpServer.ListenAndServe().Error())
	}()

	httpsServer := &http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			GetCertificate: manager.GetCertificate,
		},
		Handler: &proxy,
	}

	go func() {
		log.Print(httpsServer.ListenAndServeTLS("", ""))
	}()
	<-stop
	httpServer.Shutdown(context.Background())
	httpsServer.Shutdown(context.Background())
}
