package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"memx/api"
)

func cmdDocs(args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(2)
	}
	switch args[0] {
	case "ingest":
		cmdDocsIngest(args[1:])
	case "resolve":
		cmdDocsResolve(args[1:])
	case "chunks":
		cmdDocsChunks(args[1:])
	case "search":
		cmdDocsSearch(args[1:])
	case "ack":
		cmdDocsAck(args[1:])
	case "stale":
		cmdDocsStale(args[1:])
	case "contract":
		cmdDocsContract(args[1:])
	default:
		usage()
		os.Exit(2)
	}
}

func cmdDocsIngest(args []string) {
	fs := flag.NewFlagSet("mem docs ingest", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	title := fs.String("title", "", "document title")
	body := fs.String("body", "", "document body")
	docType := fs.String("doc-type", "", "document type (spec, requirement, design, etc.)")
	version := fs.String("version", "", "document version")
	sourcePath := fs.String("source-path", "", "source file path")
	var features flagsStringSlice
	fs.Var(&features, "feature", "feature keys (can be repeated)")
	var tags flagsStringSlice
	fs.Var(&tags, "tag", "tags (can be repeated)")
	_ = fs.Parse(args)

	if *title == "" || *body == "" || *docType == "" {
		log.Fatal("--title, --body, and --doc-type are required")
	}
	if *version == "" {
		*version = "1"
	}

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	req := api.DocsIngestRequest{
		Title:       *title,
		Body:        *body,
		DocType:     *docType,
		Version:     *version,
		SourcePath:  *sourcePath,
		FeatureKeys: features,
		Tags:        tags,
	}
	resp, apiErr := client.DocsIngest(ctx, req)
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}
	if cf.json {
		printJSON(resp)
		return
	}
	fmt.Printf("ingested: %s (version %s, %d chunks)\n", resp.DocID, resp.Version, resp.ChunkCount)
}

func cmdDocsResolve(args []string) {
	fs := flag.NewFlagSet("mem docs resolve", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	feature := fs.String("feature", "", "feature key")
	taskID := fs.String("task-id", "", "task ID")
	topic := fs.String("topic", "", "topic query")
	limit := fs.Int("limit", 10, "max results")
	_ = fs.Parse(args)

	if *feature == "" && *taskID == "" && *topic == "" {
		log.Fatal("one of --feature, --task-id, or --topic is required")
	}

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	req := api.DocsResolveRequest{
		Feature: *feature,
		TaskID:  *taskID,
		Topic:   *topic,
		Limit:   *limit,
	}
	resp, apiErr := client.DocsResolve(ctx, req)
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}
	if cf.json {
		printJSON(resp)
		return
	}
	if len(resp.Required) > 0 {
		fmt.Println("Required:")
		for _, e := range resp.Required {
			fmt.Printf("  %s\t%s\t%s\n", e.DocID, e.Version, e.Title)
		}
	}
	if len(resp.Recommended) > 0 {
		fmt.Println("Recommended:")
		for _, e := range resp.Recommended {
			fmt.Printf("  %s\t%s\t%s\n", e.DocID, e.Version, e.Title)
		}
	}
}

func cmdDocsChunks(args []string) {
	fs := flag.NewFlagSet("mem docs chunks", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	docID := fs.String("doc-id", "", "document ID")
	heading := fs.String("heading", "", "filter by heading")
	query := fs.String("query", "", "filter by query")
	limit := fs.Int("limit", 0, "max chunks")
	_ = fs.Parse(args)

	if *docID == "" {
		log.Fatal("--doc-id is required")
	}

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	req := api.ChunksGetRequest{
		DocID:   *docID,
		Heading: *heading,
		Query:   *query,
		Limit:   *limit,
	}
	resp, apiErr := client.ChunksGet(ctx, req)
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}
	if cf.json {
		printJSON(resp)
		return
	}
	for _, c := range resp.Chunks {
		fmt.Printf("## %s\n%s\n\n", c.Heading, c.Body)
	}
}

func cmdDocsSearch(args []string) {
	fs := flag.NewFlagSet("mem docs search", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	limit := fs.Int("limit", 10, "max results")
	_ = fs.Parse(args)
	if len(fs.Args()) < 1 {
		log.Fatal("query is required")
	}
	query := fs.Args()[0]

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	req := api.DocsSearchRequest{Query: query, Limit: *limit}
	resp, apiErr := client.DocsSearch(ctx, req)
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}
	if cf.json {
		printJSON(resp)
		return
	}
	for _, e := range resp.Results {
		fmt.Printf("%s\t%s\t%s\n", e.DocID, e.Version, e.Title)
	}
}

func cmdDocsAck(args []string) {
	fs := flag.NewFlagSet("mem docs ack", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	taskID := fs.String("task-id", "", "task ID")
	docID := fs.String("doc-id", "", "document ID")
	version := fs.String("version", "", "document version")
	_ = fs.Parse(args)

	if *taskID == "" || *docID == "" {
		log.Fatal("--task-id and --doc-id are required")
	}

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	req := api.ReadsAckRequest{
		TaskID:  *taskID,
		DocID:   *docID,
		Version: *version,
		Reader:  "cli",
	}
	resp, apiErr := client.ReadsAck(ctx, req)
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}
	if cf.json {
		printJSON(resp)
		return
	}
	fmt.Printf("acknowledged: %s -> %s (%s)\n", resp.Receipt.TaskID, resp.Receipt.DocID, resp.Receipt.Version)
}

func cmdDocsStale(args []string) {
	fs := flag.NewFlagSet("mem docs stale", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	taskID := fs.String("task-id", "", "task ID")
	_ = fs.Parse(args)

	if *taskID == "" {
		log.Fatal("--task-id is required")
	}

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	req := api.DocsStaleCheckRequest{TaskID: *taskID}
	resp, apiErr := client.DocsStaleCheck(ctx, req)
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}
	if cf.json {
		printJSON(resp)
		return
	}
	if len(resp.Stale) == 0 {
		fmt.Printf("task %s is fresh\n", resp.TaskID)
		return
	}
	fmt.Printf("task %s is stale:\n", resp.TaskID)
	for _, s := range resp.Stale {
		fmt.Printf("  %s: %s -> %s (%s)\n", s.DocID, s.PreviousVersion, s.CurrentVersion, s.Reason)
	}
}

func cmdDocsContract(args []string) {
	fs := flag.NewFlagSet("mem docs contract", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	feature := fs.String("feature", "", "feature key")
	taskID := fs.String("task-id", "", "task ID")
	_ = fs.Parse(args)

	if *feature == "" && *taskID == "" {
		log.Fatal("one of --feature or --task-id is required")
	}

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	req := api.ContractsResolveRequest{
		Feature: *feature,
		TaskID:  *taskID,
	}
	resp, apiErr := client.ContractsResolve(ctx, req)
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}
	if cf.json {
		printJSON(resp)
		return
	}
	if len(resp.AcceptanceCriteria) > 0 {
		fmt.Println("Acceptance Criteria:")
		for _, a := range resp.AcceptanceCriteria {
			fmt.Printf("  - %s\n", a)
		}
	}
	if len(resp.ForbiddenPatterns) > 0 {
		fmt.Println("Forbidden Patterns:")
		for _, f := range resp.ForbiddenPatterns {
			fmt.Printf("  - %s\n", f)
		}
	}
	if len(resp.DefinitionOfDone) > 0 {
		fmt.Println("Definition of Done:")
		for _, d := range resp.DefinitionOfDone {
			fmt.Printf("  - %s\n", d)
		}
	}
}