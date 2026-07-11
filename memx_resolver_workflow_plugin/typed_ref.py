"""Canonical typed_ref support for the memx-resolver workflow plugin.

Provides typed reference parsing and validation following the canonical format:
<domain>:<entity_type>:<provider>:<entity_id>

Example: memx:evidence:local:01HXXXXXXX

Migration support: 3-segment format (memx:<type>:<id>) is also accepted
and normalized to provider='local'.
"""

from __future__ import annotations

import warnings
from dataclasses import dataclass
from enum import Enum
from typing import Union


class Domain(str, Enum):
    """Domain represents the system area for typed_ref."""
    MEMX = "memx"
    AGENT_TASKSTATE = "agent-taskstate"
    TRACKER = "tracker"


class EntityType(str, Enum):
    """EntityType represents entity types within memx domain."""
    EVIDENCE = "evidence"
    EVIDENCE_CHUNK = "evidence_chunk"
    KNOWLEDGE = "knowledge"
    ARTIFACT = "artifact"
    LINEAGE = "lineage"
    DOC = "doc"
    CHUNK = "chunk"
    CARD = "card"
    TASK = "task"


class Provider(str, Enum):
    """Provider represents the data source for typed_ref."""
    LOCAL = "local"
    JIRA = "jira"
    GITHUB = "github"
    LINEAR = "linear"


DEFAULT_PROVIDER = Provider.LOCAL

KNOWN_DOMAINS = frozenset(domain.value for domain in Domain)


def _component_value(value: Union[Enum, str, None]) -> str:
    if isinstance(value, Enum):
        return str(value.value)
    return "" if value is None else str(value)


@dataclass(frozen=True)
class TypedRef:
    """Typed reference to an entity.

    Canonical format: <domain>:<entity_type>:<provider>:<entity_id>
    """
    domain: Union[Domain, str]
    type: Union[EntityType, str]
    provider: Union[Provider, str]
    id: str

    def __str__(self) -> str:
        """Return canonical format string."""
        domain_val = _component_value(self.domain).lower()
        type_val = _component_value(self.type).lower()
        provider_val = _component_value(self.provider).lower()
        return f"{domain_val}:{type_val}:{provider_val}:{self.id}"

    @property
    def entity_type(self) -> str:
        return _component_value(self.type)

    @property
    def entity_id(self) -> str:
        return self.id

    def ref(self) -> str:
        """Alias for __str__."""
        return str(self)

    def canonical(self) -> str:
        """Return canonical format."""
        return str(self)

    def is_zero(self) -> bool:
        """Check if TypedRef is zero value."""
        return not self.domain and not self.type and not self.provider and not self.id

    def is_valid(self) -> bool:
        """Check if TypedRef is valid."""
        try:
            domain = _component_value(self.domain).lower()
            _validate_domain(domain)
            _validate_entity_type(domain, _component_value(self.type))
            _validate_provider(_component_value(self.provider))
            _validate_id(self.id)
        except TypedRefParseError:
            return False
        return True


class TypedRefParseError(ValueError):
    """Error parsing typed_ref string."""
    pass


def parse_typed_ref(s: str) -> TypedRef:
    """Parse typed_ref string into TypedRef.

    Accepts both 3-segment (memx:<type>:<id>) and 4-segment
    (memx:<type>:<provider>:<id>) formats.

    3-segment format is normalized to provider='local'.

    Args:
        s: typed_ref string to parse

    Returns:
        TypedRef instance

    Raises:
        TypedRefParseError: if parsing fails
    """
    s = s.strip()
    if not s:
        raise TypedRefParseError("empty ref")

    parts = s.split(":")

    if len(parts) == 3:
        warnings.warn(
            "three-segment typed_ref is deprecated; use domain:type:provider:id",
            DeprecationWarning,
            stacklevel=2,
        )
        return _parse_three_segment(parts)

    if len(parts) == 4:
        return _parse_four_segment(parts)

    raise TypedRefParseError(
        f"invalid ref format: {s} (expected <domain>:<type>:[<provider>:]<id>)"
    )


def _parse_three_segment(parts: list[str]) -> TypedRef:
    """Parse legacy format and normalize it to the canonical provider."""
    domain_str = parts[0].lower()
    _validate_domain(domain_str)
    entity_type_str = parts[1].lower()
    _validate_entity_type(domain_str, entity_type_str)
    id_part = parts[2]
    _validate_id(id_part)
    return TypedRef(
        domain=Domain(domain_str),
        type=_known_entity_type_or_string(entity_type_str),
        provider=DEFAULT_PROVIDER,
        id=id_part,
    )


def _parse_four_segment(parts: list[str]) -> TypedRef:
    """Parse the canonical four-segment format."""
    domain_str = parts[0].lower()
    _validate_domain(domain_str)
    entity_type_str = parts[1].lower()
    _validate_entity_type(domain_str, entity_type_str)
    provider_str = parts[2].lower()
    _validate_provider(provider_str)
    id_part = parts[3]
    _validate_id(id_part)
    return TypedRef(
        domain=Domain(domain_str),
        type=_known_entity_type_or_string(entity_type_str),
        provider=_known_provider_or_string(provider_str),
        id=id_part,
    )


def _validate_domain(domain: str) -> None:
    if domain not in KNOWN_DOMAINS:
        raise TypedRefParseError(
            f"invalid domain: {domain} (expected memx, agent-taskstate, or tracker)"
        )


def _validate_entity_type(domain: str, entity_type: str) -> None:
    """Entity types are extensible; only a non-empty value is required."""
    if not entity_type:
        raise TypedRefParseError("empty entity type")


def _validate_provider(provider: str) -> None:
    if not provider:
        raise TypedRefParseError("empty provider")


def _validate_id(entity_id: str) -> None:
    if not entity_id:
        raise TypedRefParseError("empty id")


def _known_entity_type_or_string(value: str) -> Union[EntityType, str]:
    try:
        return EntityType(value)
    except ValueError:
        return value


def _known_provider_or_string(value: str) -> Union[Provider, str]:
    try:
        return Provider(value)
    except ValueError:
        return value


def new_typed_ref(entity_type: Union[EntityType, str], id: str) -> TypedRef:
    """Create TypedRef with memx domain and local provider."""
    return TypedRef(
        domain=Domain.MEMX,
        type=entity_type,
        provider=DEFAULT_PROVIDER,
        id=id,
    )


def new_typed_ref_with_provider(
    domain: Union[Domain, str],
    entity_type: Union[EntityType, str],
    provider: Union[Provider, str],
    id: str,
) -> TypedRef:
    """Create TypedRef with specified provider."""
    return TypedRef(
        domain=domain,
        type=entity_type,
        provider=provider,
        id=id,
    )


def must_parse_typed_ref(s: str) -> TypedRef:
    """Parse typed_ref or raise TypedRefParseError.

    Same as parse_typed_ref but name follows Go convention.
    """
    return parse_typed_ref(s)