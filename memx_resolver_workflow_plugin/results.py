from __future__ import annotations

from dataclasses import dataclass
from typing import Any


@dataclass(frozen=True)
class DocsResolveResult:
    required: list[dict[str, Any]]
    recommended: list[dict[str, Any]]
    errors: list[str]
    warnings: list[str]


@dataclass(frozen=True)
class DocsAckResult:
    receipts: list[dict[str, Any]]


@dataclass(frozen=True)
class DocsStaleResult:
    task_id: str
    stale: list[dict[str, Any]]
