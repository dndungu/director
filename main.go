package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
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

var domains []string

var manager autocert.Manager

var defaultTarget = target{"zatiti.com", "zatiti.com", 80, HTTP}

func findTarget(u *url.URL) (t *target, err error) {
	// TODO fetch this from the data store
	t = &defaultTarget
	return t, err
}

func director(r *http.Request) {
	t, err := findTarget(r.URL)
	if err != nil {
		t = &defaultTarget
	}
	r.URL.Scheme = t.scheme
	if t.port == 80 || t.port == 443 {
		r.URL.Host = t.address
	} else {
		r.URL.Host = fmt.Sprintf("%s:%s", t.address, t.port)
	}
	r.Header.Set("Host", t.hostname)
	if _, ok := r.Header["User-Agent"]; !ok {
		r.Header.Set("User-Agent", "")
	}
}

func hostPolicy(context.Context, string) error {
	// TODO only allow while listed domains and subdomains
	return nil
}

func init() {
	domains = []string{"zatiti.com"}
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

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(rootCerts)

	httpsServer := &http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			GetCertificate: manager.GetCertificate,
			RootCAs:        pool,
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
