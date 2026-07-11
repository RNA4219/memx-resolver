package api

import "github.com/RNA4219/memx-resolver/v2/service"

type ResolverDocument = service.ResolverDocument

type ResolverChunk = service.ResolverChunk

type ResolverMemoryCard = service.ResolverMemoryCard

type MemoryCardRankingWeights = service.MemoryCardRankingWeights

type ResolveEntry = service.ResolveEntry

type ResolverReadReceipt = service.ResolverReadReceipt

type ResolverStaleReason = service.ResolverStaleReason

type ChunkingOptions = service.ChunkingOptions

type DocsIngestRequest = service.DocsIngestRequest

type DocsResolveRequest = service.DocsResolveRequest

type ChunksGetRequest = service.ChunksGetRequest

type DocsSearchRequest = service.DocsSearchRequest

type CardsSearchRequest = service.CardsSearchRequest

type ReadsAckRequest = service.ReadsAckRequest

type DocsStaleCheckRequest = service.DocsStaleCheckRequest

type CardFeedbackRequest = service.CardFeedbackRequest

type CardFeedbackRecord = service.CardFeedbackRecord

type PromptBundleRequest = service.PromptBundleRequest

type PromptBundleResponse = service.PromptBundleResponse

type TaskStateExportRequest = service.TaskStateExportRequest

type TaskStateExportResponse = service.TaskStateExportResponse

type ContractsResolveRequest = service.ContractsResolveRequest

type DocsIngestResponse struct {
	DocID      string `json:"doc_id"`
	Version    string `json:"version"`
	ChunkCount int    `json:"chunk_count"`
	Status     string `json:"status"`
}

type DocsResolveResponse struct {
	Required    []ResolveEntry `json:"required"`
	Recommended []ResolveEntry `json:"recommended"`
}

type ChunksGetResponse struct {
	DocID       string               `json:"doc_id"`
	Chunks      []ResolverChunk      `json:"chunks"`
	MemoryCards []ResolverMemoryCard `json:"memory_cards,omitempty"`
}

type DocsSearchResponse struct {
	Results []ResolveEntry `json:"results"`
}

type CardsSearchResponse struct {
	Cards []ResolverMemoryCard `json:"cards"`
}

type ReadsAckResponse struct {
	Receipt ResolverReadReceipt `json:"receipt"`
}

type DocsStaleCheckResponse struct {
	TaskID       string                `json:"task_id"`
	Status       string                `json:"status"`
	StaleReasons []ResolverStaleReason `json:"stale_reasons"`
	Stale        []ResolverStaleReason `json:"stale"`
}

type CardFeedbackResponse struct {
	Feedback CardFeedbackRecord `json:"feedback"`
}

type ContractsResolveResponse struct {
	Required           []ResolveEntry `json:"required"`
	AcceptanceCriteria []string       `json:"acceptance_criteria"`
	ForbiddenPatterns  []string       `json:"forbidden_patterns"`
	DefinitionOfDone   []string       `json:"definition_of_done"`
	Dependencies       []string       `json:"dependencies"`
}
