package api

import (
	"testing"
)

func TestTypedRef_String(t *testing.T) {
	tests := []struct {
		ref      TypedRef
		expected string
	}{
		{TypedRef{Type: EntityTypeEvidence, ID: "01HXXXXXXX"}, "memx:evidence:01HXXXXXXX"},
		{TypedRef{Type: EntityTypeKnowledge, ID: "01HYYYYYYY"}, "memx:knowledge:01HYYYYYYY"},
		{TypedRef{Type: EntityTypeArtifact, ID: "01HZZZZZZZ"}, "memx:artifact:01HZZZZZZZ"},
		{TypedRef{Type: EntityTypeLineage, ID: "01HAAAAAAA"}, "memx:lineage:01HAAAAAAA"},
		{TypedRef{Type: EntityTypeEvidenceChunk, ID: "01HCHUNK"}, "memx:evidence_chunk:01HCHUNK"},
	}

	for _, tt := range tests {
		if got := tt.ref.String(); got != tt.expected {
			t.Errorf("TypedRef.String() = %s, want %s", got, tt.expected)
		}
	}
}

func TestParseTypedRef(t *testing.T) {
	tests := []struct {
		input    string
		expected TypedRef
		hasError bool
	}{
		{"memx:evidence:01HXXXXXXX", TypedRef{Type: EntityTypeEvidence, ID: "01HXXXXXXX"}, false},
		{"memx:knowledge:01HYYYYYYY", TypedRef{Type: EntityTypeKnowledge, ID: "01HYYYYYYY"}, false},
		{"memx:artifact:01HZZZZZZZ", TypedRef{Type: EntityTypeArtifact, ID: "01HZZZZZZZ"}, false},
		{"memx:lineage:01HAAAAAAA", TypedRef{Type: EntityTypeLineage, ID: "01HAAAAAAA"}, false},
		{"", TypedRef{}, true},                          // empty
		{"invalid", TypedRef{}, true},                   // wrong format
		{"foo:evidence:01HXXXXXXX", TypedRef{}, true},   // wrong prefix
		{"memx:unknown:01HXXXXXXX", TypedRef{}, true},   // unknown type
		{"memx:evidence:", TypedRef{}, true},            // empty id
	}

	for _, tt := range tests {
		got, err := ParseTypedRef(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("ParseTypedRef(%s) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseTypedRef(%s) unexpected error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("ParseTypedRef(%s) = %v, want %v", tt.input, got, tt.expected)
			}
		}
	}
}

func TestTypedRef_RoundTrip(t *testing.T) {
	refs := []TypedRef{
		{Type: EntityTypeEvidence, ID: "01HXXXXXXX"},
		{Type: EntityTypeKnowledge, ID: "01HYYYYYYY"},
		{Type: EntityTypeArtifact, ID: "01HZZZZZZZ"},
		{Type: EntityTypeLineage, ID: "01HAAAAAAA"},
	}

	for _, original := range refs {
		s := original.String()
		parsed, err := ParseTypedRef(s)
		if err != nil {
			t.Errorf("ParseTypedRef(%s) error: %v", s, err)
			continue
		}
		if parsed != original {
			t.Errorf("RoundTrip: got %v, want %v", parsed, original)
		}
	}
}

func TestNewTypedRef(t *testing.T) {
	ref := NewTypedRef(EntityTypeEvidence, "01HTEST")
	if ref.Type != EntityTypeEvidence {
		t.Errorf("NewTypedRef type = %s, want %s", ref.Type, EntityTypeEvidence)
	}
	if ref.ID != "01HTEST" {
		t.Errorf("NewTypedRef id = %s, want 01HTEST", ref.ID)
	}
}

func TestTypedRef_IsZero_IsValid(t *testing.T) {
	zero := TypedRef{}
	if !zero.IsZero() {
		t.Error("empty TypedRef should be zero")
	}
	if zero.IsValid() {
		t.Error("empty TypedRef should not be valid")
	}

	valid := TypedRef{Type: EntityTypeEvidence, ID: "01HTEST"}
	if valid.IsZero() {
		t.Error("non-empty TypedRef should not be zero")
	}
	if !valid.IsValid() {
		t.Error("non-empty TypedRef should be valid")
	}

	noID := TypedRef{Type: EntityTypeEvidence, ID: ""}
	if noID.IsValid() {
		t.Error("TypedRef without ID should not be valid")
	}

	noType := TypedRef{Type: "", ID: "01HTEST"}
	if noType.IsValid() {
		t.Error("TypedRef without Type should not be valid")
	}
}

func TestTypedRef_MarshalUnmarshal(t *testing.T) {
	original := TypedRef{Type: EntityTypeEvidence, ID: "01HTEST"}

	data, err := original.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if string(data) != "memx:evidence:01HTEST" {
		t.Errorf("MarshalText = %s, want memx:evidence:01HTEST", string(data))
	}

	var parsed TypedRef
	if err := parsed.UnmarshalText(data); err != nil {
		t.Fatalf("UnmarshalText error: %v", err)
	}
	if parsed != original {
		t.Errorf("UnmarshalText = %v, want %v", parsed, original)
	}
}