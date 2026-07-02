package service

import "regexp"

// ResolverDocument represents a document in the resolver store.
type ResolverDocument struct {
	DocID              string   `json:"doc_id"`
	DocType            string   `json:"doc_type"`
	Title              string   `json:"title"`
	SourcePath         string   `json:"source_path"`
	Version            string   `json:"version"`
	VersionScheme      string   `json:"version_scheme,omitempty"`
	UpdatedAt          string   `json:"updated_at"`
	Summary            string   `json:"summary"`
	Body               string   `json:"body,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	FeatureKeys        []string `json:"feature_keys,omitempty"`
	TaskIDs            []string `json:"task_ids,omitempty"`
	TrackerRefs        []string `json:"tracker_refs,omitempty"`
	BirdseyeRefs       []string `json:"birdseye_refs,omitempty"`
	AcceptanceCriteria []string `json:"acceptance_criteria,omitempty"`
	ForbiddenPatterns  []string `json:"forbidden_patterns,omitempty"`
	DefinitionOfDone   []string `json:"definition_of_done,omitempty"`
	Dependencies       []string `json:"dependencies,omitempty"`
	Importance         string   `json:"importance"`
}

// ResolverChunk represents a chunk of a document.
type ResolverChunk struct {
	ChunkID       string   `json:"chunk_id"`
	DocID         string   `json:"doc_id"`
	Heading       string   `json:"heading,omitempty"`
	HeadingPath   []string `json:"heading_path"`
	Ordinal       int      `json:"ordinal"`
	Body          string   `json:"body"`
	TokenEstimate int      `json:"token_estimate"`
	Importance    string   `json:"importance"`
	MemoryType    string   `json:"memory_type"`
	Cue           string   `json:"cue"`
}

// ResolverMemoryCard is a prompt-ready memory unit derived from a chunk.
type ResolverMemoryCard struct {
	CardID        string   `json:"card_id"`
	DocID         string   `json:"doc_id"`
	ChunkID       string   `json:"chunk_id"`
	MemoryType    string   `json:"memory_type"`
	Cue           string   `json:"cue"`
	Statement     string   `json:"statement"`
	HeadingPath   []string `json:"heading_path"`
	Importance    string   `json:"importance"`
	TokenEstimate int      `json:"token_estimate"`
	Score         int      `json:"score"`
}

// MemoryCardRankingWeights controls prompt-ready memory card scoring.
type MemoryCardRankingWeights struct {
	ImportanceRequired    int `json:"importance_required,omitempty"`
	ImportanceRecommended int `json:"importance_recommended,omitempty"`
	ImportanceReference   int `json:"importance_reference,omitempty"`
	MemoryTypeBase        int `json:"memory_type_base,omitempty"`
	QueryExact            int `json:"query_exact,omitempty"`
	QueryTerms            int `json:"query_terms,omitempty"`
	CueMatch              int `json:"cue_match,omitempty"`
	HeadingMatch          int `json:"heading_match,omitempty"`
	ShortCardBonus        int `json:"short_card_bonus,omitempty"`
	FeedbackBoost         int `json:"feedback_boost,omitempty"`
}

// resolverSection is an internal representation of a document section.
type resolverSection struct {
	Heading     string
	HeadingPath []string
	Body        string
	Importance  string
}

// ResolveEntry represents a resolved document entry.
type ResolveEntry struct {
	DocID      string   `json:"doc_id"`
	Title      string   `json:"title"`
	Version    string   `json:"version"`
	Importance string   `json:"importance"`
	Reason     string   `json:"reason"`
	TopChunks  []string `json:"top_chunks"`
}

// ResolverReadReceipt represents a read receipt for a document.
type ResolverReadReceipt struct {
	TaskID              string                  `json:"task_id"`
	DocID               string                  `json:"doc_id"`
	Version             string                  `json:"version"`
	ChunkIDs            []string                `json:"chunk_ids,omitempty"`
	ChunkSnapshots      []ResolverChunkSnapshot `json:"chunk_snapshots,omitempty"`
	PreviousReceiptHash string                  `json:"previous_receipt_hash,omitempty"`
	ReceiptHash         string                  `json:"receipt_hash,omitempty"`
	Reader              string                  `json:"reader"`
	ReadAt              string                  `json:"read_at"`
}

// ResolverStaleReason represents a stale check result.
type ResolverStaleReason struct {
	TaskID          string                `json:"task_id"`
	DocID           string                `json:"doc_id"`
	PreviousVersion string                `json:"previous_version"`
	CurrentVersion  string                `json:"current_version"`
	Reason          string                `json:"reason"`
	Severity        string                `json:"severity,omitempty"`
	ImpactScope     []string              `json:"impact_scope,omitempty"`
	ChangedChunks   []ResolverChunkChange `json:"changed_chunks,omitempty"`
	DetectedAt      string                `json:"detected_at"`
}

// ResolverChunkSnapshot captures chunk state at read time for semantic stale checks.
type ResolverChunkSnapshot struct {
	ChunkID       string   `json:"chunk_id"`
	BodyHash      string   `json:"body_hash"`
	HeadingPath   []string `json:"heading_path,omitempty"`
	MemoryType    string   `json:"memory_type,omitempty"`
	Importance    string   `json:"importance,omitempty"`
	TokenEstimate int      `json:"token_estimate,omitempty"`
}

// ResolverChunkChange describes the semantic change affecting a read chunk.
type ResolverChunkChange struct {
	ChunkID     string   `json:"chunk_id"`
	ChangeType  string   `json:"change_type"`
	HeadingPath []string `json:"heading_path,omitempty"`
	MemoryType  string   `json:"memory_type,omitempty"`
	Impact      string   `json:"impact"`
}

// ChunkingOptions controls how documents are chunked.
type ChunkingOptions struct {
	Mode     string `json:"mode,omitempty"`
	MaxChars int    `json:"max_chars,omitempty"`
}

// DocsIngestRequest is the request for ingesting a document.
type DocsIngestRequest struct {
	DocID              string          `json:"doc_id,omitempty"`
	DocType            string          `json:"doc_type"`
	Title              string          `json:"title"`
	SourcePath         string          `json:"source_path,omitempty"`
	Version            string          `json:"version"`
	VersionScheme      string          `json:"version_scheme,omitempty"`
	UpdatedAt          string          `json:"updated_at,omitempty"`
	Tags               []string        `json:"tags,omitempty"`
	FeatureKeys        []string        `json:"feature_keys,omitempty"`
	TaskIDs            []string        `json:"task_ids,omitempty"`
	TrackerRefs        []string        `json:"tracker_refs,omitempty"`
	BirdseyeRefs       []string        `json:"birdseye_refs,omitempty"`
	Summary            string          `json:"summary,omitempty"`
	Body               string          `json:"body"`
	Chunking           ChunkingOptions `json:"chunking,omitempty"`
	AcceptanceCriteria []string        `json:"acceptance_criteria,omitempty"`
	ForbiddenPatterns  []string        `json:"forbidden_patterns,omitempty"`
	DefinitionOfDone   []string        `json:"definition_of_done,omitempty"`
	Dependencies       []string        `json:"dependencies,omitempty"`
}

// DocsResolveRequest is the request for resolving documents.
type DocsResolveRequest struct {
	Feature string `json:"feature,omitempty"`
	TaskID  string `json:"task_id,omitempty"`
	Topic   string `json:"topic,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// ChunksGetRequest is the request for getting chunks.
type ChunksGetRequest struct {
	DocID    string   `json:"doc_id,omitempty"`
	Query    string   `json:"query,omitempty"`
	Heading  string   `json:"heading,omitempty"`
	Limit    int      `json:"limit,omitempty"`
	ChunkIDs []string `json:"chunk_ids,omitempty"`
}

// DocsSearchRequest is the request for searching documents.
type DocsSearchRequest struct {
	Query       string   `json:"query"`
	DocTypes    []string `json:"doc_types,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	FeatureKeys []string `json:"feature_keys,omitempty"`
	Limit       int      `json:"limit,omitempty"`
}

// CardsSearchRequest is the request for searching prompt-ready memory cards.
type CardsSearchRequest struct {
	Query          string                   `json:"query"`
	DocTypes       []string                 `json:"doc_types,omitempty"`
	Tags           []string                 `json:"tags,omitempty"`
	FeatureKeys    []string                 `json:"feature_keys,omitempty"`
	MemoryTypes    []string                 `json:"memory_types,omitempty"`
	Limit          int                      `json:"limit,omitempty"`
	TokenBudget    int                      `json:"token_budget,omitempty"`
	RankingWeights MemoryCardRankingWeights `json:"ranking_weights,omitempty"`
}

// ReadsAckRequest is the request for acknowledging a read.
type ReadsAckRequest struct {
	TaskID   string   `json:"task_id"`
	DocID    string   `json:"doc_id"`
	Version  string   `json:"version,omitempty"`
	ChunkIDs []string `json:"chunk_ids,omitempty"`
	Reader   string   `json:"reader,omitempty"`
}

// DocsStaleCheckRequest is the request for checking stale documents.
type DocsStaleCheckRequest struct {
	TaskID string `json:"task_id"`
}

// CardFeedbackRequest records actual card usage for adaptive ranking.
type CardFeedbackRequest struct {
	CardID     string `json:"card_id"`
	DocID      string `json:"doc_id,omitempty"`
	ChunkID    string `json:"chunk_id,omitempty"`
	MemoryType string `json:"memory_type,omitempty"`
	Signal     string `json:"signal"`
	Weight     int    `json:"weight,omitempty"`
	Query      string `json:"query,omitempty"`
}

// CardFeedbackRecord is a persisted memory card usage signal.
type CardFeedbackRecord struct {
	CardID     string `json:"card_id"`
	DocID      string `json:"doc_id,omitempty"`
	ChunkID    string `json:"chunk_id,omitempty"`
	MemoryType string `json:"memory_type,omitempty"`
	Signal     string `json:"signal"`
	Weight     int    `json:"weight"`
	Query      string `json:"query,omitempty"`
	RecordedAt string `json:"recorded_at"`
}

// PromptBundleRequest exports memory cards in prompt-ready form.
type PromptBundleRequest struct {
	Query       string   `json:"query"`
	Feature     string   `json:"feature,omitempty"`
	TaskID      string   `json:"task_id,omitempty"`
	MemoryTypes []string `json:"memory_types,omitempty"`
	Limit       int      `json:"limit,omitempty"`
	TokenBudget int      `json:"token_budget,omitempty"`
	Format      string   `json:"format,omitempty"`
}

// PromptBundleResponse is a prompt-ready card bundle.
type PromptBundleResponse struct {
	BundleID      string               `json:"bundle_id"`
	Query         string               `json:"query"`
	Format        string               `json:"format"`
	TokenEstimate int                  `json:"token_estimate"`
	Cards         []ResolverMemoryCard `json:"cards"`
	Prompt        string               `json:"prompt"`
	SourceRefs    []string             `json:"source_refs"`
}

// TaskStateExportRequest exports resolver state for agent-taskstate.
type TaskStateExportRequest struct {
	TaskID  string `json:"task_id"`
	Feature string `json:"feature,omitempty"`
}

// TaskStateExportResponse carries resolver state in agent-taskstate-friendly shape.
type TaskStateExportResponse struct {
	TaskRef      string                `json:"task_ref"`
	TaskID       string                `json:"task_id"`
	RequiredDocs []ResolveEntry        `json:"required_docs"`
	ReadReceipts []ResolverReadReceipt `json:"read_receipts"`
	StaleReasons []ResolverStaleReason `json:"stale_reasons"`
	SourceRefs   []string              `json:"source_refs"`
	ExportedAt   string                `json:"exported_at"`
}

// ContractsResolveRequest is the request for resolving contracts.
type ContractsResolveRequest struct {
	Feature string `json:"feature,omitempty"`
	TaskID  string `json:"task_id,omitempty"`
}

// scoredResolverDoc is an internal representation for scoring.
type scoredResolverDoc struct {
	Doc   ResolverDocument
	Score int
	Why   string
}

// markdownHeadingPattern matches markdown headings.
var markdownHeadingPattern = regexp.MustCompile(`^(#{1,6})\s+(.*\S)\s*$`)
