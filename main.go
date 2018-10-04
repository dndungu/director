package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strings"
	"time"
)

type source struct {
	hostname string
	scheme   string
}

type target struct {
	address string
	port    int
	scheme  string
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

var hostname string

var targetHost string

var targetPort string

var CommitSha string

func director(r *http.Request) {
	t := target{targetHost, targetPort, HTTP}
	r.URL.Scheme = t.scheme
	if t.port == 80 || t.port == 443 {
		r.URL.Host = t.address
	} else {
		r.URL.Host = fmt.Sprintf("%s:%s", t.address, t.port)
	}
	if _, ok := r.Header["User-Agent"]; !ok {
		r.Header.Set("User-Agent", "")
	}
}

func init() {
	var ok bool
	hostname, ok = os.LookupEnv("HOSTNAME")
	if !ok {
		log.Fatal("HOSTNAME environment variable is not set.")
	}
	targetHost, ok = os.LookupEnv("TARGET_HOST")
	if !ok {
		log.Fatal("TARGET_HOST environment variable is not set.")
	}

	targetPort, ok = os.LookupEnv("TARGET_PORT")
	if !ok {
		log.Fatal("TARGET_PORT environment variable is not set.")
	}

	manager = autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(hostname),
		Cache:      autocert.DirCache("/etc/director/certificates"),
	}
}

var transport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          1000,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
}

func main() {
	proxy := httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	httpServer := http.Server{Addr: ":80", Handler: &proxy}
	go func() {
		log.Print(httpServer.ListenAndServe().Error())
	}()

	httpsServer := &http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			GetCertificate:     manager.GetCertificate,
			InsecureSkipVerify: true,
		},
		Handler: &proxy,
	}

	go func() {
		log.Print(httpsServer.ListenAndServeTLS("", ""))
	}()
	log.Print(fmt.Sprintf("Director version %s is listening on ports :80 and :443", CommitSha))
	<-stop
	httpServer.Shutdown(context.Background())
	httpsServer.Shutdown(context.Background())
}
