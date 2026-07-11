"""Strict v2 typed_ref support for the memx-resolver workflow plugin.

The v2 wire format is:
<domain>:<entity_type>:<provider>:<entity_id>

Domain, entity type, and provider are canonicalized to lowercase. Entity IDs
are decoded on input and RFC 3986 encoded on output so colons remain
reversible.
"""

from __future__ import annotations

import re
from dataclasses import dataclass
from enum import Enum
from typing import Union
from urllib.parse import quote, unquote_to_bytes


class Domain(str, Enum):
    """Domain represents the system area for typed_ref."""

    MEMX = "memx"
    AGENT_TASKSTATE = "agent-taskstate"
    TRACKER = "tracker"


class EntityType(str, Enum):
    """Compatibility enum for common entity types."""

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
    """Compatibility enum for common providers."""

    LOCAL = "local"
    JIRA = "jira"
    GITHUB = "github"
    LINEAR = "linear"


DEFAULT_PROVIDER = Provider.LOCAL
KNOWN_DOMAINS = frozenset(domain.value for domain in Domain)
_HEX = frozenset("0123456789abcdefABCDEF")
_UNRESERVED = "-._~"


def _component_value(value: Union[Enum, str, None]) -> str:
    if isinstance(value, Enum):
        return str(value.value)
    return "" if value is None else str(value)


def _decode_entity_id(raw: str) -> str:
    if not isinstance(raw, str) or not raw:
        raise TypedRefParseError("empty id")
    index = 0
    while index < len(raw):
        if raw[index] == "%":
            if (
                index + 2 >= len(raw)
                or raw[index + 1] not in _HEX
                or raw[index + 2] not in _HEX
            ):
                raise TypedRefParseError("malformed percent encoding in id")
            index += 3
        else:
            index += 1
    try:
        decoded = unquote_to_bytes(raw).decode("utf-8")
    except UnicodeDecodeError as exc:
        raise TypedRefParseError("invalid UTF-8 percent encoding in id") from exc
    if not decoded:
        raise TypedRefParseError("empty id")
    return decoded


def _encode_entity_id(value: str) -> str:
    if not isinstance(value, str) or not value:
        raise TypedRefParseError("empty id")
    return quote(value, safe=_UNRESERVED)


@dataclass(frozen=True)
class TypedRef:
    """Typed reference using the canonical four-segment v2 format."""

    domain: Union[Domain, str]
    type: Union[EntityType, str]
    provider: Union[Provider, str]
    id: str

    def __str__(self) -> str:
        domain_val = _component_value(self.domain).lower()
        type_val = _component_value(self.type).lower()
        provider_val = _component_value(self.provider).lower()
        return f"{domain_val}:{type_val}:{provider_val}:{_encode_entity_id(self.id)}"

    @property
    def entity_type(self) -> str:
        return _component_value(self.type)

    @property
    def entity_id(self) -> str:
        return self.id

    def ref(self) -> str:
        return str(self)

    def canonical(self) -> str:
        return str(self)

    def is_zero(self) -> bool:
        return not self.domain and not self.type and not self.provider and not self.id

    def is_valid(self) -> bool:
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
    """Error raised when a typed_ref is malformed."""


def parse_typed_ref(s: str) -> TypedRef:
    """Parse a strict v2 four-segment typed_ref."""

    if not isinstance(s, str):
        raise TypedRefParseError("ref must be a string")
    s = s.strip()
    if not s:
        raise TypedRefParseError("empty ref")
    parts = s.split(":")
    if len(parts) != 4:
        raise TypedRefParseError(
            f"invalid ref format: {s} (v2 requires <domain>:<type>:<provider>:<id>)"
        )
    return _parse_four_segment(parts)


def _parse_four_segment(parts: list[str]) -> TypedRef:
    domain_str = parts[0].lower()
    _validate_domain(domain_str)
    entity_type_str = parts[1].lower()
    _validate_entity_type(domain_str, entity_type_str)
    provider_str = parts[2].lower()
    _validate_provider(provider_str)
    id_part = _decode_entity_id(parts[3])
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
    if "%" in domain:
        raise TypedRefParseError("percent encoding is only allowed in id")


def _validate_entity_type(domain: str, entity_type: str) -> None:
    if not entity_type:
        raise TypedRefParseError("empty entity type")
    if "%" in entity_type:
        raise TypedRefParseError("percent encoding is only allowed in id")


def _validate_provider(provider: str) -> None:
    if not provider:
        raise TypedRefParseError("empty provider")
    if "%" in provider:
        raise TypedRefParseError("percent encoding is only allowed in id")


def _validate_id(entity_id: str) -> None:
    _decode_entity_id(entity_id)


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
    return TypedRef(
        domain=domain,
        type=entity_type,
        provider=provider,
        id=id,
    )


def must_parse_typed_ref(s: str) -> TypedRef:
    return parse_typed_ref(s)
