package api

// v1.3: API はツール/AI 向けの安定 I/F。

type ErrorCode string

const (
	CodeInvalidArgument ErrorCode = "INVALID_ARGUMENT"
	CodeNotFound        ErrorCode = "NOT_FOUND"
	CodeConflict        ErrorCode = "CONFLICT"
	CodeGatekeepDeny    ErrorCode = "GATEKEEP_DENY"
	CodeInternal        ErrorCode = "INTERNAL"
)

type Error struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// NoteBase はノートの共通フィールド。
type NoteBase struct {
	ID             string   `json:"id"`
	Ref            TypedRef `json:"ref"`
	Title          string   `json:"title"`
	Summary        string   `json:"summary"`
	Body           string   `json:"body"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	LastAccessedAt string   `json:"last_accessed_at"`
	AccessCount    int64    `json:"access_count"`
	SourceType     string   `json:"source_type"`
	Origin         string   `json:"origin"`
	SourceTrust    string   `json:"source_trust"`
	Sensitivity    string   `json:"sensitivity"`
}

// Note は short ストアのノート（API返却モデル）。
type Note NoteBase

// ScopedNote は working_scope を持つノート（journal/knowledge）。
type ScopedNote struct {
	NoteBase
	WorkingScope string `json:"working_scope"`
	IsPinned     bool   `json:"is_pinned"`
}

// JournalNote は journal ストアのノート。
type JournalNote ScopedNote

// KnowledgeNote は knowledge ストアのノート。
type KnowledgeNote ScopedNote

// ArchiveNote は archive ストアのノート。
type ArchiveNote NoteBase

// -------------------- Short Store --------------------

type NotesIngestRequest struct {
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	Summary     string   `json:"summary,omitempty"`
	SourceType  string   `json:"source_type,omitempty"`
	Origin      string   `json:"origin,omitempty"`
	SourceTrust string   `json:"source_trust,omitempty"`
	Sensitivity string   `json:"sensitivity,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type NotesIngestResponse struct {
	Note Note `json:"note"`
}

type NotesSearchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k,omitempty"`
}

type NotesSearchResponse struct {
	Notes []Note `json:"notes"`
}

type GCOptions struct {
	DryRun bool `json:"dry_run,omitempty"`
}

type GCRunRequest struct {
	Target  string    `json:"target"`
	Options GCOptions `json:"options,omitempty"`
}

type GCRunResponse struct {
	Status string `json:"status"`
}

type SummarizeRequest struct {
	ID string `json:"id"`
}

type SummarizeResponse struct {
	Note Note `json:"note"`
}

type SummarizeBatchRequest struct {
	IDs []string `json:"ids"`
}

type SummarizeBatchResponse struct {
	Summary   string `json:"summary"`
	NoteCount int    `json:"note_count"`
}

// -------------------- Journal Store --------------------

type JournalIngestRequest struct {
	Title        string   `json:"title"`
	Body         string   `json:"body"`
	Summary      string   `json:"summary,omitempty"`
	SourceType   string   `json:"source_type,omitempty"`
	Origin       string   `json:"origin,omitempty"`
	SourceTrust  string   `json:"source_trust,omitempty"`
	Sensitivity  string   `json:"sensitivity,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	WorkingScope string   `json:"working_scope"`
	IsPinned     bool     `json:"is_pinned,omitempty"`
}

type JournalIngestResponse struct {
	Note JournalNote `json:"note"`
}

type JournalSearchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k,omitempty"`
}

type JournalSearchResponse struct {
	Notes []JournalNote `json:"notes"`
}

type JournalListByScopeRequest struct {
	WorkingScope string `json:"working_scope"`
	Limit        int    `json:"limit,omitempty"`
}

type JournalListByScopeResponse struct {
	Notes []JournalNote `json:"notes"`
}

// -------------------- Knowledge Store --------------------

type KnowledgeIngestRequest struct {
	Title        string   `json:"title"`
	Body         string   `json:"body"`
	Summary      string   `json:"summary,omitempty"`
	SourceType   string   `json:"source_type,omitempty"`
	Origin       string   `json:"origin,omitempty"`
	SourceTrust  string   `json:"source_trust,omitempty"`
	Sensitivity  string   `json:"sensitivity,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	WorkingScope string   `json:"working_scope"`
	IsPinned     bool     `json:"is_pinned,omitempty"`
}

type KnowledgeIngestResponse struct {
	Note KnowledgeNote `json:"note"`
}

type KnowledgeSearchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k,omitempty"`
}

type KnowledgeSearchResponse struct {
	Notes []KnowledgeNote `json:"notes"`
}

type KnowledgeListByScopeRequest struct {
	WorkingScope string `json:"working_scope"`
	Limit        int    `json:"limit,omitempty"`
}

type KnowledgeListByScopeResponse struct {
	Notes []KnowledgeNote `json:"notes"`
}

type KnowledgeListPinnedRequest struct {
	WorkingScope string `json:"working_scope,omitempty"`
	Limit        int    `json:"limit,omitempty"`
}

type KnowledgeListPinnedResponse struct {
	Notes []KnowledgeNote `json:"notes"`
}

type PinResponse struct {
	Success bool `json:"success"`
}

type UnpinResponse struct {
	Success bool `json:"success"`
}

// -------------------- Archive Store --------------------

type ArchiveListRequest struct {
	Limit int `json:"limit,omitempty"`
}

type ArchiveListResponse struct {
	Notes []ArchiveNote `json:"notes"`
}

type ArchiveRestoreResponse struct {
	Note Note `json:"note"`
}