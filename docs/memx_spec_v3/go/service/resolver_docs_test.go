package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"memx/db"
)

func newResolverServiceForTest(t *testing.T) *Service {
	t.Helper()
	tmpDir := t.TempDir()
	svc, err := New(db.Paths{
		Short:     filepath.Join(tmpDir, "short.db"),
		Journal:   filepath.Join(tmpDir, "journal.db"),
		Knowledge: filepath.Join(tmpDir, "knowledge.db"),
		Archive:   filepath.Join(tmpDir, "archive.db"),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return svc
}

func TestResolverDocsLifecycle(t *testing.T) {
	svc := newResolverServiceForTest(t)
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	doc, chunkCount, err := svc.DocsIngest(ctx, DocsIngestRequest{
		DocType:     "spec",
		Title:       "Memory Import Spec",
		SourcePath:  "docs/specs/memory-import.md",
		Version:     "2026-03-10",
		FeatureKeys: []string{"memory-import"},
		TaskIDs:     []string{"task:feature:local:123"},
		Tags:        []string{"memory", "import"},
		Body: `# Memory Import Spec

## Acceptance Criteria
- imports can be replayed safely

## Forbidden Patterns
- do not skip stale check

## Definition of Done
- contract response contains required docs

## Dependencies
- doc:spec:shared-schema`,
	})
	if err != nil {
		t.Fatalf("DocsIngest: %v", err)
	}
	if chunkCount < 4 {
		t.Fatalf("expected multiple chunks, got %d", chunkCount)
	}
	if doc.DocID == "" {
		t.Fatal("expected generated doc id")
	}

	required, recommended, err := svc.DocsResolve(ctx, DocsResolveRequest{Feature: "memory-import"})
	if err != nil {
		t.Fatalf("DocsResolve: %v", err)
	}
	if len(required) != 1 || required[0].DocID != doc.DocID {
		t.Fatalf("unexpected required docs: %#v", required)
	}
	if len(recommended) != 0 {
		t.Fatalf("unexpected recommended docs: %#v", recommended)
	}

	_, chunks, err := svc.ChunksGet(ctx, ChunksGetRequest{DocID: doc.DocID, Heading: "Acceptance", Limit: 5})
	if err != nil {
		t.Fatalf("ChunksGet: %v", err)
	}
	if len(chunks) != 1 || chunks[0].Importance != "required" {
		t.Fatalf("unexpected chunks: %#v", chunks)
	}

	receipt, err := svc.ReadsAck(ctx, ReadsAckRequest{TaskID: "task:feature:local:123", DocID: doc.DocID, ChunkIDs: []string{chunks[0].ChunkID}})
	if err != nil {
		t.Fatalf("ReadsAck: %v", err)
	}
	if receipt.Version != "2026-03-10" {
		t.Fatalf("unexpected receipt version: %#v", receipt)
	}

	_, _, err = svc.DocsIngest(ctx, DocsIngestRequest{
		DocID:       doc.DocID,
		DocType:     "spec",
		Title:       "Memory Import Spec",
		SourcePath:  "docs/specs/memory-import.md",
		Version:     "2026-03-11",
		FeatureKeys: []string{"memory-import"},
		TaskIDs:     []string{"task:feature:local:123"},
		Body: `# Memory Import Spec

## Acceptance Criteria
- imports can be replayed safely`,
	})
	if err != nil {
		t.Fatalf("DocsIngest update: %v", err)
	}

	stale, err := svc.DocsStaleCheck(ctx, DocsStaleCheckRequest{TaskID: "task:feature:local:123"})
	if err != nil {
		t.Fatalf("DocsStaleCheck: %v", err)
	}
	if len(stale) != 1 || stale[0].CurrentVersion != "2026-03-11" {
		t.Fatalf("unexpected stale response: %#v", stale)
	}
	if stale[0].Reason != "version_mismatch" || len(stale[0].ImpactScope) == 0 {
		t.Fatalf("expected version stale metadata impact, got %#v", stale[0])
	}

	required, acceptance, forbidden, done, dependencies, err := svc.ContractsResolve(ctx, ContractsResolveRequest{TaskID: "task:feature:local:123"})
	if err != nil {
		t.Fatalf("ContractsResolve: %v", err)
	}
	if len(required) != 1 {
		t.Fatalf("unexpected required docs: %#v", required)
	}
	if len(acceptance) != 1 || acceptance[0] != "imports can be replayed safely" {
		t.Fatalf("unexpected acceptance criteria: %#v", acceptance)
	}
	if len(forbidden) != 0 {
		t.Fatalf("expected no forbidden patterns after update, got %#v", forbidden)
	}
	if len(done) != 0 {
		t.Fatalf("expected no definition of done after update, got %#v", done)
	}
	if len(dependencies) != 0 {
		t.Fatalf("expected no dependencies after update, got %#v", dependencies)
	}
}

func TestDocsSearchSupportsStructuredFilters(t *testing.T) {
	svc := newResolverServiceForTest(t)
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	specDoc, _, err := svc.DocsIngest(ctx, DocsIngestRequest{
		DocType:     "spec",
		Title:       "Search Filter Spec",
		Version:     "2026-03-10",
		FeatureKeys: []string{"resolver-search"},
		Tags:        []string{"memory"},
		Body:        "# Search Filter Spec\n\nresolver filter behavior",
	})
	if err != nil {
		t.Fatalf("DocsIngest spec: %v", err)
	}
	_, _, err = svc.DocsIngest(ctx, DocsIngestRequest{
		DocType:     "runbook",
		Title:       "Search Filter Runbook",
		Version:     "2026-03-10",
		FeatureKeys: []string{"resolver-runbook"},
		Tags:        []string{"memory"},
		Body:        "# Search Filter Runbook\n\nresolver filter behavior",
	})
	if err != nil {
		t.Fatalf("DocsIngest runbook: %v", err)
	}

	results, err := svc.DocsSearch(ctx, DocsSearchRequest{
		Query:       "resolver filter",
		DocTypes:    []string{"spec"},
		FeatureKeys: []string{"resolver-search"},
		Tags:        []string{"memory"},
	})
	if err != nil {
		t.Fatalf("DocsSearch: %v", err)
	}
	if len(results) != 1 || results[0].DocID != specDoc.DocID {
		t.Fatalf("unexpected filtered results: %#v", results)
	}

	results, err = svc.DocsSearch(ctx, DocsSearchRequest{
		Query:       "resolver filter",
		DocTypes:    []string{"spec"},
		FeatureKeys: []string{"resolver-runbook"},
	})
	if err != nil {
		t.Fatalf("DocsSearch mismatch: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no results for mismatched filters, got %#v", results)
	}
}

func TestDocsSearchUsesResolverFTSAcrossDocumentFields(t *testing.T) {
	svc := newResolverServiceForTest(t)
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	doc, _, err := svc.DocsIngest(ctx, DocsIngestRequest{
		DocType: "spec",
		Title:   "Alpha Resolver Spec",
		Version: "2026-03-10",
		Body:    "# Search\n\nBudget behavior is described here.",
	})
	if err != nil {
		t.Fatalf("DocsIngest: %v", err)
	}

	results, err := svc.DocsSearch(ctx, DocsSearchRequest{Query: "alpha budget"})
	if err != nil {
		t.Fatalf("DocsSearch: %v", err)
	}
	if len(results) != 1 || results[0].DocID != doc.DocID || results[0].Reason != "fts matched" {
		t.Fatalf("expected FTS result across title/body fields, got %#v", results)
	}
}

func TestCardsSearchUsesResolverChunkFTSAcrossHeadingAndBody(t *testing.T) {
	svc := newResolverServiceForTest(t)
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	_, _, err := svc.DocsIngest(ctx, DocsIngestRequest{
		DocType: "spec",
		Title:   "Chunk FTS Spec",
		Version: "2026-03-10",
		Body: `# Chunk FTS Spec

## Acceptance Criteria
- budget item is prompt ready

## Background
- unrelated acceptance note`,
	})
	if err != nil {
		t.Fatalf("DocsIngest: %v", err)
	}

	cards, err := svc.CardsSearch(ctx, CardsSearchRequest{Query: "acceptance budget", Limit: 5})
	if err != nil {
		t.Fatalf("CardsSearch: %v", err)
	}
	if len(cards) == 0 {
		t.Fatal("expected cards from chunk FTS")
	}
	if cards[0].MemoryType != "acceptance" || cards[0].Statement != "budget item is prompt ready" {
		t.Fatalf("expected chunk FTS to prefer matching acceptance card, got %#v", cards)
	}
}

func TestDocsStaleCheckReportsSemanticDiffImpact(t *testing.T) {
	svc := newResolverServiceForTest(t)
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	doc, _, err := svc.DocsIngest(ctx, DocsIngestRequest{
		DocType: "spec",
		Title:   "Semantic Stale Spec",
		Version: "2026-03-10",
		TaskIDs: []string{"task:semantic"},
		Body: `# Semantic Stale Spec

## Acceptance Criteria
- old behavior must be preserved`,
	})
	if err != nil {
		t.Fatalf("DocsIngest: %v", err)
	}
	_, chunks, err := svc.ChunksGet(ctx, ChunksGetRequest{DocID: doc.DocID, Heading: "Acceptance"})
	if err != nil {
		t.Fatalf("ChunksGet: %v", err)
	}
	if _, err := svc.ReadsAck(ctx, ReadsAckRequest{TaskID: "task:semantic", DocID: doc.DocID, ChunkIDs: []string{chunks[0].ChunkID}}); err != nil {
		t.Fatalf("ReadsAck: %v", err)
	}
	if _, _, err := svc.DocsIngest(ctx, DocsIngestRequest{
		DocID:   doc.DocID,
		DocType: "spec",
		Title:   "Semantic Stale Spec",
		Version: "2026-03-11",
		TaskIDs: []string{"task:semantic"},
		Body: `# Semantic Stale Spec

## Acceptance Criteria
- new behavior must be preserved`,
	}); err != nil {
		t.Fatalf("DocsIngest update: %v", err)
	}

	stale, err := svc.DocsStaleCheck(ctx, DocsStaleCheckRequest{TaskID: "task:semantic"})
	if err != nil {
		t.Fatalf("DocsStaleCheck: %v", err)
	}
	if len(stale) != 1 || stale[0].Reason != "semantic_diff" || stale[0].Severity == "" {
		t.Fatalf("expected semantic stale reason, got %#v", stale)
	}
	if len(stale[0].ChangedChunks) != 1 || len(stale[0].ImpactScope) == 0 {
		t.Fatalf("expected changed chunk impact scope, got %#v", stale[0])
	}
}

func TestReadsAckStoresHashChainAndAuditLog(t *testing.T) {
	svc := newResolverServiceForTest(t)
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	doc, _, err := svc.DocsIngest(ctx, DocsIngestRequest{DocType: "spec", Title: "Receipt Audit Spec", Version: "2026-03-10", Body: "# Spec"})
	if err != nil {
		t.Fatalf("DocsIngest: %v", err)
	}
	first, err := svc.ReadsAck(ctx, ReadsAckRequest{TaskID: "task:audit", DocID: doc.DocID, Reader: "agent-a"})
	if err != nil {
		t.Fatalf("ReadsAck first: %v", err)
	}
	second, err := svc.ReadsAck(ctx, ReadsAckRequest{TaskID: "task:audit", DocID: doc.DocID, Reader: "agent-a"})
	if err != nil {
		t.Fatalf("ReadsAck second: %v", err)
	}
	if first.ReceiptHash == "" {
		t.Fatalf("expected first receipt hash: %#v", first)
	}
	if second.PreviousReceiptHash != first.ReceiptHash || second.ReceiptHash == "" || second.ReceiptHash == first.ReceiptHash {
		t.Fatalf("unexpected hash chain: first=%#v second=%#v", first, second)
	}

	var auditCount int
	if err := svc.resolverDB().QueryRow(`
SELECT COUNT(*)
FROM resolver_audit_log
WHERE operation = 'reads_ack' AND receipt_hash = ? AND result = 'ok';
`, second.ReceiptHash).Scan(&auditCount); err != nil {
		t.Fatalf("query audit log: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("expected one audit row for second receipt, got %d", auditCount)
	}
}

func TestChunksExposePromptReadyMemoryCards(t *testing.T) {
	svc := newResolverServiceForTest(t)
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	doc, _, err := svc.DocsIngest(ctx, DocsIngestRequest{
		DocType: "spec",
		Title:   "Memory Card Spec",
		Version: "2026-03-10",
		Body: `# Memory Card Spec

## Acceptance Criteria
- cards expose prompt-ready statements
- cards keep provenance`,
	})
	if err != nil {
		t.Fatalf("DocsIngest: %v", err)
	}
	_, chunks, err := svc.ChunksGet(ctx, ChunksGetRequest{DocID: doc.DocID, Heading: "Acceptance"})
	if err != nil {
		t.Fatalf("ChunksGet: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("expected one acceptance chunk, got %#v", chunks)
	}
	if chunks[0].MemoryType != "acceptance" || chunks[0].Cue == "" {
		t.Fatalf("expected chunk memory metadata, got %#v", chunks[0])
	}

	cards := BuildResolverMemoryCards(chunks, 10)
	if len(cards) != 2 {
		t.Fatalf("expected list items to become cards, got %#v", cards)
	}
	if cards[0].MemoryType != "acceptance" || cards[0].ChunkID != chunks[0].ChunkID || cards[0].Statement == "" {
		t.Fatalf("unexpected memory card: %#v", cards[0])
	}
}

func TestMemoryCardsRankByTypeQueryAndBudget(t *testing.T) {
	chunks := []ResolverChunk{
		{
			ChunkID:       "chunk:doc:test:001",
			DocID:         "doc:test",
			Heading:       "Background",
			HeadingPath:   []string{"Spec", "Background"},
			Body:          "- general background\n- another reference",
			TokenEstimate: 20,
			Importance:    "reference",
			MemoryType:    "reference",
			Cue:           "Spec > Background",
		},
		{
			ChunkID:       "chunk:doc:test:002",
			DocID:         "doc:test",
			Heading:       "Acceptance Criteria",
			HeadingPath:   []string{"Spec", "Acceptance Criteria"},
			Body:          "- resolver cards should rank acceptance query matches first\n- very long unrelated acceptance item that should not fit a tiny token budget because it has many many many many many words",
			TokenEstimate: 40,
			Importance:    "required",
			MemoryType:    "acceptance",
			Cue:           "Spec > Acceptance Criteria",
		},
	}

	cards := BuildRankedResolverMemoryCards(chunks, "resolver cards", 10, 20)

	if len(cards) == 0 {
		t.Fatal("expected ranked cards")
	}
	if cards[0].MemoryType != "acceptance" || cards[0].Statement != "resolver cards should rank acceptance query matches first" {
		t.Fatalf("expected matching acceptance card first, got %#v", cards[0])
	}
	totalTokens := 0
	for _, card := range cards {
		totalTokens += card.TokenEstimate
		if card.TokenEstimate > 20 {
			t.Fatalf("card should respect token budget, got %#v", card)
		}
	}
	if totalTokens > 20 {
		t.Fatalf("cards exceed token budget: %d", totalTokens)
	}
}

func TestMemoryCardsRankingWeightsCanOverrideQueryDominance(t *testing.T) {
	chunks := []ResolverChunk{
		{
			ChunkID:       "chunk:doc:test:001",
			DocID:         "doc:test",
			HeadingPath:   []string{"Spec", "Acceptance Criteria"},
			Body:          "- plain acceptance item",
			TokenEstimate: 10,
			Importance:    "required",
			MemoryType:    "acceptance",
			Cue:           "Spec > Acceptance Criteria",
		},
		{
			ChunkID:       "chunk:doc:test:002",
			DocID:         "doc:test",
			HeadingPath:   []string{"Spec", "Reference"},
			Body:          "- exact-query-match",
			TokenEstimate: 10,
			Importance:    "reference",
			MemoryType:    "reference",
			Cue:           "Spec > Reference",
		},
	}

	cards := BuildRankedResolverMemoryCardsWithWeights(chunks, "exact-query-match", 2, 0, MemoryCardRankingWeights{
		QueryExact:     1,
		MemoryTypeBase: 40,
	}, nil)
	if len(cards) != 2 {
		t.Fatalf("expected two cards, got %#v", cards)
	}
	if cards[0].MemoryType != "acceptance" {
		t.Fatalf("expected configured memory type weight to dominate, got %#v", cards)
	}
}

func TestCardsSearchReturnsRankedFilteredCards(t *testing.T) {
	svc := newResolverServiceForTest(t)
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	_, _, err := svc.DocsIngest(ctx, DocsIngestRequest{
		DocType:     "spec",
		Title:       "Cards Search Spec",
		Version:     "2026-03-10",
		FeatureKeys: []string{"cards-search"},
		Tags:        []string{"memory"},
		Body: `# Cards Search Spec

## Acceptance Criteria
- card search returns acceptance memory for resolver cards

## Background
- resolver cards have background context`,
	})
	if err != nil {
		t.Fatalf("DocsIngest: %v", err)
	}

	cards, err := svc.CardsSearch(ctx, CardsSearchRequest{
		Query:       "resolver cards",
		FeatureKeys: []string{"cards-search"},
		MemoryTypes: []string{"acceptance"},
		Limit:       3,
		TokenBudget: 30,
	})
	if err != nil {
		t.Fatalf("CardsSearch: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected one filtered card, got %#v", cards)
	}
	if cards[0].MemoryType != "acceptance" || cards[0].Score <= 0 {
		t.Fatalf("unexpected card search result: %#v", cards[0])
	}
}

func TestMemoryCardFeedbackAdjustsRanking(t *testing.T) {
	svc := newResolverServiceForTest(t)
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	_, _, err := svc.DocsIngest(ctx, DocsIngestRequest{
		DocType: "spec",
		Title:   "Feedback Ranking Spec",
		Version: "2026-03-10",
		Body: `# Feedback Ranking Spec

## Acceptance Criteria
- alpha resolver cards answer

## Risk
- beta resolver cards risk`,
	})
	if err != nil {
		t.Fatalf("DocsIngest: %v", err)
	}

	initial, err := svc.CardsSearch(ctx, CardsSearchRequest{Query: "resolver cards", Limit: 5})
	if err != nil {
		t.Fatalf("CardsSearch initial: %v", err)
	}
	if len(initial) < 2 {
		t.Fatalf("expected multiple cards, got %#v", initial)
	}
	var riskCard ResolverMemoryCard
	for _, card := range initial {
		if card.MemoryType == "risk" {
			riskCard = card
			break
		}
	}
	if riskCard.CardID == "" {
		t.Fatalf("expected risk card in results: %#v", initial)
	}
	if _, err := svc.CardFeedback(ctx, CardFeedbackRequest{
		CardID:     riskCard.CardID,
		DocID:      riskCard.DocID,
		ChunkID:    riskCard.ChunkID,
		MemoryType: riskCard.MemoryType,
		Signal:     "helpful",
		Weight:     20,
		Query:      "resolver cards",
	}); err != nil {
		t.Fatalf("CardFeedback: %v", err)
	}

	adjusted, err := svc.CardsSearch(ctx, CardsSearchRequest{Query: "resolver cards", Limit: 5})
	if err != nil {
		t.Fatalf("CardsSearch adjusted: %v", err)
	}
	if adjusted[0].CardID != riskCard.CardID {
		t.Fatalf("expected feedback to boost risk card, got %#v", adjusted)
	}
}

func TestPromptBundleAndTaskStateExport(t *testing.T) {
	svc := newResolverServiceForTest(t)
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	doc, _, err := svc.DocsIngest(ctx, DocsIngestRequest{
		DocType:      "spec",
		Title:        "Prompt Export Spec",
		Version:      "2026-03-10",
		FeatureKeys:  []string{"prompt-export"},
		TaskIDs:      []string{"task:prompt"},
		TrackerRefs:  []string{"tracker:issue:github:owner/repo#123"},
		BirdseyeRefs: []string{"docs.requirements"},
		Body: `# Prompt Export Spec

## Acceptance Criteria
- prompt bundle exports cards`,
	})
	if err != nil {
		t.Fatalf("DocsIngest: %v", err)
	}
	if _, err := svc.ReadsAck(ctx, ReadsAckRequest{TaskID: "task:prompt", DocID: doc.DocID}); err != nil {
		t.Fatalf("ReadsAck: %v", err)
	}

	bundle, err := svc.PromptBundle(ctx, PromptBundleRequest{Query: "prompt bundle", Feature: "prompt-export", Limit: 3, TokenBudget: 60})
	if err != nil {
		t.Fatalf("PromptBundle: %v", err)
	}
	if bundle.Prompt == "" || len(bundle.Cards) == 0 || len(bundle.SourceRefs) == 0 {
		t.Fatalf("unexpected prompt bundle: %#v", bundle)
	}

	exported, err := svc.TaskStateExport(ctx, TaskStateExportRequest{TaskID: "task:prompt", Feature: "prompt-export"})
	if err != nil {
		t.Fatalf("TaskStateExport: %v", err)
	}
	if exported.TaskRef != "agent-taskstate:task:local:task_prompt" {
		t.Fatalf("unexpected task ref: %#v", exported)
	}
	if len(exported.RequiredDocs) != 1 || len(exported.ReadReceipts) != 1 || len(exported.SourceRefs) == 0 {
		t.Fatalf("unexpected taskstate export: %#v", exported)
	}
	if !containsFold(exported.SourceRefs, "tracker:issue:github:owner/repo#123") {
		t.Fatalf("expected tracker source ref in export: %#v", exported.SourceRefs)
	}
	if !containsFold(exported.SourceRefs, "birdseye:view:local:docs.requirements") {
		t.Fatalf("expected Birdseye source ref in export: %#v", exported.SourceRefs)
	}
}

func TestDocsIngestRejectsOlderVersion(t *testing.T) {
	svc := newResolverServiceForTest(t)
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	doc, _, err := svc.DocsIngest(ctx, DocsIngestRequest{DocType: "spec", Title: "Versioned Spec", Version: "2026-03-11", Body: "# Spec"})
	if err != nil {
		t.Fatalf("initial DocsIngest: %v", err)
	}

	_, _, err = svc.DocsIngest(ctx, DocsIngestRequest{DocID: doc.DocID, DocType: "spec", Title: "Versioned Spec", Version: "2026-03-10", Body: "# Older"})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestDocsIngestUsesSemverVersionScheme(t *testing.T) {
	svc := newResolverServiceForTest(t)
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	doc, _, err := svc.DocsIngest(ctx, DocsIngestRequest{
		DocType:       "spec",
		Title:         "Semver Spec",
		Version:       "v1.0.0",
		VersionScheme: "semver",
		TaskIDs:       []string{"task:semver"},
		Body:          "# Spec",
	})
	if err != nil {
		t.Fatalf("DocsIngest initial: %v", err)
	}
	if doc.VersionScheme != "semver" {
		t.Fatalf("expected semver scheme, got %#v", doc)
	}
	if _, err := svc.ReadsAck(ctx, ReadsAckRequest{TaskID: "task:semver", DocID: doc.DocID}); err != nil {
		t.Fatalf("ReadsAck: %v", err)
	}
	if _, _, err := svc.DocsIngest(ctx, DocsIngestRequest{
		DocID:         doc.DocID,
		DocType:       "spec",
		Title:         "Semver Spec",
		Version:       "1.0.0",
		VersionScheme: "semver",
		TaskIDs:       []string{"task:semver"},
		Body:          "# Spec",
	}); err != nil {
		t.Fatalf("DocsIngest equivalent semver: %v", err)
	}
	stale, err := svc.DocsStaleCheck(ctx, DocsStaleCheckRequest{TaskID: "task:semver"})
	if err != nil {
		t.Fatalf("DocsStaleCheck: %v", err)
	}
	if len(stale) != 0 {
		t.Fatalf("expected equivalent semver to stay fresh, got %#v", stale)
	}

	_, _, err = svc.DocsIngest(ctx, DocsIngestRequest{
		DocID:         doc.DocID,
		DocType:       "spec",
		Title:         "Semver Spec",
		Version:       "0.9.0",
		VersionScheme: "semver",
		Body:          "# Older",
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected semver ErrConflict, got %v", err)
	}
}

func TestDocsStaleCheckUsesLatestReceiptPerDoc(t *testing.T) {
	svc := newResolverServiceForTest(t)
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	doc, _, err := svc.DocsIngest(ctx, DocsIngestRequest{DocType: "spec", Title: "Latest Receipt Spec", Version: "2026-03-10", TaskIDs: []string{"task:latest"}, Body: "# Spec"})
	if err != nil {
		t.Fatalf("DocsIngest: %v", err)
	}
	if _, err := svc.ReadsAck(ctx, ReadsAckRequest{TaskID: "task:latest", DocID: doc.DocID, Version: "2026-03-09"}); err != nil {
		t.Fatalf("ReadsAck old: %v", err)
	}
	if _, err := svc.ReadsAck(ctx, ReadsAckRequest{TaskID: "task:latest", DocID: doc.DocID, Version: "2026-03-10"}); err != nil {
		t.Fatalf("ReadsAck latest: %v", err)
	}

	stale, err := svc.DocsStaleCheck(ctx, DocsStaleCheckRequest{TaskID: "task:latest"})
	if err != nil {
		t.Fatalf("DocsStaleCheck: %v", err)
	}
	if len(stale) != 0 {
		t.Fatalf("expected latest receipt to clear stale, got %#v", stale)
	}
}

func TestDocsIngestFixedChunking(t *testing.T) {
	svc := newResolverServiceForTest(t)
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	doc, chunkCount, err := svc.DocsIngest(ctx, DocsIngestRequest{
		DocType: "spec",
		Title:   "Fixed Chunk Spec",
		Version: "2026-03-10",
		Body:    "0123456789ABCDEFGHIJ0123456789ABCDEFGHIJ",
		Chunking: ChunkingOptions{
			Mode:     "fixed",
			MaxChars: 10,
		},
	})
	if err != nil {
		t.Fatalf("DocsIngest fixed: %v", err)
	}
	if chunkCount < 4 {
		t.Fatalf("expected fixed chunks, got %d", chunkCount)
	}
	_, chunks, err := svc.ChunksGet(ctx, ChunksGetRequest{DocID: doc.DocID})
	if err != nil {
		t.Fatalf("ChunksGet: %v", err)
	}
	if chunks[0].Heading != "Fixed Chunk Spec" {
		t.Fatalf("expected fixed chunk heading to stay on title, got %#v", chunks[0])
	}
	if len(chunks[0].Body) > 10 {
		t.Fatalf("expected fixed chunk body length <= 10, got %d", len(chunks[0].Body))
	}
}

func TestResolverDocsCanUseSeparateResolverStore(t *testing.T) {
	tmpDir := t.TempDir()
	svc, err := New(db.Paths{
		Short:     filepath.Join(tmpDir, "short.db"),
		Journal:   filepath.Join(tmpDir, "journal.db"),
		Knowledge: filepath.Join(tmpDir, "knowledge.db"),
		Archive:   filepath.Join(tmpDir, "archive.db"),
		Resolver:  filepath.Join(tmpDir, "resolver.db"),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = svc.Close() }()
	ctx := context.Background()

	doc, chunkCount, err := svc.DocsIngest(ctx, DocsIngestRequest{
		DocType:     "spec",
		Title:       "Separated Resolver Spec",
		Version:     "2026-03-10",
		FeatureKeys: []string{"resolver-split"},
		Body: `# Separated Resolver Spec

## Acceptance Criteria
- resolver store can be separated`,
	})
	if err != nil {
		t.Fatalf("DocsIngest: %v", err)
	}
	if chunkCount == 0 {
		t.Fatal("expected chunks to be generated")
	}

	var shortCount int
	if err := svc.Conn.ShortDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='resolver_documents';").Scan(&shortCount); err != nil {
		t.Fatalf("check short resolver table: %v", err)
	}
	if shortCount != 0 {
		t.Fatalf("expected short.db to not own resolver tables, got count=%d", shortCount)
	}

	required, _, err := svc.DocsResolve(ctx, DocsResolveRequest{Feature: "resolver-split"})
	if err != nil {
		t.Fatalf("DocsResolve: %v", err)
	}
	if len(required) != 1 || required[0].DocID != doc.DocID {
		t.Fatalf("unexpected required docs: %#v", required)
	}
}
