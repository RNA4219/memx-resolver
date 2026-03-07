package api

import (
	"fmt"
	"strings"
)

// EntityType は memx 内のエンティティ種別。
type EntityType string

const (
	EntityTypeEvidence      EntityType = "evidence"
	EntityTypeEvidenceChunk EntityType = "evidence_chunk"
	EntityTypeKnowledge     EntityType = "knowledge"
	EntityTypeArtifact      EntityType = "artifact"
	EntityTypeLineage       EntityType = "lineage"
)

// TypedRef は memx 内のエンティティへの型付き参照。
// 形式: memx:<entity_type>:<id>
// 例: memx:evidence:01HXXXXXXX
type TypedRef struct {
	Type EntityType `json:"type"`
	ID   string     `json:"id"`
}

// String は TypedRef を memx:<type>:<id> 形式の文字列に変換する。
func (r TypedRef) String() string {
	return fmt.Sprintf("memx:%s:%s", r.Type, r.ID)
}

// MarshalText は encoding.TextMarshaler インターフェースを実装する。
func (r TypedRef) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

// UnmarshalText は encoding.TextUnmarshaler インターフェースを実装する。
func (r *TypedRef) UnmarshalText(text []byte) error {
	ref, err := ParseTypedRef(string(text))
	if err != nil {
		return err
	}
	*r = ref
	return nil
}

// ParseTypedRef は memx:<type>:<id> 形式の文字列を TypedRef に変換する。
func ParseTypedRef(s string) (TypedRef, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return TypedRef{}, fmt.Errorf("empty ref")
	}

	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return TypedRef{}, fmt.Errorf("invalid ref format: %s (expected memx:<type>:<id>)", s)
	}

	if parts[0] != "memx" {
		return TypedRef{}, fmt.Errorf("invalid ref prefix: %s (expected 'memx')", parts[0])
	}

	entityType := EntityType(parts[1])
	switch entityType {
	case EntityTypeEvidence, EntityTypeEvidenceChunk, EntityTypeKnowledge, EntityTypeArtifact, EntityTypeLineage:
		// valid
	default:
		return TypedRef{}, fmt.Errorf("invalid entity type: %s", entityType)
	}

	id := parts[2]
	if id == "" {
		return TypedRef{}, fmt.Errorf("empty id in ref: %s", s)
	}

	return TypedRef{Type: entityType, ID: id}, nil
}

// MustParseTypedRef は ParseTypedRef のパニック版。
func MustParseTypedRef(s string) TypedRef {
	ref, err := ParseTypedRef(s)
	if err != nil {
		panic(err)
	}
	return ref
}

// NewTypedRef は指定された型とIDから TypedRef を作成する。
func NewTypedRef(entityType EntityType, id string) TypedRef {
	return TypedRef{Type: entityType, ID: id}
}

// Ref は TypedRef.String() のエイリアス。
func (r TypedRef) Ref() string {
	return r.String()
}

// IsZero は TypedRef がゼロ値かどうかを返す。
func (r TypedRef) IsZero() bool {
	return r.Type == "" && r.ID == ""
}

// IsValid は TypedRef が有効かどうかを返す。
func (r TypedRef) IsValid() bool {
	return r.Type != "" && r.ID != ""
}