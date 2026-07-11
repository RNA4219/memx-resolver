package api

import "testing"

func TestTypedRefStringCanonicalAndEncoding(t *testing.T) {
	tests := []struct {
		ref  TypedRef
		want string
	}{
		{TypedRef{Domain: DomainMemx, Type: EntityTypeEvidence, Provider: ProviderLocal, ID: "01HXXXXXXX"}, "memx:evidence:local:01HXXXXXXX"},
		{TypedRef{Domain: DomainTracker, Type: "Issue", Provider: "JIRA", ID: "A:B / C"}, "tracker:issue:jira:A%3AB%20%2F%20C"},
		{TypedRef{Domain: DomainAgentTaskstate, Type: "Task", Provider: "Custom", ID: "AZaz09-._~"}, "agent-taskstate:task:custom:AZaz09-._~"},
	}
	for _, tt := range tests {
		if got := tt.ref.String(); got != tt.want {
			t.Fatalf("String()=%q, want %q", got, tt.want)
		}
	}
}

func TestParseTypedRefV2(t *testing.T) {
	tests := []struct {
		input string
		want  TypedRef
	}{
		{"memx:evidence:local:01HXXXXXXX", TypedRef{Domain: DomainMemx, Type: EntityTypeEvidence, Provider: ProviderLocal, ID: "01HXXXXXXX"}},
		{"TRACKER:Issue:JIRA:Proj-ABC", TypedRef{Domain: DomainTracker, Type: "issue", Provider: ProviderJira, ID: "Proj-ABC"}},
		{"memx:doc:custom-provider:id", TypedRef{Domain: DomainMemx, Type: "doc", Provider: "custom-provider", ID: "id"}},
		{"tracker:issue:github:owner%2Frepo%23123", TypedRef{Domain: DomainTracker, Type: "issue", Provider: ProviderGitHub, ID: "owner/repo#123"}},
	}
	for _, tt := range tests {
		got, err := ParseTypedRef(tt.input)
		if err != nil {
			t.Fatalf("ParseTypedRef(%q): %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("ParseTypedRef(%q)=%+v, want %+v", tt.input, got, tt.want)
		}
	}
}

func TestParseTypedRefRejectsLegacyAndMalformed(t *testing.T) {
	inputs := []string{
		"",
		"memx:evidence:01HXXXXXXX",
		"memx:evidence:local:",
		"memx:evidence::id",
		"memx:evidence:local:bad%2",
		"memx:evidence:local:bad%GG",
		"memx:evidence:local:%FF",
		"unknown:evidence:local:id",
	}
	for _, input := range inputs {
		if _, err := ParseTypedRef(input); err == nil {
			t.Errorf("ParseTypedRef(%q) expected error", input)
		}
	}
}

func TestTypedRefRoundTrip(t *testing.T) {
	original := TypedRef{Domain: DomainMemx, Type: EntityType("doc"), Provider: ProviderLocal, ID: "A:B / C"}
	parsed, err := ParseTypedRef(original.String())
	if err != nil {
		t.Fatal(err)
	}
	if parsed != original {
		t.Fatalf("round trip=%+v, want %+v", parsed, original)
	}
}

func TestTypedRefHelpersAndValidity(t *testing.T) {
	ref := NewTypedRefWithProvider(DomainTracker, "issue", ProviderGitHub, "gh:123")
	if !ref.IsValid() || ref.IsZero() {
		t.Fatalf("valid ref reported invalid or zero: %+v", ref)
	}
	if !(TypedRef{}).IsZero() {
		t.Fatal("zero ref not zero")
	}
	if (TypedRef{Domain: DomainMemx, Type: EntityTypeEvidence, Provider: ProviderLocal}).IsValid() {
		t.Fatal("ref without id reported valid")
	}
	if got := ref.Canonical(); got != "tracker:issue:github:gh%3A123" {
		t.Fatalf("canonical=%q", got)
	}
}

func TestTypedRefMarshalUnmarshal(t *testing.T) {
	original := TypedRef{Domain: DomainMemx, Type: EntityTypeEvidence, Provider: ProviderLocal, ID: "A:B"}
	data, err := original.MarshalText()
	if err != nil || string(data) != "memx:evidence:local:A%3AB" {
		t.Fatalf("MarshalText=%q err=%v", data, err)
	}
	var parsed TypedRef
	if err := parsed.UnmarshalText(data); err != nil {
		t.Fatal(err)
	}
	if parsed != original {
		t.Fatalf("parsed=%+v want %+v", parsed, original)
	}
}
