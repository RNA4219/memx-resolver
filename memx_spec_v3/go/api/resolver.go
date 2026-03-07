package api

import (
	"context"
	"fmt"
)

// Resolver は typed_ref からエンティティを解決するインターフェース。
// P4 Phase 3B: current memx-core adapter 用の最小実装。
type Resolver interface {
	// ResolveRef は単一の typed_ref を解決する。
	ResolveRef(ctx context.Context, ref TypedRef) (ResolvedRef, error)

	// ResolveMany は複数の typed_ref を一括解決する。
	ResolveMany(ctx context.Context, refs []TypedRef) (*ResolveReport, error)

	// LoadSummary は要約を取得する（summary-first retrieval）。
	LoadSummary(ctx context.Context, ref TypedRef) (*SummaryPayload, error)

	// LoadSelectedRaw は必要時のみ raw データを取得する。
	LoadSelectedRaw(ctx context.Context, ref TypedRef, selector RawSelector) (*RawPayload, error)
}

// ResolvedRef は解決結果を表す。
type ResolvedRef struct {
	Ref      TypedRef   `json:"ref"`
	Status   RefStatus  `json:"status"`
	Summary  string     `json:"summary,omitempty"`
	Metadata RefMetadata `json:"metadata,omitempty"`
}

// RefStatus は解決状態を表す。
type RefStatus string

const (
	RefStatusResolved   RefStatus = "resolved"
	RefStatusUnresolved RefStatus = "unresolved"
	RefStatusUnsupported RefStatus = "unsupported"
)

// RefMetadata は参照先のメタデータ。
type RefMetadata struct {
	Title       string `json:"title,omitempty"`
	SourceType  string `json:"source_type,omitempty"`
	Sensitivity string `json:"sensitivity,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// ResolveReport は一括解決結果を表す。
type ResolveReport struct {
	Resolved   []ResolvedRef `json:"resolved"`
	Unresolved []TypedRef    `json:"unresolved"`
	Unsupported []TypedRef   `json:"unsupported"`
	Diagnostics ResolverDiagnostics `json:"diagnostics"`
}

// ResolverDiagnostics は解決診断情報。
type ResolverDiagnostics struct {
	MissingRefs    []TypedRef `json:"missing_refs"`
	UnsupportedRefs []TypedRef `json:"unsupported_refs"`
	ResolverWarnings []string `json:"resolver_warnings"`
	PartialBundle   bool      `json:"partial_bundle"`
}

// SummaryPayload は要約取得結果。
type SummaryPayload struct {
	Ref     TypedRef `json:"ref"`
	Summary string   `json:"summary"`
	Exists  bool     `json:"exists"`
}

// RawSelector は raw データ取得時のセレクタ。
type RawSelector struct {
	IncludeBody   bool     `json:"include_body"`
	IncludeTags   bool     `json:"include_tags"`
	Fields        []string `json:"fields,omitempty"`
}

// RawPayload は raw データ取得結果。
type RawPayload struct {
	Ref   TypedRef `json:"ref"`
	Raw   string   `json:"raw,omitempty"`
	Found bool     `json:"found"`
}

// ErrUnsupportedRef は未対応の typed_ref が指定された場合のエラー。
type ErrUnsupportedRef struct {
	Ref TypedRef
}

func (e *ErrUnsupportedRef) Error() string {
	return fmt.Sprintf("unsupported typed_ref: %s", e.Ref)
}

// ErrUnresolvedRef は解決できない typed_ref が指定された場合のエラー。
type ErrUnresolvedRef struct {
	Ref TypedRef
	Reason string
}

func (e *ErrUnresolvedRef) Error() string {
	return fmt.Sprintf("unresolved typed_ref: %s (%s)", e.Ref, e.Reason)
}