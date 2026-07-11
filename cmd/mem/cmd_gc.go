package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/RNA4219/memx-resolver/v2/api"
)

func cmdGC(args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(2)
	}
	switch args[0] {
	case "short":
		fs := flag.NewFlagSet("mem gc short", flag.ExitOnError)
		cf := &commonFlags{}
		cf.bind(fs)
		dryRun := fs.Bool("dry-run", false, "dry-run mode (show planned operations without executing)")
		enableGC := fs.Bool("enable-gc", false, "enable GC execution (disabled by default)")
		_ = fs.Parse(args[1:])

		ctx := context.Background()
		client, cleanup, err := makeClient(ctx, cf)
		if err != nil {
			log.Fatal(err)
		}
		defer cleanup()

		if !*enableGC && !*dryRun {
			log.Fatal("GC is disabled by default. Use --enable-gc to enable or --dry-run to preview.")
		}

		resp, apiErr := client.GCRun(ctx, api.GCRunRequest{
			Target:  "short",
			Options: api.GCOptions{DryRun: *dryRun},
		})
		if apiErr != nil {
			log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
		}

		if cf.json || *dryRun {
			fmt.Println(resp.Status)
			return
		}
		fmt.Printf("GC completed: %s\n", resp.Status)

	default:
		usage()
		os.Exit(2)
	}
}
