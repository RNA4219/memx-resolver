package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/RNA4219/memx-resolver/v2/api"
)

func cmdOutJournal(args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(2)
	}
	switch args[0] {
	case "search":
		fs := flag.NewFlagSet("mem out journal search", flag.ExitOnError)
		cf := &commonFlags{}
		cf.bind(fs)
		topK := fs.Int("k", 20, "top k")
		_ = fs.Parse(args[1:])
		query := strings.Join(fs.Args(), " ")

		ctx := context.Background()
		client, cleanup, err := makeClient(ctx, cf)
		if err != nil {
			log.Fatal(err)
		}
		defer cleanup()

		resp, apiErr := client.JournalSearch(ctx, api.JournalSearchRequest{Query: query, TopK: *topK})
		if apiErr != nil {
			log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
		}
		if cf.json {
			printJSON(resp)
			return
		}
		for _, n := range resp.Notes {
			fmt.Printf("%s\t%s\t[%s]\n", n.ID, n.Title, n.WorkingScope)
		}

	case "show":
		fs := flag.NewFlagSet("mem out journal show", flag.ExitOnError)
		cf := &commonFlags{}
		cf.bind(fs)
		_ = fs.Parse(args[1:])
		if len(fs.Args()) < 1 {
			log.Fatal("ID is required")
		}
		id := fs.Args()[0]

		ctx := context.Background()
		client, cleanup, err := makeClient(ctx, cf)
		if err != nil {
			log.Fatal(err)
		}
		defer cleanup()

		n, apiErr := client.JournalGet(ctx, id)
		if apiErr != nil {
			log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
		}
		if cf.json {
			printJSON(n)
			return
		}
		fmt.Printf("# %s\n\nScope: %s\n\n%s\n", n.Title, n.WorkingScope, n.Body)

	case "list":
		fs := flag.NewFlagSet("mem out journal list", flag.ExitOnError)
		cf := &commonFlags{}
		cf.bind(fs)
		scope := fs.String("scope", "", "working scope (required)")
		limit := fs.Int("limit", 20, "limit")
		_ = fs.Parse(args[1:])

		if *scope == "" {
			log.Fatal("--scope is required")
		}

		ctx := context.Background()
		client, cleanup, err := makeClient(ctx, cf)
		if err != nil {
			log.Fatal(err)
		}
		defer cleanup()

		resp, apiErr := client.JournalListByScope(ctx, api.JournalListByScopeRequest{
			WorkingScope: *scope,
			Limit:        *limit,
		})
		if apiErr != nil {
			log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
		}
		if cf.json {
			printJSON(resp)
			return
		}
		for _, n := range resp.Notes {
			fmt.Printf("%s\t%s\n", n.ID, n.Title)
		}

	default:
		usage()
		os.Exit(2)
	}
}
