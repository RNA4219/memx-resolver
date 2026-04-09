from __future__ import annotations

import json
from pathlib import Path
from typing import Any, Protocol


class ReceiptStore(Protocol):
    def load(self, *, repo_root: Path) -> list[dict]:
        ...

    def save(self, *, repo_root: Path, receipts: list[dict]) -> None:
        ...


class ResolveCacheStore(Protocol):
    def load(self, *, repo_root: Path, cache_key: str, signature: str) -> dict[str, Any] | None:
        ...

    def save(self, *, repo_root: Path, cache_key: str, signature: str, payload: dict[str, Any]) -> None:
        ...


class JsonReceiptStore:
    def __init__(self, *, path: str = ".workflow-cache/memx-doc-receipts.json") -> None:
        self._path = path

    def _target(self, repo_root: Path) -> Path:
        return repo_root / self._path

    def load(self, *, repo_root: Path) -> list[dict]:
        path = self._target(repo_root)
        if not path.is_file():
            return []
        payload = json.loads(path.read_text(encoding="utf-8"))
        return payload if isinstance(payload, list) else []

    def save(self, *, repo_root: Path, receipts: list[dict]) -> None:
        path = self._target(repo_root)
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(json.dumps(receipts, ensure_ascii=False, indent=2), encoding="utf-8")


class JsonResolveCacheStore:
    def __init__(self, *, path: str = ".workflow-cache/memx-doc-resolve.json") -> None:
        self._path = path

    def _target(self, repo_root: Path) -> Path:
        return repo_root / self._path

    def _load_map(self, *, repo_root: Path) -> dict[str, Any]:
        path = self._target(repo_root)
        if not path.is_file():
            return {}
        payload = json.loads(path.read_text(encoding="utf-8"))
        return payload if isinstance(payload, dict) else {}

    def load(self, *, repo_root: Path, cache_key: str, signature: str) -> dict[str, Any] | None:
        payload = self._load_map(repo_root=repo_root)
        entry = payload.get(cache_key)
        if not isinstance(entry, dict):
            return None
        if entry.get("signature") != signature:
            return None
        result = entry.get("payload")
        return result if isinstance(result, dict) else None

    def save(self, *, repo_root: Path, cache_key: str, signature: str, payload: dict[str, Any]) -> None:
        current = self._load_map(repo_root=repo_root)
        current[cache_key] = {"signature": signature, "payload": payload}
        path = self._target(repo_root)
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(json.dumps(current, ensure_ascii=False, indent=2), encoding="utf-8")
