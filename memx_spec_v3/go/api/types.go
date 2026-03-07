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
	ID             string `json:"id"`
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	Body           string `json:"body"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	LastAccessedAt string `json:"last_accessed_at"`
	AccessCount    int64  `json:"access_count"`
	SourceType     string `json:"source_type"`
	Origin         string `json:"origin"`
	SourceTrust    string `json:"source_trust"`
	Sensitivity    string `json:"sensitivity"`
}

// Note は short ストアのノート（API返却モデル）。
type Note NoteBase

// ScopedNote は working_scope を持つノート（chronicle/memopedia）。
type ScopedNote struct {
	NoteBase
	WorkingScope string `json:"working_scope"`
	IsPinned     bool   `json:"is_pinned"`
}

// ChronicleNote は chronicle ストアのノート。
type ChronicleNote ScopedNote

// MemopediaNote は memopedia ストアのノート。
type MemopediaNote ScopedNote

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

// -------------------- Chronicle Store --------------------

type ChronicleIngestRequest struct {
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

type ChronicleIngestResponse struct {
	Note ChronicleNote `json:"note"`
}

type ChronicleSearchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k,omitempty"`
}

type ChronicleSearchResponse struct {
	Notes []ChronicleNote `json:"notes"`
}

type ChronicleListByScopeRequest struct {
	WorkingScope string `json:"working_scope"`
	Limit        int    `json:"limit,omitempty"`
}

type ChronicleListByScopeResponse struct {
	Notes []ChronicleNote `json:"notes"`
}

// -------------------- Memopedia Store --------------------

type MemopediaIngestRequest struct {
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

type MemopediaIngestResponse struct {
	Note MemopediaNote `json:"note"`
}

type MemopediaSearchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k,omitempty"`
}

type MemopediaSearchResponse struct {
	Notes []MemopediaNote `json:"notes"`
}

type MemopediaListByScopeRequest struct {
	WorkingScope string `json:"working_scope"`
	Limit        int    `json:"limit,omitempty"`
}

type MemopediaListByScopeResponse struct {
	Notes []MemopediaNote `json:"notes"`
}

type MemopediaListPinnedRequest struct {
	WorkingScope string `json:"working_scope,omitempty"`
	Limit        int    `json:"limit,omitempty"`
}

type MemopediaListPinnedResponse struct {
	Notes []MemopediaNote `json:"notes"`
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