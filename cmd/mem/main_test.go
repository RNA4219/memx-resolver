package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/RNA4219/memx-resolver/v2/db"
	"github.com/RNA4219/memx-resolver/v2/service"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	return repoRoot(t)
}

func buildMemBinary(t *testing.T) string {
	t.Helper()
	binName := "mem"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(t.TempDir(), binName)
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/mem")
	cmd.Dir = moduleRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build mem binary: %v\n%s", err, out)
	}
	return binPath
}

func runMem(t *testing.T, binPath, workdir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Dir = workdir
	cmd.Env = append(os.Environ(),
		"OPENAI_API_KEY=",
		"MEMX_OPENAI_API_KEY=",
		"DASHSCOPE_API_KEY=",
		"MEMX_ALIBABA_API_KEY=",
		"MEMX_DASHSCOPE_API_KEY=",
		"MEMX_LLM_PROVIDER=",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run %q: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func ingestIDFromJSON(t *testing.T, out string) string {
	t.Helper()
	var resp struct {
		Note struct {
			ID string `json:"id"`
		} `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("decode ingest response: %v\n%s", err, out)
	}
	if resp.Note.ID == "" {
		t.Fatalf("missing note id in response: %s", out)
	}
	return resp.Note.ID
}

func noteIDsFromSearchJSON(t *testing.T, out string) map[string]struct{} {
	t.Helper()
	var resp struct {
		Notes []struct {
			ID string `json:"id"`
		} `json:"notes"`
	}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("decode search response: %v\n%s", err, out)
	}
	ids := make(map[string]struct{}, len(resp.Notes))
	for _, note := range resp.Notes {
		ids[note.ID] = struct{}{}
	}
	return ids
}

func TestSkillDocsDescribeWorkingCommands(t *testing.T) {
	root := repoRoot(t)

	recallDoc, err := os.ReadFile(filepath.Join(root, ".claude", "commands", "recall.md"))
	if err != nil {
		t.Fatalf("read recall.md: %v", err)
	}
	if !strings.Contains(string(recallDoc), `go run ./cmd/mem out search --json "<query>"`) {
		t.Fatalf("recall skill should document the supported --json command order\n%s", recallDoc)
	}

	showDoc, err := os.ReadFile(filepath.Join(root, ".claude", "commands", "show.md"))
	if err != nil {
		t.Fatalf("read show.md: %v", err)
	}
	if !strings.Contains(string(showDoc), "Works for notes from any store") {
		t.Fatalf("show skill should describe cross-store resolution\n%s", showDoc)
	}

	helpDoc, err := os.ReadFile(filepath.Join(root, ".claude", "commands", "memx-help.md"))
	if err != nil {
		t.Fatalf("read memx-help.md: %v", err)
	}
	if !strings.Contains(string(helpDoc), `go run ./cmd/mem out search --json "query"`) {
		t.Fatalf("memx-help should show the supported search command\n%s", helpDoc)
	}
}

func TestSkillCommandsUseDefaultStorePaths(t *testing.T) {
	binPath := buildMemBinary(t)
	workdir := t.TempDir()

	journalID := ingestIDFromJSON(t, runMem(t, binPath, workdir,
		"in", "journal", "--json",
		"--title", "Skill Journal",
		"--body", "default path journal entry",
		"--scope", "project:memx",
		"--sensitivity", "internal",
	))
	knowledgeID := ingestIDFromJSON(t, runMem(t, binPath, workdir,
		"in", "knowledge", "--json",
		"--title", "Skill Knowledge",
		"--body", "default path knowledge entry",
		"--scope", "glossary",
		"--pinned",
		"--sensitivity", "internal",
	))

	for _, name := range []string{"short.db", "journal.db", "knowledge.db", "archive.db"} {
		if _, err := os.Stat(filepath.Join(workdir, name)); err != nil {
			t.Fatalf("expected %s to exist: %v", name, err)
		}
	}

	journalOut := runMem(t, binPath, workdir, "out", "journal", "show", journalID)
	if !strings.Contains(journalOut, "default path journal entry") {
		t.Fatalf("journal show output missing entry: %s", journalOut)
	}

	knowledgeOut := runMem(t, binPath, workdir, "out", "knowledge", "show", knowledgeID)
	if !strings.Contains(knowledgeOut, "default path knowledge entry") {
		t.Fatalf("knowledge show output missing entry: %s", knowledgeOut)
	}
}

func TestRecallSkillSearchesAcrossStores(t *testing.T) {
	binPath := buildMemBinary(t)
	workdir := t.TempDir()
	marker := "skillcheck-20260308-shared"

	shortID := ingestIDFromJSON(t, runMem(t, binPath, workdir,
		"in", "short", "--json",
		"--title", "Short Recall",
		"--body", "short store "+marker,
		"--sensitivity", "internal",
	))
	journalID := ingestIDFromJSON(t, runMem(t, binPath, workdir,
		"in", "journal", "--json",
		"--title", "Journal Recall",
		"--body", "journal store "+marker,
		"--scope", "project:memx",
		"--sensitivity", "internal",
	))
	knowledgeID := ingestIDFromJSON(t, runMem(t, binPath, workdir,
		"in", "knowledge", "--json",
		"--title", "Knowledge Recall",
		"--body", "knowledge store "+marker,
		"--scope", "glossary",
		"--sensitivity", "internal",
	))

	ids := noteIDsFromSearchJSON(t, runMem(t, binPath, workdir, "out", "search", "--json", marker))
	for _, id := range []string{shortID, journalID, knowledgeID} {
		if _, ok := ids[id]; !ok {
			t.Fatalf("expected %s in cross-store recall results: %#v", id, ids)
		}
	}
}

func TestShowSkillFindsNotesAcrossStoresIncludingArchive(t *testing.T) {
	binPath := buildMemBinary(t)
	workdir := t.TempDir()

	archiveCandidateID := ingestIDFromJSON(t, runMem(t, binPath, workdir,
		"in", "short", "--json",
		"--title", "Archive Candidate",
		"--body", "archive store body",
		"--sensitivity", "internal",
	))
	journalID := ingestIDFromJSON(t, runMem(t, binPath, workdir,
		"in", "journal", "--json",
		"--title", "Journal Show",
		"--body", "journal show body",
		"--scope", "project:memx",
		"--sensitivity", "internal",
	))
	knowledgeID := ingestIDFromJSON(t, runMem(t, binPath, workdir,
		"in", "knowledge", "--json",
		"--title", "Knowledge Show",
		"--body", "knowledge show body",
		"--scope", "glossary",
		"--pinned",
		"--sensitivity", "internal",
	))

	svc, err := service.New(db.Paths{
		Short:     filepath.Join(workdir, "short.db"),
		Journal:   filepath.Join(workdir, "journal.db"),
		Knowledge: filepath.Join(workdir, "knowledge.db"),
		Archive:   filepath.Join(workdir, "archive.db"),
	})
	if err != nil {
		t.Fatalf("service.New: %v", err)
	}
	defer func() { _ = svc.Close() }()
	if _, err := svc.ArchiveNoteFromShort(context.Background(), archiveCandidateID); err != nil {
		t.Fatalf("ArchiveNoteFromShort: %v", err)
	}

	journalOut := runMem(t, binPath, workdir, "out", "show", journalID)
	if !strings.Contains(journalOut, "journal show body") || !strings.Contains(journalOut, "Scope: project:memx") {
		t.Fatalf("journal note should resolve via top-level show: %s", journalOut)
	}

	knowledgeOut := runMem(t, binPath, workdir, "out", "show", knowledgeID)
	if !strings.Contains(knowledgeOut, "knowledge show body") || !strings.Contains(knowledgeOut, "Scope: glossary") {
		t.Fatalf("knowledge note should resolve via top-level show: %s", knowledgeOut)
	}

	archiveOut := runMem(t, binPath, workdir, "out", "show", archiveCandidateID)
	if !strings.Contains(archiveOut, "archive store body") {
		t.Fatalf("archive note should resolve via top-level show: %s", archiveOut)
	}
}

func TestIngestCommandsAcceptNoLLMFlag(t *testing.T) {
	binPath := buildMemBinary(t)
	workdir := t.TempDir()

	for _, args := range [][]string{
		{"in", "short", "--json", "--no-llm", "--title", "No LLM Short", "--body", "body", "--sensitivity", "internal"},
		{"in", "journal", "--json", "--no-llm", "--title", "No LLM Journal", "--body", "body", "--scope", "project:memx", "--sensitivity", "internal"},
		{"in", "knowledge", "--json", "--no-llm", "--title", "No LLM Knowledge", "--body", "body", "--scope", "glossary", "--sensitivity", "internal"},
	} {
		out := runMem(t, binPath, workdir, args...)
		if ingestIDFromJSON(t, out) == "" {
			t.Fatalf("missing note id for args %v", args)
		}
	}
}

func TestDocsCommandsCanUseSeparateResolverStore(t *testing.T) {
	binPath := buildMemBinary(t)
	workdir := t.TempDir()
	resolverPath := filepath.Join(workdir, "resolver.db")

	out := runMem(t, binPath, workdir,
		"docs", "ingest", "--json",
		"--resolver", resolverPath,
		"--title", "CLI Resolver Split",
		"--body", "# CLI Resolver Split\n\n## Acceptance Criteria\n- separated resolver store works",
		"--doc-type", "spec",
		"--version", "2026-03-10",
		"--feature", "resolver-cli-split",
	)
	var ingestResp struct {
		DocID      string `json:"doc_id"`
		ChunkCount int    `json:"chunk_count"`
	}
	if err := json.Unmarshal([]byte(out), &ingestResp); err != nil {
		t.Fatalf("decode docs ingest response: %v\n%s", err, out)
	}
	if ingestResp.DocID == "" || ingestResp.ChunkCount == 0 {
		t.Fatalf("unexpected docs ingest response: %#v", ingestResp)
	}

	svc, err := service.New(db.Paths{
		Short:     filepath.Join(workdir, "short.db"),
		Journal:   filepath.Join(workdir, "journal.db"),
		Knowledge: filepath.Join(workdir, "knowledge.db"),
		Archive:   filepath.Join(workdir, "archive.db"),
		Resolver:  resolverPath,
	})
	if err != nil {
		t.Fatalf("service.New: %v", err)
	}
	defer func() { _ = svc.Close() }()

	var shortCount int
	if err := svc.Conn.ShortDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='resolver_documents';").Scan(&shortCount); err != nil {
		t.Fatalf("check short resolver table: %v", err)
	}
	if shortCount != 0 {
		t.Fatalf("expected short.db to not own resolver tables, got count=%d", shortCount)
	}

	required, _, err := svc.DocsResolve(context.Background(), service.DocsResolveRequest{Feature: "resolver-cli-split"})
	if err != nil {
		t.Fatalf("DocsResolve: %v", err)
	}
	if len(required) != 1 || required[0].DocID != ingestResp.DocID {
		t.Fatalf("unexpected required docs: %#v", required)
	}

	resolveOut := runMem(t, binPath, workdir,
		"docs", "resolve", "--json",
		"--resolver", resolverPath,
		"--feature", "resolver-cli-split",
	)
	var resolveResp struct {
		Required []struct {
			TopChunks []string `json:"top_chunks"`
		} `json:"required"`
	}
	if err := json.Unmarshal([]byte(resolveOut), &resolveResp); err != nil {
		t.Fatalf("decode docs resolve response: %v\n%s", err, resolveOut)
	}
	if len(resolveResp.Required) == 0 || len(resolveResp.Required[0].TopChunks) == 0 {
		t.Fatalf("expected resolve response to include top chunks: %#v", resolveResp)
	}

	searchOut := runMem(t, binPath, workdir,
		"docs", "search", "--json",
		"--resolver", resolverPath,
		"--feature", "resolver-cli-split",
		"--doc-type", "spec",
		"separated resolver",
	)
	var searchResp struct {
		Results []struct {
			DocID string `json:"doc_id"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(searchOut), &searchResp); err != nil {
		t.Fatalf("decode docs search response: %v\n%s", err, searchOut)
	}
	if len(searchResp.Results) != 1 || searchResp.Results[0].DocID != ingestResp.DocID {
		t.Fatalf("unexpected docs search response: %#v", searchResp)
	}

	chunksOut := runMem(t, binPath, workdir,
		"docs", "chunks", "--json",
		"--resolver", resolverPath,
		"--chunk-id", resolveResp.Required[0].TopChunks[0],
	)
	var chunksResp struct {
		Chunks []struct {
			ChunkID string `json:"chunk_id"`
		} `json:"chunks"`
	}
	if err := json.Unmarshal([]byte(chunksOut), &chunksResp); err != nil {
		t.Fatalf("decode docs chunks response: %v\n%s", err, chunksOut)
	}
	if len(chunksResp.Chunks) != 1 || chunksResp.Chunks[0].ChunkID != resolveResp.Required[0].TopChunks[0] {
		t.Fatalf("unexpected chunks response: %#v", chunksResp)
	}

	cardsOut := runMem(t, binPath, workdir,
		"docs", "cards", "--json",
		"--resolver", resolverPath,
		"--query", "separated resolver",
		"--feature", "resolver-cli-split",
		"--memory-type", "acceptance",
		"--token-budget", "40",
	)
	var cardsResp struct {
		Cards []struct {
			CardID     string `json:"card_id"`
			DocID      string `json:"doc_id"`
			ChunkID    string `json:"chunk_id"`
			MemoryType string `json:"memory_type"`
			Score      int    `json:"score"`
		} `json:"cards"`
	}
	if err := json.Unmarshal([]byte(cardsOut), &cardsResp); err != nil {
		t.Fatalf("decode docs cards response: %v\n%s", err, cardsOut)
	}
	if len(cardsResp.Cards) != 1 || cardsResp.Cards[0].MemoryType != "acceptance" || cardsResp.Cards[0].Score <= 0 {
		t.Fatalf("unexpected docs cards response: %#v", cardsResp)
	}

	feedbackOut := runMem(t, binPath, workdir,
		"docs", "cards-feedback", "--json",
		"--resolver", resolverPath,
		"--card-id", cardsResp.Cards[0].CardID,
		"--doc-id", cardsResp.Cards[0].DocID,
		"--chunk-id", cardsResp.Cards[0].ChunkID,
		"--memory-type", cardsResp.Cards[0].MemoryType,
		"--signal", "helpful",
		"--query", "separated resolver",
	)
	if !strings.Contains(feedbackOut, `"signal": "helpful"`) {
		t.Fatalf("unexpected feedback output: %s", feedbackOut)
	}

	bundleOut := runMem(t, binPath, workdir,
		"docs", "bundle", "--json",
		"--resolver", resolverPath,
		"--query", "separated resolver",
		"--feature", "resolver-cli-split",
		"--token-budget", "80",
	)
	if !strings.Contains(bundleOut, `"prompt"`) || !strings.Contains(bundleOut, `"source_refs"`) {
		t.Fatalf("unexpected bundle output: %s", bundleOut)
	}

	ackOut := runMem(t, binPath, workdir,
		"docs", "ack", "--json",
		"--resolver", resolverPath,
		"--task-id", "task:cli:split",
		"--doc-id", ingestResp.DocID,
	)
	if !strings.Contains(ackOut, `"chunk_snapshots"`) {
		t.Fatalf("unexpected ack output: %s", ackOut)
	}

	exportOut := runMem(t, binPath, workdir,
		"docs", "taskstate-export", "--json",
		"--resolver", resolverPath,
		"--task-id", "task:cli:split",
		"--feature", "resolver-cli-split",
	)
	if !strings.Contains(exportOut, `"task_ref": "agent-taskstate:task:local:task_cli_split"`) || !strings.Contains(exportOut, `"source_refs"`) {
		t.Fatalf("unexpected taskstate export output: %s", exportOut)
	}
}

func TestDocsMigrateResolverStoreCommand(t *testing.T) {
	binPath := buildMemBinary(t)
	workdir := t.TempDir()
	shortPath := filepath.Join(workdir, "short.db")
	resolverPath := filepath.Join(workdir, "resolver.db")

	out := runMem(t, binPath, workdir,
		"docs", "ingest", "--json",
		"--title", "Migration Source Spec",
		"--body", "# Migration Source Spec\n\n## Acceptance Criteria\n- migrated resolver store resolves docs",
		"--doc-type", "spec",
		"--version", "2026-03-10",
		"--feature", "resolver-migration",
	)
	var ingestResp struct {
		DocID string `json:"doc_id"`
	}
	if err := json.Unmarshal([]byte(out), &ingestResp); err != nil {
		t.Fatalf("decode docs ingest response: %v\n%s", err, out)
	}
	if ingestResp.DocID == "" {
		t.Fatalf("missing doc id: %s", out)
	}

	dryRunOut := runMem(t, binPath, workdir,
		"docs", "migrate-resolver-store", "--json", "--dry-run",
		"--from", shortPath,
		"--to", resolverPath,
	)
	if !strings.Contains(dryRunOut, `"dry_run": true`) || !strings.Contains(dryRunOut, `"resolver_documents": 1`) {
		t.Fatalf("unexpected dry-run output: %s", dryRunOut)
	}

	migrateOut := runMem(t, binPath, workdir,
		"docs", "migrate-resolver-store", "--json",
		"--from", shortPath,
		"--to", resolverPath,
	)
	if !strings.Contains(migrateOut, `"status": "migrated"`) || !strings.Contains(migrateOut, `"resolver_chunks"`) {
		t.Fatalf("unexpected migrate output: %s", migrateOut)
	}

	resolveOut := runMem(t, binPath, workdir,
		"docs", "resolve", "--json",
		"--resolver", resolverPath,
		"--feature", "resolver-migration",
	)
	if !strings.Contains(resolveOut, ingestResp.DocID) {
		t.Fatalf("expected migrated resolver doc to resolve, got: %s", resolveOut)
	}
}
