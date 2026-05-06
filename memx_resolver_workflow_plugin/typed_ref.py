"""typed_ref module for memx-resolver workflow plugin.

Provides typed reference parsing and validation following the canonical format:
<domain>:<entity_type>:<provider>:<entity_id>

Example: memx:evidence:local:01HXXXXXXX

Migration support: 3-segment format (memx:<type>:<id>) is also accepted
and normalized to provider='local'.
"""

from __future__ import annotations

from dataclasses import dataclass
from enum import Enum
from typing import Optional, Union


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


class Provider(str, Enum):
    """Provider represents the data source for typed_ref."""
    LOCAL = "local"
    JIRA = "jira"
    GITHUB = "github"
    LINEAR = "linear"


DEFAULT_PROVIDER = Provider.LOCAL


@dataclass(frozen=True)
class TypedRef:
    """Typed reference to an entity.

    Canonical format: <domain>:<entity_type>:<provider>:<entity_id>
    """
    domain: Domain
    type: Union[EntityType, str]  # EntityType for memx, str for other domains
    provider: Provider
    id: str

    def __str__(self) -> str:
        """Return canonical format string."""
        domain_val = self.domain.value if self.domain else ""
        type_val = self.type.value if isinstance(self.type, EntityType) else self.type
        provider_val = self.provider.value if self.provider else ""
        return f"{domain_val}:{type_val}:{provider_val}:{self.id}"

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
        return bool(self.domain and self.type and self.provider and self.id)


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
        return _parse_three_segment(parts)

    if len(parts) == 4:
        return _parse_four_segment(parts)

    raise TypedRefParseError(
        f"invalid ref format: {s} (expected <domain>:<type>:[<provider>:]<id>)"
    )


def _parse_three_segment(parts: list[str]) -> TypedRef:
    """Parse 3-segment format and normalize to provider='local'."""
    domain_str = parts[0]
    if domain_str != Domain.MEMX.value:
        raise TypedRefParseError(
            f"invalid domain: {domain_str} (expected 'memx' for 3-segment format)"
        )

    entity_type_str = parts[1]
    _validate_entity_type(domain_str, entity_type_str)

    id_part = parts[2]
    if not id_part:
        raise TypedRefParseError("empty id in ref")

    return TypedRef(
        domain=Domain.MEMX,
        type=EntityType(entity_type_str),
        provider=DEFAULT_PROVIDER,
        id=id_part,
    )


def _parse_four_segment(parts: list[str]) -> TypedRef:
    """Parse 4-segment canonical format."""
    domain_str = parts[0]
    _validate_domain(domain_str)

    entity_type_str = parts[1]
    _validate_entity_type(domain_str, entity_type_str)

    provider_str = parts[2]
    if not provider_str:
        raise TypedRefParseError("empty provider in ref")

    id_part = parts[3]
    if not id_part:
        raise TypedRefParseError("empty id in ref")

    # For memx domain, use EntityType enum; for other domains, use string
    if domain_str == Domain.MEMX.value:
        entity_type = EntityType(entity_type_str)
    else:
        entity_type = entity_type_str

    return TypedRef(
        domain=Domain(domain_str),
        type=entity_type,
        provider=Provider(provider_str),
        id=id_part,
    )


def _validate_domain(domain: str) -> None:
    """Validate domain is known."""
    valid_domains = {d.value for d in Domain}
    if domain not in valid_domains:
        raise TypedRefParseError(
            f"invalid domain: {domain} (expected memx, agent-taskstate, or tracker)"
        )


def _validate_entity_type(domain: str, entity_type: str) -> None:
    """Validate entity type.

    For memx domain, only known types are allowed.
    For other domains, just check non-empty.
    """
    if not entity_type:
        raise TypedRefParseError("empty entity type")

    if domain == Domain.MEMX.value:
        valid_types = {t.value for t in EntityType}
        if entity_type not in valid_types:
            raise TypedRefParseError(
                f"invalid entity type for memx domain: {entity_type}"
            )


def new_typed_ref(entity_type: EntityType, id: str) -> TypedRef:
    """Create TypedRef with memx domain and local provider."""
    return TypedRef(
        domain=Domain.MEMX,
        type=entity_type,
        provider=DEFAULT_PROVIDER,
        id=id,
    )


def new_typed_ref_with_provider(
    domain: Domain,
    entity_type: EntityType,
    provider: Provider,
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