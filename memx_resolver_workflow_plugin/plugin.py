from __future__ import annotations

import hashlib
import json
from pathlib import Path
from .markdown import linked_markdown_paths, read_markdown
from .policy import DocsSelectionPolicy
from .results import DocsAckResult, DocsResolveResult, DocsStaleResult
from .store import JsonReceiptStore, JsonResolveCacheStore, ReceiptStore, ResolveCacheStore


def _version_for_path(path: Path) -> str:
    digest = hashlib.sha1(path.read_bytes()).hexdigest()
    return digest[:12]


class MemxResolverWorkflowPlugin:
    capabilities = ("docs.resolve", "docs.ack", "docs.stale_check")

    def __init__(
        self,
        *,
        receipts_path: str = ".workflow-cache/memx-doc-receipts.json",
        resolve_cache_path: str = ".workflow-cache/memx-doc-resolve.json",
        receipt_store: ReceiptStore | None = None,
        resolve_cache_store: ResolveCacheStore | None = None,
        cache_enabled: bool = True,
        selection_policy: DocsSelectionPolicy | None = None,
    ) -> None:
        self._receipt_store = receipt_store or JsonReceiptStore(path=receipts_path)
        self._resolve_cache_store = resolve_cache_store or JsonResolveCacheStore(path=resolve_cache_path)
        self._cache_enabled = cache_enabled
        self._selection_policy = selection_policy or DocsSelectionPolicy()

    def resolve_docs(self, *, repo_root: Path, task_id: str, intent_id: str | None = None) -> DocsResolveResult:
        task_path = self._find_task_path(repo_root, task_id)
        if task_path is None:
            return DocsResolveResult(
                required=[],
                recommended=[],
                errors=[f"Task '{task_id}' was not found."],
                warnings=[],
            )

        text, _, _ = read_markdown(task_path)
        core_docs = self._selection_policy.core_paths(repo_root=repo_root)
        recommended_docs = self._selection_policy.recommended_paths(repo_root=repo_root)
        linked_docs = self._linked_docs(task_path, text)
        acceptance_docs: list[Path] = []
        if intent_id:
            for acceptance_path in self._selection_policy.acceptance_paths(repo_root=repo_root):
                _, fields, _ = read_markdown(acceptance_path)
                if fields.get("intent_id") == intent_id or fields.get("task_id") == task_id:
                    acceptance_docs.append(acceptance_path)

        cache_key = json.dumps({"task_id": task_id, "intent_id": intent_id or ""}, ensure_ascii=False, sort_keys=True)
        signature = self._build_resolve_signature(
            repo_root=repo_root,
            task_path=task_path,
            core_docs=core_docs,
            recommended_docs=recommended_docs,
            linked_docs=linked_docs,
            acceptance_docs=acceptance_docs,
        )
        if self._cache_enabled:
            cached = self._resolve_cache_store.load(
                repo_root=repo_root,
                cache_key=cache_key,
                signature=signature,
            )
            if cached is not None:
                return DocsResolveResult(
                    required=list(cached.get("required", [])),
                    recommended=list(cached.get("recommended", [])),
                    errors=[str(item) for item in cached.get("errors", [])],
                    warnings=[str(item) for item in cached.get("warnings", [])],
                )

        required = self._render_entries(
            core_docs + linked_docs + acceptance_docs,
            repo_root=repo_root,
            reason=self._selection_policy.required_reason,
        )
        recommended = self._render_entries(
            recommended_docs,
            repo_root=repo_root,
            reason=self._selection_policy.recommended_reason,
        )
        payload = DocsResolveResult(required=required, recommended=recommended, errors=[], warnings=[])
        if self._cache_enabled:
            self._resolve_cache_store.save(
                repo_root=repo_root,
                cache_key=cache_key,
                signature=signature,
                payload={
                    "required": payload.required,
                    "recommended": payload.recommended,
                    "errors": payload.errors,
                    "warnings": payload.warnings,
                },
            )
        return payload

    def ack_docs(self, *, repo_root: Path, task_id: str, doc_ids: list[str], reader: str) -> DocsAckResult:
        receipts = self._receipt_store.load(repo_root=repo_root)
        updated: list[dict] = []
        for doc_id in doc_ids:
            target = (repo_root / doc_id).resolve()
            if not target.is_file():
                continue
            record = {
                "task_id": task_id,
                "doc_id": Path(doc_id).as_posix(),
                "version": _version_for_path(target),
                "reader": reader,
            }
            receipts = [
                existing
                for existing in receipts
                if not (existing.get("task_id") == task_id and existing.get("doc_id") == record["doc_id"])
            ]
            receipts.append(record)
            updated.append(record)
        self._receipt_store.save(repo_root=repo_root, receipts=receipts)
        return DocsAckResult(receipts=updated)

    def stale_check(self, *, repo_root: Path, task_id: str) -> DocsStaleResult:
        stale: list[dict] = []
        for receipt in self._receipt_store.load(repo_root=repo_root):
            if receipt.get("task_id") != task_id:
                continue
            doc_id = str(receipt.get("doc_id", ""))
            target = (repo_root / doc_id).resolve()
            if not target.is_file():
                stale.append(
                    {
                        "task_id": task_id,
                        "doc_id": doc_id,
                        "previous_version": receipt.get("version", ""),
                        "current_version": "",
                        "reason": "document_missing",
                    }
                )
                continue
            current_version = _version_for_path(target)
            if current_version != receipt.get("version"):
                stale.append(
                    {
                        "task_id": task_id,
                        "doc_id": doc_id,
                        "previous_version": receipt.get("version", ""),
                        "current_version": current_version,
                        "reason": "version_mismatch",
                    }
                )
        return DocsStaleResult(task_id=task_id, stale=stale)

    def _find_task_path(self, repo_root: Path, task_id: str) -> Path | None:
        for path in (repo_root / "docs" / "tasks").glob("*.md"):
            _, fields, _ = read_markdown(path)
            if fields.get("task_id") == task_id:
                return path
        return None

    def _linked_docs(self, source_path: Path, text: str) -> list[Path]:
        return linked_markdown_paths(source_path, text)

    def _build_resolve_signature(
        self,
        *,
        repo_root: Path,
        task_path: Path,
        core_docs: list[Path],
        recommended_docs: list[Path],
        linked_docs: list[Path],
        acceptance_docs: list[Path],
    ) -> str:
        signature_items: list[tuple[str, str]] = []
        for path in [task_path, *core_docs, *recommended_docs, *linked_docs, *acceptance_docs]:
            if not path.is_file():
                continue
            signature_items.append((path.relative_to(repo_root).as_posix(), _version_for_path(path)))
        encoded = json.dumps(signature_items, ensure_ascii=False, sort_keys=True)
        return hashlib.sha1(encoded.encode("utf-8")).hexdigest()[:16]

    def _render_entries(self, paths: list[Path], *, repo_root: Path, reason: str) -> list[dict]:
        rendered: list[dict] = []
        seen: set[str] = set()
        for path in paths:
            if not path.is_file():
                continue
            relative = path.relative_to(repo_root).as_posix()
            if relative in seen:
                continue
            seen.add(relative)
            text, _, title = read_markdown(path)
            rendered.append(
                {
                    "doc_id": relative,
                    "path": relative,
                    "title": title or path.stem,
                    "version": _version_for_path(path),
                    "reason": reason,
                }
            )
        return rendered

def create_plugin(**kwargs: object) -> MemxResolverWorkflowPlugin:
    return MemxResolverWorkflowPlugin(**kwargs)
