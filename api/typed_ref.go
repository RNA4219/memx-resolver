package api

import (
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"
)

// Domain は typed_ref のシステム領域を表す。
type Domain string

const (
	DomainMemx           Domain = "memx"
	DomainAgentTaskstate Domain = "agent-taskstate"
	DomainTracker        Domain = "tracker"
)

// EntityType はエンティティ種別。未知の値も拡張用に受理する。
type EntityType string

const (
	EntityTypeEvidence      EntityType = "evidence"
	EntityTypeEvidenceChunk EntityType = "evidence_chunk"
	EntityTypeKnowledge     EntityType = "knowledge"
	EntityTypeArtifact      EntityType = "artifact"
	EntityTypeLineage       EntityType = "lineage"
)

// Provider はデータソース。未知の値も拡張用に受理する。
type Provider string

const (
	ProviderLocal  Provider = "local"
	ProviderJira   Provider = "jira"
	ProviderGitHub Provider = "github"
	ProviderLinear Provider = "linear"
)

const DefaultProvider Provider = ProviderLocal

// TypedRef はv2の4セグメント型付き参照。
type TypedRef struct {
	Domain   Domain     `json:"domain"`
	Type     EntityType `json:"type"`
	Provider Provider   `json:"provider"`
	ID       string     `json:"id"`
}

// String はdomain/type/providerを小文字化し、IDをRFC 3986形式で符号化する。
func (r TypedRef) String() string {
	return fmt.Sprintf(
		"%s:%s:%s:%s",
		strings.ToLower(string(r.Domain)),
		strings.ToLower(string(r.Type)),
		strings.ToLower(string(r.Provider)),
		percentEncodeID(r.ID),
	)
}

func (r TypedRef) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *TypedRef) UnmarshalText(text []byte) error {
	ref, err := ParseTypedRef(string(text))
	if err != nil {
		return err
	}
	*r = ref
	return nil
}

// ParseTypedRef はv2の4セグメント形式だけを受理する。
func ParseTypedRef(s string) (TypedRef, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return TypedRef{}, fmt.Errorf("empty ref")
	}
	parts := strings.Split(s, ":")
	if len(parts) != 4 {
		return TypedRef{}, fmt.Errorf("invalid ref format: %s (v2 requires <domain>:<type>:<provider>:<id>)", s)
	}
	return parseFourSegment(parts)
}

func parseFourSegment(parts []string) (TypedRef, error) {
	domain := Domain(strings.ToLower(parts[0]))
	if err := validateDomain(domain); err != nil {
		return TypedRef{}, err
	}
	entityType := EntityType(strings.ToLower(parts[1]))
	if err := validateEntityType(domain, entityType); err != nil {
		return TypedRef{}, err
	}
	provider := Provider(strings.ToLower(parts[2]))
	if err := validateProvider(provider); err != nil {
		return TypedRef{}, err
	}
	id, err := decodeEntityID(parts[3])
	if err != nil {
		return TypedRef{}, err
	}
	return TypedRef{Domain: domain, Type: entityType, Provider: provider, ID: id}, nil
}

func validateDomain(domain Domain) error {
	switch domain {
	case DomainMemx, DomainAgentTaskstate, DomainTracker:
		return nil
	default:
		return fmt.Errorf("invalid domain: %s (expected memx, agent-taskstate, or tracker)", domain)
	}
}

func validateEntityType(_ Domain, entityType EntityType) error {
	if entityType == "" {
		return fmt.Errorf("empty entity type")
	}
	if strings.Contains(string(entityType), "%") {
		return fmt.Errorf("percent encoding is only allowed in id")
	}
	return nil
}

func validateProvider(provider Provider) error {
	if provider == "" {
		return fmt.Errorf("empty provider")
	}
	if strings.Contains(string(provider), "%") {
		return fmt.Errorf("percent encoding is only allowed in id")
	}
	return nil
}

func decodeEntityID(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("empty id in ref")
	}
	for i := 0; i < len(raw); i++ {
		if raw[i] != '%' {
			continue
		}
		if i+2 >= len(raw) || !isHex(raw[i+1]) || !isHex(raw[i+2]) {
			return "", fmt.Errorf("malformed percent encoding in id")
		}
		i += 2
	}
	decoded, err := url.PathUnescape(raw)
	if err != nil {
		return "", fmt.Errorf("malformed percent encoding in id: %w", err)
	}
	if decoded == "" {
		return "", fmt.Errorf("empty id in ref")
	}
	if !utf8.ValidString(decoded) {
		return "", fmt.Errorf("invalid UTF-8 percent encoding in id")
	}
	return decoded, nil
}

func percentEncodeID(id string) string {
	var b strings.Builder
	const hex = "0123456789ABCDEF"
	for i := 0; i < len(id); i++ {
		c := id[i]
		if isUnreserved(c) {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hex[c>>4])
		b.WriteByte(hex[c&0x0f])
	}
	return b.String()
}

func isUnreserved(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '.' || c == '_' || c == '~'
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'f') ||
		(c >= 'A' && c <= 'F')
}

func MustParseTypedRef(s string) TypedRef {
	ref, err := ParseTypedRef(s)
	if err != nil {
		panic(err)
	}
	return ref
}

func NewTypedRef(entityType EntityType, id string) TypedRef {
	return TypedRef{Domain: DomainMemx, Type: entityType, Provider: DefaultProvider, ID: id}
}

func NewTypedRefWithProvider(domain Domain, entityType EntityType, provider Provider, id string) TypedRef {
	return TypedRef{Domain: domain, Type: entityType, Provider: provider, ID: id}
}

func (r TypedRef) Ref() string {
	return r.String()
}

func (r TypedRef) IsZero() bool {
	return r.Domain == "" && r.Type == "" && r.Provider == "" && r.ID == ""
}

func (r TypedRef) IsValid() bool {
	if validateDomain(Domain(strings.ToLower(string(r.Domain)))) != nil {
		return false
	}
	if validateEntityType(r.Domain, EntityType(strings.ToLower(string(r.Type)))) != nil {
		return false
	}
	if validateProvider(Provider(strings.ToLower(string(r.Provider)))) != nil {
		return false
	}
	_, err := decodeEntityID(r.ID)
	return err == nil
}

func (r TypedRef) Canonical() string {
	return r.String()
}
