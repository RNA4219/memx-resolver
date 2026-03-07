package api

import (
	"context"
	"errors"
	"testing"
)

func TestShortNoteResolver_ResolveRef(t *testing.T) {
	tests := []struct {
		name       string
		ref        TypedRef
		mockNote   *Note
		mockErr    error
		wantStatus RefStatus
		wantErr    bool
	}{
		{
			name: "resolved evidence ref",
			ref:  NewTypedRef(EntityTypeEvidence, "test-id-1"),
			mockNote: &Note{
				ID:       "test-id-1",
				Title:    "Test Evidence",
				Summary:  "Test summary",
				Body:     "Test body",
			},
			wantStatus: RefStatusResolved,
		},
		{
			name: "resolved knowledge ref",
			ref:  NewTypedRef(EntityTypeKnowledge, "test-id-2"),
			mockNote: &Note{
				ID:       "test-id-2",
				Title:    "Test Knowledge",
				Summary:  "Knowledge summary",
			},
			wantStatus: RefStatusResolved,
		},
		{
			name:       "unresolved ref",
			ref:        NewTypedRef(EntityTypeEvidence, "not-found"),
			mockErr:    errors.New("not found"),
			wantStatus: RefStatusUnresolved,
		},
		{
			name:    "unsupported domain",
			ref:     NewTypedRefWithProvider(DomainWorkx, "task", ProviderLocal, "task-1"),
			wantErr: true,
		},
		{
			name:    "unsupported type",
			ref:     TypedRef{Domain: DomainMemx, Type: "unknown", Provider: ProviderLocal, ID: "id"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchFunc := func(ctx context.Context, query string, topK int) ([]Note, error) {
				return nil, nil
			}
			showFunc := func(ctx context.Context, id string) (*Note, error) {
				if tt.mockErr != nil {
					return nil, tt.mockErr
				}
				if tt.mockNote != nil && tt.mockNote.ID == id {
					return tt.mockNote, nil
				}
				return nil, errors.New("not found")
			}

			resolver := NewShortNoteResolver(searchFunc, showFunc)
			got, err := resolver.ResolveRef(context.Background(), tt.ref)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got.Status != tt.wantStatus {
				t.Errorf("status = %s, want %s", got.Status, tt.wantStatus)
			}

			if tt.wantStatus == RefStatusResolved {
				if got.Summary == "" {
					t.Error("expected summary, got empty")
				}
			}
		})
	}
}

func TestShortNoteResolver_ResolveMany(t *testing.T) {
	notes := map[string]*Note{
		"id-1": {ID: "id-1", Title: "Note 1", Summary: "Summary 1"},
		"id-2": {ID: "id-2", Title: "Note 2", Summary: "Summary 2"},
	}

	searchFunc := func(ctx context.Context, query string, topK int) ([]Note, error) {
		return nil, nil
	}
	showFunc := func(ctx context.Context, id string) (*Note, error) {
		if note, ok := notes[id]; ok {
			return note, nil
		}
		return nil, errors.New("not found")
	}

	resolver := NewShortNoteResolver(searchFunc, showFunc)

	refs := []TypedRef{
		NewTypedRef(EntityTypeEvidence, "id-1"),
		NewTypedRef(EntityTypeKnowledge, "id-2"),
		NewTypedRef(EntityTypeEvidence, "not-found"),
		NewTypedRefWithProvider(DomainWorkx, "task", ProviderLocal, "task-1"),
	}

	report, err := resolver.ResolveMany(context.Background(), refs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Resolved) != 2 {
		t.Errorf("resolved count = %d, want 2", len(report.Resolved))
	}

	if len(report.Unresolved) != 1 {
		t.Errorf("unresolved count = %d, want 1", len(report.Unresolved))
	}

	if len(report.Unsupported) != 1 {
		t.Errorf("unsupported count = %d, want 1", len(report.Unsupported))
	}

	if !report.Diagnostics.PartialBundle {
		t.Error("expected partial bundle to be true")
	}
}

func TestShortNoteResolver_LoadSummary(t *testing.T) {
	notes := map[string]*Note{
		"test-id": {ID: "test-id", Title: "Test", Summary: "Test summary"},
	}

	searchFunc := func(ctx context.Context, query string, topK int) ([]Note, error) {
		return nil, nil
	}
	showFunc := func(ctx context.Context, id string) (*Note, error) {
		if note, ok := notes[id]; ok {
			return note, nil
		}
		return nil, errors.New("not found")
	}

	resolver := NewShortNoteResolver(searchFunc, showFunc)

	t.Run("existing ref", func(t *testing.T) {
		ref := NewTypedRef(EntityTypeEvidence, "test-id")
		payload, err := resolver.LoadSummary(context.Background(), ref)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !payload.Exists {
			t.Error("expected exists to be true")
		}
		if payload.Summary != "Test summary" {
			t.Errorf("summary = %s, want 'Test summary'", payload.Summary)
		}
	})

	t.Run("non-existing ref", func(t *testing.T) {
		ref := NewTypedRef(EntityTypeEvidence, "not-found")
		payload, err := resolver.LoadSummary(context.Background(), ref)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if payload.Exists {
			t.Error("expected exists to be false")
		}
	})
}

func TestShortNoteResolver_LoadSelectedRaw(t *testing.T) {
	notes := map[string]*Note{
		"test-id": {ID: "test-id", Title: "Test", Summary: "Summary", Body: "Full body content"},
	}

	searchFunc := func(ctx context.Context, query string, topK int) ([]Note, error) {
		return nil, nil
	}
	showFunc := func(ctx context.Context, id string) (*Note, error) {
		if note, ok := notes[id]; ok {
			return note, nil
		}
		return nil, errors.New("not found")
	}

	resolver := NewShortNoteResolver(searchFunc, showFunc)

	t.Run("with body", func(t *testing.T) {
		ref := NewTypedRef(EntityTypeEvidence, "test-id")
		selector := RawSelector{IncludeBody: true}
		payload, err := resolver.LoadSelectedRaw(context.Background(), ref, selector)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !payload.Found {
			t.Error("expected found to be true")
		}
		if payload.Raw != "Full body content" {
			t.Errorf("raw = %s, want 'Full body content'", payload.Raw)
		}
	})

	t.Run("without body", func(t *testing.T) {
		ref := NewTypedRef(EntityTypeEvidence, "test-id")
		selector := RawSelector{IncludeBody: false}
		payload, err := resolver.LoadSelectedRaw(context.Background(), ref, selector)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !payload.Found {
			t.Error("expected found to be true")
		}
		if payload.Raw != "Summary" {
			t.Errorf("raw = %s, want 'Summary'", payload.Raw)
		}
	})
}

func TestValidateTypedRefForResolve(t *testing.T) {
	tests := []struct {
		name    string
		ref     TypedRef
		wantErr bool
	}{
		{
			name:    "valid memx local ref",
			ref:     NewTypedRef(EntityTypeEvidence, "test-id"),
			wantErr: false,
		},
		{
			name:    "invalid empty ref",
			ref:     TypedRef{},
			wantErr: true,
		},
		{
			name:    "unsupported workx domain",
			ref:     NewTypedRefWithProvider(DomainWorkx, "task", ProviderLocal, "task-1"),
			wantErr: true,
		},
		{
			name:    "unsupported non-local provider",
			ref:     NewTypedRefWithProvider(DomainMemx, EntityTypeEvidence, ProviderJira, "id"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTypedRefForResolve(tt.ref)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}