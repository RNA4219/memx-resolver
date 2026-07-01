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


@dataclass(frozen=True, init=False)
class DocsStaleResult:
    task_id: str
    status: str
    stale_reasons: list[dict[str, Any]]

    def __init__(
        self,
        task_id: str,
        stale_reasons: list[dict[str, Any]] | None = None,
        status: str | None = None,
        stale: list[dict[str, Any]] | None = None,
    ) -> None:
        reasons = stale_reasons if stale_reasons is not None else stale
        if reasons is None:
            reasons = []
        object.__setattr__(self, "task_id", task_id)
        object.__setattr__(self, "stale_reasons", reasons)
        object.__setattr__(self, "status", status or ("stale" if reasons else "fresh"))

    @property
    def stale(self) -> list[dict[str, Any]]:
        return self.stale_reasons
