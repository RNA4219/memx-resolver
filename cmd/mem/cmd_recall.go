package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/RNA4219/memx-resolver/v2/api"
)

func cmdOutRecall(args []string) {
	fs := flag.NewFlagSet("mem out recall", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	topK := fs.Int("k", 10, "top k results")
	msgRange := fs.Int("range", 5, "message range for context")
	stores := fs.String("stores", "short", "comma-separated stores to search (short,journal,knowledge)")
	fallbackFTS := fs.Bool("fallback-fts", false, "use FTS fallback if embedding unavailable")
	_ = fs.Parse(args)
	query := strings.Join(fs.Args(), " ")

	if query == "" {
		log.Fatal("QUERY is required")
	}

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	storeList := parseRecallStores(*stores)

	resp, apiErr := client.Recall(ctx, api.RecallRequest{
		Query:        query,
		TopK:         *topK,
		MessageRange: *msgRange,
		Stores:       storeList,
		FallbackFTS:  *fallbackFTS,
	})
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}

	if cf.json {
		printJSON(resp)
		return
	}

	for _, nwc := range resp.Results {
		anchor := nwc.Anchor
		fmt.Printf("=== [%s] %s (score: %.4f) ===\n", anchor.Store, anchor.Title, anchor.Score)
		if anchor.Summary != "" {
			fmt.Printf("Summary: %s\n", anchor.Summary)
		}
		fmt.Printf("ID: %s\n", anchor.ID)
		if len(nwc.Before) > 0 {
			fmt.Printf("\n--- Before (%d) ---\n", len(nwc.Before))
			for _, b := range nwc.Before {
				fmt.Printf("  [%s] %s\n", b.ID, b.Title)
			}
		}
		if len(nwc.After) > 0 {
			fmt.Printf("\n--- After (%d) ---\n", len(nwc.After))
			for _, a := range nwc.After {
				fmt.Printf("  [%s] %s\n", a.ID, a.Title)
			}
		}
		fmt.Println()
	}
}

func parseRecallStores(s string) []string {
	if s == "" {
		return []string{"short"}
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if p == "short" || p == "journal" || p == "knowledge" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return []string{"short"}
	}
	return result
}
