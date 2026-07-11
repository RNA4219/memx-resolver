package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"memx/api"
)

func cmdIn(args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(2)
	}
	switch args[0] {
	case "short":
		cmdInShort(args[1:])
	case "journal":
		cmdInJournal(args[1:])
	case "knowledge":
		cmdInKnowledge(args[1:])
	default:
		usage()
		os.Exit(2)
	}
}

func cmdInShort(args []string) {
	fs := flag.NewFlagSet("mem in short", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	title := fs.String("title", "", "note title")
	body := fs.String("body", "", "note body")
	summary := fs.String("summary", "", "note summary")
	stdin := fs.Bool("stdin", false, "read body from stdin")
	noLLM := fs.Bool("no-llm", false, "skip LLM auto-summary")
	tags := multiStringFlag{}
	fs.Var(&tags, "tag", "tag (repeatable)")
	sourceType := fs.String("source-type", "manual", "source type")
	origin := fs.String("origin", "", "origin")
	sourceTrust := fs.String("source-trust", "user_input", "source trust")
	sensitivity := fs.String("sensitivity", "internal", "sensitivity")
	_ = fs.Parse(args)

	b := *body
	if *stdin {
		b = readAllStdin()
	}

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	resp, apiErr := client.NotesIngest(ctx, api.NotesIngestRequest{
		Title:       *title,
		Body:        b,
		Summary:     *summary,
		SourceType:  *sourceType,
		Origin:      *origin,
		SourceTrust: *sourceTrust,
		Sensitivity: *sensitivity,
		Tags:        tags,
		NoLLM:       *noLLM,
	})
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}

	if cf.json {
		printJSON(resp)
		return
	}
	fmt.Printf("ok id=%s\n", resp.Note.ID)
}

func cmdInJournal(args []string) {
	fs := flag.NewFlagSet("mem in journal", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	title := fs.String("title", "", "note title")
	body := fs.String("body", "", "note body")
	stdin := fs.Bool("stdin", false, "read body from stdin")
	noLLM := fs.Bool("no-llm", false, "skip LLM auto-summary")
	scope := fs.String("scope", "", "working scope (required)")
	tags := multiStringFlag{}
	fs.Var(&tags, "tag", "tag (repeatable)")
	sourceType := fs.String("source-type", "manual", "source type")
	origin := fs.String("origin", "", "origin")
	sourceTrust := fs.String("source-trust", "user_input", "source trust")
	sensitivity := fs.String("sensitivity", "internal", "sensitivity")
	_ = fs.Parse(args)

	b := *body
	if *stdin {
		b = readAllStdin()
	}

	if *scope == "" {
		log.Fatal("--scope is required for journal")
	}

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	resp, apiErr := client.JournalIngest(ctx, api.JournalIngestRequest{
		Title:        *title,
		Body:         b,
		SourceType:   *sourceType,
		Origin:       *origin,
		SourceTrust:  *sourceTrust,
		Sensitivity:  *sensitivity,
		Tags:         tags,
		WorkingScope: *scope,
		NoLLM:        *noLLM,
	})
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}

	if cf.json {
		printJSON(resp)
		return
	}
	fmt.Printf("ok id=%s\n", resp.Note.ID)
}

func cmdInKnowledge(args []string) {
	fs := flag.NewFlagSet("mem in knowledge", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	title := fs.String("title", "", "note title")
	body := fs.String("body", "", "note body")
	stdin := fs.Bool("stdin", false, "read body from stdin")
	noLLM := fs.Bool("no-llm", false, "skip LLM auto-summary")
	scope := fs.String("scope", "", "working scope (required)")
	tags := multiStringFlag{}
	fs.Var(&tags, "tag", "tag (repeatable)")
	pinned := fs.Bool("pinned", false, "pin the note")
	sourceType := fs.String("source-type", "manual", "source type")
	origin := fs.String("origin", "", "origin")
	sourceTrust := fs.String("source-trust", "user_input", "source trust")
	sensitivity := fs.String("sensitivity", "internal", "sensitivity")
	_ = fs.Parse(args)

	b := *body
	if *stdin {
		b = readAllStdin()
	}

	if *scope == "" {
		log.Fatal("--scope is required for knowledge")
	}

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	resp, apiErr := client.KnowledgeIngest(ctx, api.KnowledgeIngestRequest{
		Title:        *title,
		Body:         b,
		SourceType:   *sourceType,
		Origin:       *origin,
		SourceTrust:  *sourceTrust,
		Sensitivity:  *sensitivity,
		Tags:         tags,
		WorkingScope: *scope,
		IsPinned:     *pinned,
		NoLLM:        *noLLM,
	})
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}

	if cf.json {
		printJSON(resp)
		return
	}
	fmt.Printf("ok id=%s\n", resp.Note.ID)
}