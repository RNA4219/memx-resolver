package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/RNA4219/memx-resolver/v2/api"
	"github.com/RNA4219/memx-resolver/v2/service"
)

func cmdAPI(args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(2)
	}
	switch args[0] {
	case "serve":
		fs := flag.NewFlagSet("mem api serve", flag.ExitOnError)
		cf := &commonFlags{}
		cf.bind(fs)
		addr := fs.String("addr", "127.0.0.1:7766", "listen address")
		maxRequestBytes := fs.Int64("max-request-bytes", api.DefaultMaxRequestBytes, "maximum JSON request body size")
		allowNonLoopback := fs.Bool("allow-non-loopback", false, "acknowledge unauthenticated non-loopback exposure")
		_ = fs.Parse(args[1:])

		svc, err := service.New(cf.paths())
		if err != nil {
			log.Fatal(err)
		}
		if err := attachOpenAIFromEnv(svc); err != nil {
			_ = svc.Close()
			log.Fatal(err)
		}
		defer func() { _ = svc.Close() }()

		if *maxRequestBytes <= 0 {
			log.Fatal("max-request-bytes must be greater than zero")
		}
		srv := api.NewHTTPServer(svc)
		srv.MaxRequestBytes = *maxRequestBytes
		h := srv.Handler()

		if err := validateBindAddress(*addr, *allowNonLoopback); err != nil {
			log.Print(err)
			_ = svc.Close()
			os.Exit(2)
		}
		if !isLoopbackAddress(*addr) {
			log.Printf("WARNING: non-loopback API has no authentication or TLS: %s", *addr)
		}

		server := &http.Server{
			Addr:              *addr,
			Handler:           h,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       120 * time.Second,
		}
		shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		go func() {
			<-shutdownSignal.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				log.Printf("API shutdown failed: %v", err)
			}
		}()

		log.Printf("memx API listening on http://%s", *addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	default:
		usage()
		os.Exit(2)
	}
}
func isLoopbackAddress(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func validateBindAddress(addr string, allowNonLoopback bool) error {
	if isLoopbackAddress(addr) || allowNonLoopback {
		return nil
	}
	return fmt.Errorf("ERROR: non-loopback API requires --allow-non-loopback: %s", addr)
}
