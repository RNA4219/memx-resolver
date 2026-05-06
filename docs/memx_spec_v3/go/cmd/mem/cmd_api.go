package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"memx/api"
	"memx/service"
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

		srv := api.NewHTTPServer(svc)
		h := srv.Handler()

		log.Printf("memx API listening on http://%s", *addr)
		log.Fatal(http.ListenAndServe(*addr, h))
	default:
		usage()
		os.Exit(2)
	}
}