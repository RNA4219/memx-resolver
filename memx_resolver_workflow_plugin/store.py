from __future__ import annotations

import json
import os
import tempfile
import threading
import time
from contextlib import contextmanager
from pathlib import Path
from typing import Any, Protocol

from filelock import FileLock, Timeout

from .paths import resolve_repository_path

_THREAD_LOCKS_GUARD = threading.Lock()
_THREAD_LOCKS: dict[str, Any] = {}



class JsonStoreError(RuntimeError):
    """Base error for workflow JSON stores."""


class JsonStoreCorruptionError(JsonStoreError):
    """Raised when a store contains invalid JSON or an invalid root value."""


class JsonStoreLockTimeoutError(JsonStoreError):
    """Raised when a store lock cannot be acquired in time."""


def _target(repo_root: Path, path: str) -> Path:
    return resolve_repository_path(
        repo_root=repo_root,
        value=path,
        require_file=False,
        reject_absolute=True,
    )


def _read_json(path: Path, *, expected: type) -> Any:
    if not path.is_file():
        return expected()
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, UnicodeError, json.JSONDecodeError) as exc:
        raise JsonStoreCorruptionError(f"cannot read JSON store {path}: {exc}") from exc
    if not isinstance(payload, expected):
        raise JsonStoreCorruptionError(
            f"invalid JSON root in {path}: expected {expected.__name__}"
        )
    return payload

def _replace_with_retry(source: Path, destination: Path) -> None:
    attempts = 5 if os.name == "nt" else 1
    for attempt in range(attempts):
        try:
            os.replace(source, destination)
            return
        except PermissionError:
            if attempt + 1 >= attempts:
                raise
            time.sleep(0.01 * (2**attempt))




def _atomic_write_json(path: Path, payload: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary: Path | None = None
    try:
        with tempfile.NamedTemporaryFile(
            mode="w",
            encoding="utf-8",
            newline="\n",
            dir=path.parent,
            prefix=f".{path.name}.",
            suffix=".tmp",
            delete=False,
        ) as handle:
            temporary = Path(handle.name)
            json.dump(payload, handle, ensure_ascii=False, indent=2)
            handle.write("\n")
            handle.flush()
            os.fsync(handle.fileno())
        _replace_with_retry(temporary, path)
    finally:
        if temporary is not None and temporary.exists():
            temporary.unlink()


@contextmanager
def _lock(path: Path, timeout: float):
    deadline = time.monotonic() + timeout
    key = os.path.normcase(os.path.abspath(path))
    with _THREAD_LOCKS_GUARD:
        thread_lock = _THREAD_LOCKS.setdefault(key, threading.Lock())
    if not thread_lock.acquire(timeout=max(0.0, timeout)):
        raise JsonStoreLockTimeoutError(f"timed out locking JSON store {path}")
    try:
        remaining = max(0.0, deadline - time.monotonic())
        try:
            with FileLock(str(path) + ".lock", timeout=remaining):
                yield
        except Timeout as exc:
            raise JsonStoreLockTimeoutError(
                f"timed out locking JSON store {path}"
            ) from exc
    finally:
        thread_lock.release()


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
    def __init__(
        self,
        *,
        path: str = ".workflow-cache/memx-doc-receipts.json",
        lock_timeout: float = 10.0,
    ) -> None:
        self._path = path
        self._lock_timeout = lock_timeout

    def load(self, *, repo_root: Path) -> list[dict]:
        path = _target(repo_root, self._path)
        with _lock(path, self._lock_timeout):
            return _read_json(path, expected=list)

    def save(self, *, repo_root: Path, receipts: list[dict]) -> None:
        path = _target(repo_root, self._path)
        with _lock(path, self._lock_timeout):
            _atomic_write_json(path, receipts)

    def upsert(self, *, repo_root: Path, records: list[dict]) -> list[dict]:
        path = _target(repo_root, self._path)
        with _lock(path, self._lock_timeout):
            receipts = _read_json(path, expected=list)
            for record in records:
                receipts = [
                    existing
                    for existing in receipts
                    if not (
                        existing.get("task_id") == record.get("task_id")
                        and existing.get("doc_id") == record.get("doc_id")
                    )
                ]
                receipts.append(record)
            _atomic_write_json(path, receipts)
        return records


class JsonResolveCacheStore:
    def __init__(
        self,
        *,
        path: str = ".workflow-cache/memx-doc-resolve.json",
        lock_timeout: float = 10.0,
    ) -> None:
        self._path = path
        self._lock_timeout = lock_timeout

    def load(self, *, repo_root: Path, cache_key: str, signature: str) -> dict[str, Any] | None:
        path = _target(repo_root, self._path)
        with _lock(path, self._lock_timeout):
            payload = _read_json(path, expected=dict)
        entry = payload.get(cache_key)
        if not isinstance(entry, dict) or entry.get("signature") != signature:
            return None
        result = entry.get("payload")
        return result if isinstance(result, dict) else None

    def save(self, *, repo_root: Path, cache_key: str, signature: str, payload: dict[str, Any]) -> None:
        path = _target(repo_root, self._path)
        with _lock(path, self._lock_timeout):
            current = _read_json(path, expected=dict)
            current[cache_key] = {"signature": signature, "payload": payload}
            _atomic_write_json(path, current)
