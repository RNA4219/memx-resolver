package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"memx/api"
	"memx/db"
	"memx/service"
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
	case "cards":
		cmdDocsCards(args[1:])
	case "cards-feedback":
		cmdDocsCardsFeedback(args[1:])
	case "bundle":
		cmdDocsBundle(args[1:])
	case "taskstate-export":
		cmdDocsTaskStateExport(args[1:])
	case "migrate-resolver-store":
		cmdDocsMigrateResolverStore(args[1:])
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

type resolverStoreMigrationReport struct {
	From     string         `json:"from"`
	To       string         `json:"to"`
	DryRun   bool           `json:"dry_run"`
	Counts   map[string]int `json:"counts"`
	Copied   map[string]int `json:"copied,omitempty"`
	Status   string         `json:"status"`
	Warnings []string       `json:"warnings,omitempty"`
}

func cmdDocsMigrateResolverStore(args []string) {
	fs := flag.NewFlagSet("mem docs migrate-resolver-store", flag.ExitOnError)
	from := fs.String("from", "", "source DB path that contains resolver tables, usually short.db")
	to := fs.String("to", "", "destination resolver.db path")
	dryRun := fs.Bool("dry-run", false, "show migration plan without copying")
	jsonOut := fs.Bool("json", false, "output JSON")
	_ = fs.Parse(args)
	if *from == "" || *to == "" {
		log.Fatal("--from and --to are required")
	}
	report, err := migrateResolverStore(*from, *to, *dryRun)
	if err != nil {
		log.Fatal(err)
	}
	if *jsonOut {
		printJSON(report)
		return
	}
	if report.DryRun {
		fmt.Printf("resolver store migration dry-run: %s -> %s\n", report.From, report.To)
		for _, table := range resolverMigrationTables() {
			fmt.Printf("  %s: %d rows\n", table, report.Counts[table])
		}
		return
	}
	fmt.Printf("resolver store migrated: %s -> %s\n", report.From, report.To)
	for _, table := range resolverMigrationTables() {
		fmt.Printf("  %s: %d rows\n", table, report.Copied[table])
	}
}

func migrateResolverStore(from string, to string, dryRun bool) (resolverStoreMigrationReport, error) {
	initializer, err := service.New(db.Paths{Short: from, Resolver: to})
	if err != nil {
		return resolverStoreMigrationReport{}, err
	}
	_ = initializer.Close()

	src, err := sql.Open("sqlite", fmt.Sprintf("file:%s", from))
	if err != nil {
		return resolverStoreMigrationReport{}, err
	}
	defer src.Close()
	counts, err := countResolverTables(src)
	if err != nil {
		return resolverStoreMigrationReport{}, err
	}
	report := resolverStoreMigrationReport{
		From:   from,
		To:     to,
		DryRun: dryRun,
		Counts: counts,
		Status: "planned",
	}
	if dryRun {
		return report, nil
	}

	dst, err := sql.Open("sqlite", fmt.Sprintf("file:%s", to))
	if err != nil {
		return resolverStoreMigrationReport{}, err
	}
	defer dst.Close()
	copied, err := copyResolverTables(dst, from)
	if err != nil {
		return resolverStoreMigrationReport{}, err
	}
	report.Copied = copied
	report.Status = "migrated"
	return report, nil
}

func resolverMigrationTables() []string {
	return []string{
		"resolver_documents",
		"resolver_chunks",
		"resolver_document_links",
		"resolver_read_receipts",
		"resolver_memory_card_feedback",
		"resolver_audit_log",
	}
}

func countResolverTables(conn *sql.DB) (map[string]int, error) {
	counts := make(map[string]int)
	for _, table := range resolverMigrationTables() {
		var count int
		if err := conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", table)).Scan(&count); err != nil {
			return nil, fmt.Errorf("count %s: %w", table, err)
		}
		counts[table] = count
	}
	return counts, nil
}

func copyResolverTables(dst *sql.DB, sourcePath string) (map[string]int, error) {
	if _, err := dst.Exec(`PRAGMA foreign_keys = OFF;`); err != nil {
		return nil, err
	}
	if _, err := dst.Exec(fmt.Sprintf("ATTACH DATABASE '%s' AS src;", sourcePath)); err != nil {
		return nil, fmt.Errorf("attach source: %w", err)
	}
	defer func() { _, _ = dst.Exec(`DETACH DATABASE src;`) }()
	tx, err := dst.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	for _, table := range []string{"resolver_audit_log", "resolver_memory_card_feedback", "resolver_read_receipts", "resolver_document_links", "resolver_chunks", "resolver_documents"} {
		if _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s;", table)); err != nil {
			return nil, fmt.Errorf("clear %s: %w", table, err)
		}
	}
	statements := map[string]string{
		"resolver_documents": `
INSERT INTO resolver_documents(doc_id, doc_type, title, source_path, version, version_scheme, updated_at, summary, body, tags_json, feature_keys_json, task_ids_json, tracker_refs_json, birdseye_refs_json, acceptance_criteria_json, forbidden_patterns_json, definition_of_done_json, dependencies_json, importance)
SELECT doc_id, doc_type, title, source_path, version, version_scheme, updated_at, summary, body, tags_json, feature_keys_json, task_ids_json, tracker_refs_json, birdseye_refs_json, acceptance_criteria_json, forbidden_patterns_json, definition_of_done_json, dependencies_json, importance FROM src.resolver_documents;`,
		"resolver_chunks": `
INSERT INTO resolver_chunks(chunk_id, doc_id, heading, heading_path_json, ordinal, body, token_estimate, importance)
SELECT chunk_id, doc_id, heading, heading_path_json, ordinal, body, token_estimate, importance FROM src.resolver_chunks;`,
		"resolver_document_links": `
INSERT INTO resolver_document_links(src_doc_id, dst_doc_id, link_type)
SELECT src_doc_id, dst_doc_id, link_type FROM src.resolver_document_links;`,
		"resolver_read_receipts": `
INSERT INTO resolver_read_receipts(task_id, doc_id, version, chunk_ids_json, chunk_snapshots_json, previous_receipt_hash, receipt_hash, reader, read_at)
SELECT task_id, doc_id, version, chunk_ids_json, chunk_snapshots_json, previous_receipt_hash, receipt_hash, reader, read_at FROM src.resolver_read_receipts;`,
		"resolver_memory_card_feedback": `
INSERT INTO resolver_memory_card_feedback(card_id, doc_id, chunk_id, memory_type, signal, weight, query, recorded_at)
SELECT card_id, doc_id, chunk_id, memory_type, signal, weight, query, recorded_at FROM src.resolver_memory_card_feedback;`,
		"resolver_audit_log": `
INSERT INTO resolver_audit_log(operation, actor, target_type, target_id, result, receipt_hash, details_json, recorded_at)
SELECT operation, actor, target_type, target_id, result, receipt_hash, details_json, recorded_at FROM src.resolver_audit_log;`,
	}
	copied := make(map[string]int)
	for _, table := range resolverMigrationTables() {
		res, err := tx.Exec(statements[table])
		if err != nil {
			return nil, fmt.Errorf("copy %s: %w", table, err)
		}
		rows, _ := res.RowsAffected()
		copied[table] = int(rows)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	if _, err := dst.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return nil, err
	}
	return copied, nil
}

func cmdDocsIngest(args []string) {
	fs := flag.NewFlagSet("mem docs ingest", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	title := fs.String("title", "", "document title")
	body := fs.String("body", "", "document body")
	docType := fs.String("doc-type", "", "document type (spec, requirement, design, etc.)")
	version := fs.String("version", "", "document version")
	versionScheme := fs.String("version-scheme", "", "version comparison scheme (semver, iso_datetime, git_revision, string)")
	sourcePath := fs.String("source-path", "", "source file path")
	var features flagsStringSlice
	fs.Var(&features, "feature", "feature keys (can be repeated)")
	var tags flagsStringSlice
	fs.Var(&tags, "tag", "tags (can be repeated)")
	var trackerRefs flagsStringSlice
	fs.Var(&trackerRefs, "tracker-ref", "tracker typed refs such as tracker:issue:github:owner/repo#123 (can be repeated)")
	var birdseyeRefs flagsStringSlice
	fs.Var(&birdseyeRefs, "birdseye-ref", "Birdseye view refs or node IDs (can be repeated)")
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
		Title:         *title,
		Body:          *body,
		DocType:       *docType,
		Version:       *version,
		VersionScheme: *versionScheme,
		SourcePath:    *sourcePath,
		FeatureKeys:   features,
		Tags:          tags,
		TrackerRefs:   trackerRefs,
		BirdseyeRefs:  birdseyeRefs,
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
	var chunkIDs flagsStringSlice
	fs.Var(&chunkIDs, "chunk-id", "chunk IDs (can be repeated)")
	heading := fs.String("heading", "", "filter by heading")
	query := fs.String("query", "", "filter by query")
	limit := fs.Int("limit", 0, "max chunks")
	_ = fs.Parse(args)

	if *docID == "" && len(chunkIDs) == 0 {
		log.Fatal("--doc-id or --chunk-id is required")
	}

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	req := api.ChunksGetRequest{
		DocID:    *docID,
		Heading:  *heading,
		Query:    *query,
		Limit:    *limit,
		ChunkIDs: chunkIDs,
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
	var docTypes flagsStringSlice
	fs.Var(&docTypes, "doc-type", "document types to include (can be repeated)")
	var tags flagsStringSlice
	fs.Var(&tags, "tag", "tags to include (can be repeated)")
	var features flagsStringSlice
	fs.Var(&features, "feature", "feature keys to include (can be repeated)")
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

	req := api.DocsSearchRequest{
		Query:       query,
		DocTypes:    docTypes,
		Tags:        tags,
		FeatureKeys: features,
		Limit:       *limit,
	}
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

func cmdDocsCards(args []string) {
	fs := flag.NewFlagSet("mem docs cards", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	query := fs.String("query", "", "card query")
	limit := fs.Int("limit", 10, "max cards")
	tokenBudget := fs.Int("token-budget", 0, "max estimated tokens across cards")
	weightQueryExact := fs.Int("weight-query-exact", 0, "override exact query match score")
	weightQueryTerms := fs.Int("weight-query-terms", 0, "override all query terms match score")
	weightMemoryTypeBase := fs.Int("weight-memory-type-base", 0, "override memory type rank multiplier")
	weightFeedbackBoost := fs.Int("weight-feedback-boost", 0, "override feedback boost multiplier")
	var docTypes flagsStringSlice
	fs.Var(&docTypes, "doc-type", "document types to include (can be repeated)")
	var tags flagsStringSlice
	fs.Var(&tags, "tag", "tags to include (can be repeated)")
	var features flagsStringSlice
	fs.Var(&features, "feature", "feature keys to include (can be repeated)")
	var memoryTypes flagsStringSlice
	fs.Var(&memoryTypes, "memory-type", "memory types to include (can be repeated)")
	_ = fs.Parse(args)

	if *query == "" {
		log.Fatal("--query is required")
	}

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	req := api.CardsSearchRequest{
		Query:       *query,
		DocTypes:    docTypes,
		Tags:        tags,
		FeatureKeys: features,
		MemoryTypes: memoryTypes,
		Limit:       *limit,
		TokenBudget: *tokenBudget,
		RankingWeights: api.MemoryCardRankingWeights{
			QueryExact:     *weightQueryExact,
			QueryTerms:     *weightQueryTerms,
			MemoryTypeBase: *weightMemoryTypeBase,
			FeedbackBoost:  *weightFeedbackBoost,
		},
	}
	resp, apiErr := client.CardsSearch(ctx, req)
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}
	if cf.json {
		printJSON(resp)
		return
	}
	for _, card := range resp.Cards {
		fmt.Printf("[%s] %s\n%s\n(source: %s / %s, score: %d)\n\n", card.MemoryType, card.Cue, card.Statement, card.DocID, card.ChunkID, card.Score)
	}
}

func cmdDocsCardsFeedback(args []string) {
	fs := flag.NewFlagSet("mem docs cards-feedback", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	cardID := fs.String("card-id", "", "memory card ID")
	docID := fs.String("doc-id", "", "document ID")
	chunkID := fs.String("chunk-id", "", "chunk ID")
	memoryType := fs.String("memory-type", "", "memory type")
	signal := fs.String("signal", "", "usage signal (used, helpful, pinned, irrelevant, skipped)")
	weight := fs.Int("weight", 1, "signal weight")
	query := fs.String("query", "", "query that produced the card")
	_ = fs.Parse(args)

	if *cardID == "" || *signal == "" {
		log.Fatal("--card-id and --signal are required")
	}

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	resp, apiErr := client.CardFeedback(ctx, api.CardFeedbackRequest{
		CardID:     *cardID,
		DocID:      *docID,
		ChunkID:    *chunkID,
		MemoryType: *memoryType,
		Signal:     *signal,
		Weight:     *weight,
		Query:      *query,
	})
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}
	if cf.json {
		printJSON(resp)
		return
	}
	fmt.Printf("recorded feedback: %s %s x%d\n", resp.Feedback.CardID, resp.Feedback.Signal, resp.Feedback.Weight)
}

func cmdDocsBundle(args []string) {
	fs := flag.NewFlagSet("mem docs bundle", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	query := fs.String("query", "", "card query")
	feature := fs.String("feature", "", "feature key")
	taskID := fs.String("task-id", "", "task ID")
	limit := fs.Int("limit", 10, "max cards")
	tokenBudget := fs.Int("token-budget", 0, "max estimated tokens across cards")
	format := fs.String("format", "markdown", "bundle format (markdown or jsonl)")
	var memoryTypes flagsStringSlice
	fs.Var(&memoryTypes, "memory-type", "memory types to include (can be repeated)")
	_ = fs.Parse(args)

	if *query == "" && *feature == "" && *taskID == "" {
		log.Fatal("one of --query, --feature, or --task-id is required")
	}

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	resp, apiErr := client.PromptBundle(ctx, api.PromptBundleRequest{
		Query:       *query,
		Feature:     *feature,
		TaskID:      *taskID,
		MemoryTypes: memoryTypes,
		Limit:       *limit,
		TokenBudget: *tokenBudget,
		Format:      *format,
	})
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}
	if cf.json {
		printJSON(resp)
		return
	}
	fmt.Print(resp.Prompt)
}

func cmdDocsTaskStateExport(args []string) {
	fs := flag.NewFlagSet("mem docs taskstate-export", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	taskID := fs.String("task-id", "", "task ID")
	feature := fs.String("feature", "", "feature key")
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

	resp, apiErr := client.TaskStateExport(ctx, api.TaskStateExportRequest{TaskID: *taskID, Feature: *feature})
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}
	if cf.json {
		printJSON(resp)
		return
	}
	fmt.Printf("taskstate export: %s (%d receipts, %d stale reasons)\n", resp.TaskRef, len(resp.ReadReceipts), len(resp.StaleReasons))
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
	staleReasons := resp.StaleReasons
	if len(staleReasons) == 0 {
		staleReasons = resp.Stale
	}
	if resp.Status == "fresh" || (resp.Status == "" && len(staleReasons) == 0) {
		fmt.Printf("task %s is fresh\n", resp.TaskID)
		return
	}
	fmt.Printf("task %s is stale:\n", resp.TaskID)
	for _, s := range staleReasons {
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
