package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/RNA4219/memx-resolver/v2/api"
)

func cmdOutArchive(args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(2)
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("mem out archive list", flag.ExitOnError)
		cf := &commonFlags{}
		cf.bind(fs)
		limit := fs.Int("limit", 20, "limit")
		_ = fs.Parse(args[1:])

		ctx := context.Background()
		client, cleanup, err := makeClient(ctx, cf)
		if err != nil {
			log.Fatal(err)
		}
		defer cleanup()

		resp, apiErr := client.ArchiveList(ctx, api.ArchiveListRequest{Limit: *limit})
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

	case "show":
		fs := flag.NewFlagSet("mem out archive show", flag.ExitOnError)
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

		n, apiErr := client.ArchiveGet(ctx, id)
		if apiErr != nil {
			log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
		}
		if cf.json {
			printJSON(n)
			return
		}
		fmt.Printf("# %s\n\n%s\n", n.Title, n.Body)

	case "restore":
		fs := flag.NewFlagSet("mem out archive restore", flag.ExitOnError)
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

		resp, apiErr := client.ArchiveRestore(ctx, id)
		if apiErr != nil {
			log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
		}
		if cf.json {
			printJSON(resp)
			return
		}
		fmt.Printf("restored: %s\n", resp.Note.ID)

	default:
		usage()
		os.Exit(2)
	}
}
