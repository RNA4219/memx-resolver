package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/RNA4219/memx-resolver/v2/api"
)

func cmdSummarize(args []string) {
	fs := flag.NewFlagSet("mem summarize", flag.ExitOnError)
	cf := &commonFlags{}
	cf.bind(fs)
	ids := fs.String("ids", "", "comma-separated note IDs for batch summarization")
	_ = fs.Parse(args)

	ctx := context.Background()
	client, cleanup, err := makeClient(ctx, cf)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	// Batch summarization
	if *ids != "" {
		idList := strings.Split(*ids, ",")
		for i, id := range idList {
			idList[i] = strings.TrimSpace(id)
		}
		resp, apiErr := client.SummarizeBatch(ctx, api.SummarizeBatchRequest{IDs: idList})
		if apiErr != nil {
			log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
		}
		if cf.json {
			printJSON(resp)
			return
		}
		fmt.Printf("Summary (%d notes):\n%s\n", resp.NoteCount, resp.Summary)
		return
	}

	// Single note summarization
	if len(fs.Args()) < 1 {
		log.Fatal("NOTE_ID is required (or use --ids for batch)")
	}
	id := fs.Args()[0]

	resp, apiErr := client.Summarize(ctx, id)
	if apiErr != nil {
		log.Fatalf("%s: %s", apiErr.Code, apiErr.Message)
	}
	if cf.json {
		printJSON(resp)
		return
	}
	fmt.Printf("# %s\n\nSummary: %s\n", resp.Note.Title, resp.Note.Summary)
}
