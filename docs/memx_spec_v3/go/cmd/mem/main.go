package main

import (
	"fmt"
	"os"
)

func main() {
	runSetup()
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "api":
		cmdAPI(os.Args[2:])
	case "in":
		cmdIn(os.Args[2:])
	case "out":
		cmdOut(os.Args[2:])
	case "docs":
		cmdDocs(os.Args[2:])
	case "gc":
		cmdGC(os.Args[2:])
	case "summarize":
		cmdSummarize(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `mem - memx CLI (v1.4)

Usage:
  mem api serve   [--addr 127.0.0.1:7766] [--short short.db] [--resolver resolver.db] ...
  mem in short    --title TITLE [--body BODY | --stdin] [--tag TAG ...] [--no-llm]
  mem in journal --title TITLE --body BODY --scope SCOPE [--tag TAG ...] [--no-llm]
  mem in knowledge --title TITLE --body BODY --scope SCOPE [--tag TAG ...] [--pinned] [--no-llm]
  mem out search  QUERY
  mem out show    NOTE_ID
  mem out recall  QUERY [--top-k N] [--range N] [--stores short,journal,knowledge] [--fallback-fts]
  mem out journal search QUERY
  mem out journal show ID
  mem out journal list --scope SCOPE
  mem out knowledge search QUERY
  mem out knowledge show ID
  mem out knowledge list --scope SCOPE
  mem out knowledge pinned [--scope SCOPE]
  mem out knowledge pin ID
  mem out knowledge unpin ID
  mem out archive list
  mem out archive show ID
  mem out archive restore ID
  mem docs ingest   --title TITLE --body BODY --doc-type TYPE [--version VER] [--feature KEY ...]
  mem docs resolve  --feature KEY | --task-id ID | --topic QUERY [--limit N]
  mem docs chunks   --doc-id ID | --chunk-id ID [--heading H] [--query Q] [--limit N]
  mem docs search   QUERY [--doc-type TYPE ...] [--tag TAG ...] [--feature KEY ...] [--limit N]
  mem docs cards    --query QUERY [--memory-type TYPE ...] [--token-budget N] [--weight-feedback-boost N]
  mem docs cards-feedback --card-id ID --signal used|helpful|pinned|irrelevant|skipped
  mem docs bundle   --query QUERY [--memory-type TYPE ...] [--token-budget N] [--format markdown|jsonl]
  mem docs taskstate-export --task-id ID [--feature KEY]
  mem docs ack      --task-id ID --doc-id ID [--version VER]
  mem docs stale    --task-id ID
  mem docs contract --feature KEY | --task-id ID
  mem gc short    [--dry-run] [--enable-gc]
  mem summarize   NOTE_ID [--json]
  mem summarize   --ids ID1,ID2,... [--json]

Global (for client-mode):
  --api-url http://127.0.0.1:7766   # if set, CLI calls HTTP API
  --json                           # JSON output

DB flags (in-proc / server):
  --short short.db
  --journal journal.db
  --knowledge knowledge.db
  --archive archive.db
  --resolver resolver.db   # optional, defaults to short.db

GC flags:
  --dry-run      Show planned operations without executing
  --enable-gc    Enable GC execution (disabled by default)

Recall flags:
  --top-k N         Number of top results (default: 10)
  --range N         Context range around anchor (default: 5)
  --stores STORES   Comma-separated stores: short,journal,knowledge
  --fallback-fts    Fallback to FTS if embedding unavailable
`)
}
