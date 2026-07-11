"""Tests for typed_ref module."""

from __future__ import annotations

import pytest

from memx_resolver_workflow_plugin.typed_ref import (
    Domain,
    EntityType,
    Provider,
    TypedRef,
    TypedRefParseError,
    parse_typed_ref,
    must_parse_typed_ref,
    new_typed_ref,
    new_typed_ref_with_provider,
    DEFAULT_PROVIDER,
)


class TestParseTypedRef:
    """Test parse_typed_ref function."""

    def test_parse_three_segment_memx_evidence(self) -> None:
        """Parse 3-segment format: memx:evidence:id."""
        ref = parse_typed_ref("memx:evidence:01HXXXXXXX")
        assert ref.domain == Domain.MEMX
        assert ref.type == EntityType.EVIDENCE
        assert ref.provider == DEFAULT_PROVIDER
        assert ref.id == "01HXXXXXXX"

    def test_parse_three_segment_memx_knowledge(self) -> None:
        """Parse 3-segment format: memx:knowledge:id."""
        ref = parse_typed_ref("memx:knowledge:01HYYYYYYY")
        assert ref.domain == Domain.MEMX
        assert ref.type == EntityType.KNOWLEDGE
        assert ref.provider == DEFAULT_PROVIDER
        assert ref.id == "01HYYYYYYY"

    def test_parse_four_segment_canonical(self) -> None:
        """Parse 4-segment canonical format."""
        ref = parse_typed_ref("memx:evidence:local:01HXXXXXXX")
        assert ref.domain == Domain.MEMX
        assert ref.type == EntityType.EVIDENCE
        assert ref.provider == Provider.LOCAL
        assert ref.id == "01HXXXXXXX"

    def test_parse_four_segment_with_jira_provider(self) -> None:
        """Parse 4-segment with jira provider."""
        ref = parse_typed_ref("tracker:issue:jira:PROJ-123")
        assert ref.domain == Domain.TRACKER
        assert ref.type == "issue"  # tracker domain uses string, not EntityType
        assert ref.provider == Provider.JIRA
        assert ref.id == "PROJ-123"

    def test_parse_empty_string_raises_error(self) -> None:
        """Empty string should raise TypedRefParseError."""
        with pytest.raises(TypedRefParseError, match="empty ref"):
            parse_typed_ref("")

    def test_parse_invalid_segment_count_raises_error(self) -> None:
        """Invalid segment count should raise TypedRefParseError."""
        with pytest.raises(TypedRefParseError, match="invalid ref format"):
            parse_typed_ref("memx:evidence")

    def test_parse_invalid_domain_raises_error(self) -> None:
        """Invalid domain should raise TypedRefParseError."""
        with pytest.raises(TypedRefParseError, match="invalid domain"):
            parse_typed_ref("unknown:evidence:local:id")

    def test_parse_extensible_entity_type_for_memx(self) -> None:
        """Memx entity types are extensible across resolver capabilities."""
        ref = parse_typed_ref("memx:doc:custom-provider:id")
        assert ref.entity_type == "doc"
        assert str(ref) == "memx:doc:custom-provider:id"

    def test_parse_empty_id_raises_error(self) -> None:
        """Empty id should raise TypedRefParseError."""
        with pytest.raises(TypedRefParseError, match="empty id"):
            parse_typed_ref("memx:evidence:local:")

    def test_parse_empty_provider_raises_error(self) -> None:
        """Empty provider in 4-segment should raise TypedRefParseError."""
        with pytest.raises(TypedRefParseError, match="empty provider"):
            parse_typed_ref("memx:evidence::id")


class TestTypedRef:
    """Test TypedRef class."""

    def test_str_returns_canonical_format(self) -> None:
        """TypedRef.__str__ returns canonical format."""
        ref = TypedRef(
            domain=Domain.MEMX,
            type=EntityType.EVIDENCE,
            provider=Provider.LOCAL,
            id="01HXXXXXXX",
        )
        assert str(ref) == "memx:evidence:local:01HXXXXXXX"

    def test_ref_alias(self) -> None:
        """TypedRef.ref() is alias for __str__."""
        ref = TypedRef(
            domain=Domain.MEMX,
            type=EntityType.EVIDENCE,
            provider=Provider.LOCAL,
            id="01HXXXXXXX",
        )
        assert ref.ref() == str(ref)

    def test_is_zero(self) -> None:
        """TypedRef.is_zero() returns True for zero value."""
        ref = TypedRef(
            domain=Domain.MEMX,
            type=EntityType.EVIDENCE,
            provider=Provider.LOCAL,
            id="",
        )
        assert not ref.is_zero()  # has domain, type, provider

        zero_ref = TypedRef(
            domain=None,
            type=None,
            provider=None,
            id=None,
        )
        assert zero_ref.is_zero()

    def test_is_valid(self) -> None:
        """TypedRef.is_valid() returns True for valid ref."""
        ref = TypedRef(
            domain=Domain.MEMX,
            type=EntityType.EVIDENCE,
            provider=Provider.LOCAL,
            id="01HXXXXXXX",
        )
        assert ref.is_valid()

        invalid_ref = TypedRef(
            domain=Domain.MEMX,
            type=EntityType.EVIDENCE,
            provider=Provider.LOCAL,
            id="",
        )
        assert not invalid_ref.is_valid()

    def test_frozen(self) -> None:
        """TypedRef is frozen (immutable)."""
        ref = TypedRef(
            domain=Domain.MEMX,
            type=EntityType.EVIDENCE,
            provider=Provider.LOCAL,
            id="01HXXXXXXX",
        )
        with pytest.raises(Exception):
            ref.id = "new-id"  # type: ignore


class TestHelperFunctions:
    """Test helper functions."""

    def test_new_typed_ref(self) -> None:
        """new_typed_ref creates TypedRef with memx domain and local provider."""
        ref = new_typed_ref(EntityType.KNOWLEDGE, "01HZZZZZZZ")
        assert ref.domain == Domain.MEMX
        assert ref.type == EntityType.KNOWLEDGE
        assert ref.provider == Provider.LOCAL
        assert ref.id == "01HZZZZZZZ"

    def test_new_typed_ref_with_provider(self) -> None:
        """new_typed_ref_with_provider creates TypedRef with specified provider."""
        ref = new_typed_ref_with_provider(
            domain=Domain.TRACKER,
            entity_type=EntityType.ARTIFACT,
            provider=Provider.GITHUB,
            id="gh-123",
        )
        assert ref.domain == Domain.TRACKER
        assert ref.type == EntityType.ARTIFACT
        assert ref.provider == Provider.GITHUB
        assert ref.id == "gh-123"

    def test_must_parse_typed_ref(self) -> None:
        """must_parse_typed_ref is same as parse_typed_ref."""
        ref = must_parse_typed_ref("memx:evidence:local:01HXXXXXXX")
        assert ref.domain == Domain.MEMX
        assert ref.type == EntityType.EVIDENCE


class TestNormalization:
    """Test normalization from 3-segment to 4-segment."""

    def test_three_segment_normalizes_to_local_provider(self) -> None:
        """3-segment format normalizes to provider='local'."""
        ref = parse_typed_ref("memx:evidence:01HXXXXXXX")
        assert ref.provider == Provider.LOCAL
        assert str(ref) == "memx:evidence:local:01HXXXXXXX"

    def test_output_is_canonical(self) -> None:
        """Output is always canonical format."""
        ref = parse_typed_ref("memx:knowledge:01HYYYYYYY")
        canonical = ref.canonical()
        assert canonical.count(":") == 3  # 4 segments
        assert canonical.startswith("memx:knowledge:local:")