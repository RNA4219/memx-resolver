package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func cmdOut(args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(2)
	}
	switch args[0] {
	case "search":
		cmdOutSearch(args[1:])
	case "show":
		cmdOutShow(args[1:])
	case "recall":
		cmdOutRecall(args[1:])
	case "journal":
		cmdOutJournal(args[1:])
	case "knowledge":
		cmdOutKnowledge(args[1:])
	case "archive":
		cmdOutArchive(args[1:])
	default:
		usage()
		os.Exit(2)
	}
}

func cmdOutSearch(args []string) {
	fs := flag.NewFlagSet("mem out search", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	topK := fs.Int("k", 20, "top k")
	_ = fs.Parse(args)
	query := strings.Join(fs.Args(), " ")

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	resp, apiErr := searchAcrossStores(ctx, client, query, *topK)
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
}

func cmdOutShow(args []string) {
	fs := flag.NewFlagSet("mem out show", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	_ = fs.Parse(args)
	if len(fs.Args()) < 1 {
		log.Fatal("NOTE_ID is required")
	}
	id := fs.Args()[0]

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	n, apiErr := resolveNoteAcrossStores(ctx, client, id)
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}
	if cf.json {
		printJSON(n)
		return
	}
	printResolvedNote(n)
}
