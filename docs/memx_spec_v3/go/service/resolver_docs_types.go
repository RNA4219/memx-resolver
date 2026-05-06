package service

import "regexp"

// ResolverDocument represents a document in the resolver store.
type ResolverDocument struct {
	DocID              string   `json:"doc_id"`
	DocType            string   `json:"doc_type"`
	Title              string   `json:"title"`
	SourcePath         string   `json:"source_path"`
	Version            string   `json:"version"`
	UpdatedAt          string   `json:"updated_at"`
	Summary            string   `json:"summary"`
	Body               string   `json:"body,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	FeatureKeys        []string `json:"feature_keys,omitempty"`
	TaskIDs            []string `json:"task_ids,omitempty"`
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
	TaskID   string   `json:"task_id"`
	DocID    string   `json:"doc_id"`
	Version  string   `json:"version"`
	ChunkIDs []string `json:"chunk_ids,omitempty"`
	Reader   string   `json:"reader"`
	ReadAt   string   `json:"read_at"`
}

// ResolverStaleReason represents a stale check result.
type ResolverStaleReason struct {
	TaskID          string `json:"task_id"`
	DocID           string `json:"doc_id"`
	PreviousVersion string `json:"previous_version"`
	CurrentVersion  string `json:"current_version"`
	Reason          string `json:"reason"`
	DetectedAt      string `json:"detected_at"`
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
	UpdatedAt          string          `json:"updated_at,omitempty"`
	Tags               []string        `json:"tags,omitempty"`
	FeatureKeys        []string        `json:"feature_keys,omitempty"`
	TaskIDs            []string        `json:"task_ids,omitempty"`
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
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
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