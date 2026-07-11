"""Tests for strict v2 typed_ref behavior."""

from __future__ import annotations

import pytest

from memx_resolver_workflow_plugin.typed_ref import (
    DEFAULT_PROVIDER,
    Domain,
    EntityType,
    Provider,
    TypedRef,
    TypedRefParseError,
    must_parse_typed_ref,
    new_typed_ref,
    new_typed_ref_with_provider,
    parse_typed_ref,
)


def test_parse_canonical_ref() -> None:
    ref = parse_typed_ref("memx:evidence:local:01HXXXXXXX")
    assert ref.domain == Domain.MEMX
    assert ref.type == EntityType.EVIDENCE
    assert ref.provider == DEFAULT_PROVIDER
    assert ref.id == "01HXXXXXXX"


def test_parse_lowercases_components_but_preserves_id_case() -> None:
    ref = parse_typed_ref("TRACKER:Issue:JIRA:Proj-ABC")
    assert ref.domain == Domain.TRACKER
    assert ref.entity_type == "issue"
    assert ref.provider == Provider.JIRA
    assert ref.id == "Proj-ABC"
    assert str(ref) == "tracker:issue:jira:Proj-ABC"


def test_three_segment_ref_is_rejected() -> None:
    with pytest.raises(TypedRefParseError, match="v2 requires"):
        parse_typed_ref("memx:evidence:01HXXXXXXX")


def test_empty_and_unknown_components_are_rejected() -> None:
    with pytest.raises(TypedRefParseError, match="empty provider"):
        parse_typed_ref("memx:evidence::id")
    with pytest.raises(TypedRefParseError, match="empty id"):
        parse_typed_ref("memx:evidence:local:")
    with pytest.raises(TypedRefParseError, match="invalid domain"):
        parse_typed_ref("unknown:evidence:local:id")


def test_percent_encoding_round_trip_for_colon_and_reserved_chars() -> None:
    ref = TypedRef(Domain.MEMX, EntityType.DOC, Provider.LOCAL, "A:B / C")
    assert str(ref) == "memx:doc:local:A%3AB%20%2F%20C"
    parsed = parse_typed_ref(str(ref))
    assert parsed.id == "A:B / C"
    assert str(parsed) == str(ref)


@pytest.mark.parametrize(
    "value",
    ["memx:doc:local:bad%2", "memx:doc:local:bad%GG", "memx:doc:local:%FF"],
)
def test_malformed_percent_encoding_is_rejected(value: str) -> None:
    with pytest.raises(TypedRefParseError):
        parse_typed_ref(value)


def test_unreserved_id_characters_are_not_encoded() -> None:
    ref = new_typed_ref(EntityType.KNOWLEDGE, "AZaz09-._~")
    assert str(ref) == "memx:knowledge:local:AZaz09-._~"


def test_custom_entity_type_and_provider_are_supported() -> None:
    ref = parse_typed_ref("memx:doc:custom-provider:id")
    assert ref.entity_type == "doc"
    assert ref.provider == "custom-provider"
    assert str(ref) == "memx:doc:custom-provider:id"


def test_invalid_and_zero_values() -> None:
    valid = TypedRef(Domain.MEMX, EntityType.EVIDENCE, Provider.LOCAL, "id")
    assert valid.is_valid()
    invalid = TypedRef(Domain.MEMX, EntityType.EVIDENCE, Provider.LOCAL, "")
    assert not invalid.is_valid()
    zero = TypedRef(None, None, None, None)  # type: ignore[arg-type]
    assert zero.is_zero()
    with pytest.raises(Exception):
        valid.id = "new-id"  # type: ignore[misc]


def test_helper_functions_and_must_parse() -> None:
    ref = new_typed_ref_with_provider(
        domain=Domain.TRACKER,
        entity_type=EntityType.ARTIFACT,
        provider=Provider.GITHUB,
        id="gh:123",
    )
    assert ref.entity_id == "gh:123"
    assert ref.canonical() == "tracker:artifact:github:gh%3A123"
    assert must_parse_typed_ref(ref.canonical()) == ref
