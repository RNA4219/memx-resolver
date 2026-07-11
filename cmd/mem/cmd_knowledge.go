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

func cmdOutKnowledge(args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(2)
	}
	switch args[0] {
	case "search":
		fs := flag.NewFlagSet("mem out knowledge search", flag.ExitOnError)
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

		resp, apiErr := client.KnowledgeSearch(ctx, api.KnowledgeSearchRequest{Query: query, TopK: *topK})
		if apiErr != nil {
			log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
		}
		if cf.json {
			printJSON(resp)
			return
		}
		for _, n := range resp.Notes {
			pinned := ""
			if n.IsPinned {
				pinned = " [pinned]"
			}
			fmt.Printf("%s\t%s\t[%s]%s\n", n.ID, n.Title, n.WorkingScope, pinned)
		}

	case "show":
		fs := flag.NewFlagSet("mem out knowledge show", flag.ExitOnError)
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

		n, apiErr := client.KnowledgeGet(ctx, id)
		if apiErr != nil {
			log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
		}
		if cf.json {
			printJSON(n)
			return
		}
		pinned := ""
		if n.IsPinned {
			pinned = " (pinned)"
		}
		fmt.Printf("# %s%s\n\nScope: %s\n\n%s\n", n.Title, pinned, n.WorkingScope, n.Body)

	case "list":
		fs := flag.NewFlagSet("mem out knowledge list", flag.ExitOnError)
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

		resp, apiErr := client.KnowledgeListByScope(ctx, api.KnowledgeListByScopeRequest{
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

	case "pinned":
		fs := flag.NewFlagSet("mem out knowledge pinned", flag.ExitOnError)
		cf := &commonFlags{}
		cf.bind(fs)
		scope := fs.String("scope", "", "working scope (optional)")
		limit := fs.Int("limit", 20, "limit")
		_ = fs.Parse(args[1:])

		ctx := context.Background()
		client, cleanup, err := makeClient(ctx, cf)
		if err != nil {
			log.Fatal(err)
		}
		defer cleanup()

		resp, apiErr := client.KnowledgeListPinned(ctx, api.KnowledgeListPinnedRequest{
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
			fmt.Printf("%s\t%s\t[%s]\n", n.ID, n.Title, n.WorkingScope)
		}

	case "pin":
		fs := flag.NewFlagSet("mem out knowledge pin", flag.ExitOnError)
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

		_, apiErr := client.KnowledgePin(ctx, id)
		if apiErr != nil {
			log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
		}
		fmt.Println("ok")

	case "unpin":
		fs := flag.NewFlagSet("mem out knowledge unpin", flag.ExitOnError)
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

		_, apiErr := client.KnowledgeUnpin(ctx, id)
		if apiErr != nil {
			log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
		}
		fmt.Println("ok")

	default:
		usage()
		os.Exit(2)
	}
}
