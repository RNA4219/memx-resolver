from __future__ import annotations

from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
import pytest
from filelock import FileLock

from memx_resolver_workflow_plugin.plugin import create_plugin
from memx_resolver_workflow_plugin.policy import DocsSelectionPolicy
from memx_resolver_workflow_plugin.results import DocsResolveResult
from memx_resolver_workflow_plugin.paths import PathBoundaryError
from memx_resolver_workflow_plugin.store import (
    JsonReceiptStore,
    JsonResolveCacheStore,
    JsonStoreCorruptionError,
    JsonStoreLockTimeoutError,
)
from memx_resolver_workflow_plugin.markdown import linked_markdown_paths


def _write(path: Path, text: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text.strip() + "\n", encoding="utf-8")


def test_resolve_ack_and_stale_cycle(tmp_path: Path) -> None:
    _write(tmp_path / "README.md", "# Repo")
    _write(tmp_path / "BLUEPRINT.md", "# Blueprint")
    _write(tmp_path / "RUNBOOK.md", "# Runbook")
    _write(tmp_path / "EVALUATION.md", "# Evaluation")
    _write(tmp_path / "CHECKLISTS.md", "# Checklists")
    _write(
        tmp_path / "docs" / "tasks" / "task-sample.md",
        """
---
task_id: 20260410-01
intent_id: INT-001
owner: docs-core
status: active
---

# Task Seed

- [design](../design.md)
""",
    )
    _write(tmp_path / "docs" / "design.md", "# Design")

    plugin = create_plugin()
    resolved = plugin.resolve_docs(repo_root=tmp_path, task_id="20260410-01")
    assert isinstance(resolved, DocsResolveResult)
    assert resolved.required

    doc_ids = [entry["doc_id"] for entry in resolved.required[:2]]
    acked = plugin.ack_docs(
        repo_root=tmp_path,
        task_id="20260410-01",
        doc_ids=doc_ids,
        reader="tester",
    )
    assert len(acked.receipts) == 2

    _write(tmp_path / doc_ids[0], "# Repo changed")
    stale = plugin.stale_check(repo_root=tmp_path, task_id="20260410-01")
    assert stale.status == "stale"
    assert stale.stale_reasons[0]["doc_id"] == doc_ids[0]
    assert stale.stale[0]["doc_id"] == doc_ids[0]


def test_receipt_and_resolve_cache_stores_persist_payloads(tmp_path: Path) -> None:
    receipt_store = JsonReceiptStore(path=".workflow-cache/test-receipts.json")
    receipt_store.save(repo_root=tmp_path, receipts=[{"task_id": "T-1", "doc_id": "README.md"}])
    assert receipt_store.load(repo_root=tmp_path)[0]["task_id"] == "T-1"

    cache_store = JsonResolveCacheStore(path=".workflow-cache/test-resolve.json")
    cache_store.save(
        repo_root=tmp_path,
        cache_key="T-1",
        signature="sig-1",
        payload={"required": [{"doc_id": "README.md"}], "recommended": [], "errors": [], "warnings": []},
    )
    assert cache_store.load(repo_root=tmp_path, cache_key="T-1", signature="sig-1") is not None


def test_resolve_docs_uses_signature_cache_and_refreshes_when_inputs_change(tmp_path: Path) -> None:
    _write(tmp_path / "README.md", "# Repo")
    _write(tmp_path / "BLUEPRINT.md", "# Blueprint")
    _write(tmp_path / "RUNBOOK.md", "# Runbook")
    _write(tmp_path / "EVALUATION.md", "# Evaluation")
    _write(tmp_path / "CHECKLISTS.md", "# Checklists")
    _write(
        tmp_path / "docs" / "tasks" / "task-sample.md",
        """
---
task_id: 20260410-01
intent_id: INT-001
owner: docs-core
status: active
---

# Task Seed
""",
    )

    plugin = create_plugin()
    first = plugin.resolve_docs(repo_root=tmp_path, task_id="20260410-01")
    _write(tmp_path / "README.md", "# Repo changed")
    second = plugin.resolve_docs(repo_root=tmp_path, task_id="20260410-01")

    assert first.required[0]["version"] != second.required[0]["version"]


def test_selection_policy_can_override_core_docs(tmp_path: Path) -> None:
    _write(tmp_path / "CUSTOM.md", "# Custom")
    _write(
        tmp_path / "docs" / "tasks" / "task-sample.md",
        """
---
task_id: 20260410-01
intent_id: INT-001
owner: docs-core
status: active
---

# Task Seed
""",
    )
    plugin = create_plugin(
        selection_policy=DocsSelectionPolicy(
            core_docs=("CUSTOM.md",),
            recommended_docs=(),
        )
    )

    resolved = plugin.resolve_docs(repo_root=tmp_path, task_id="20260410-01")

    assert resolved.required[0]["doc_id"] == "CUSTOM.md"


def test_linked_markdown_paths_resolves_markdown_links_with_fragments(tmp_path: Path) -> None:
    task_path = tmp_path / "docs" / "tasks" / "task-sample.md"
    design_path = tmp_path / "docs" / "design.md"
    _write(design_path, "# Design")
    _write(
        task_path,
        """
# Task Seed

- [design acceptance](../design.md#acceptance)
- [design query](../design.md?plain=1)
""",
    )

    assert linked_markdown_paths(task_path, task_path.read_text(encoding="utf-8")) == [design_path.resolve()]


@pytest.mark.parametrize("doc_id", ["../outside.md", "C:/outside.md"])
def test_ack_docs_rejects_paths_outside_repository(tmp_path: Path, doc_id: str) -> None:
    repo_root = tmp_path / "repo"
    _write(repo_root / "README.md", "# Repo")
    plugin = create_plugin(receipts_path=".workflow-cache/test-receipts.json")

    with pytest.raises(PathBoundaryError):
        plugin.ack_docs(
            repo_root=repo_root,
            task_id="TASK.path",
            doc_ids=[doc_id],
            reader="tester",
        )

    assert not (repo_root / ".workflow-cache" / "test-receipts.json").exists()


def test_ack_docs_rejects_symlink_escape(tmp_path: Path) -> None:
    repo_root = tmp_path / "repo"
    repo_root.mkdir()
    outside = tmp_path / "outside.md"
    _write(outside, "# Outside")
    link = repo_root / "escape.md"
    try:
        link.symlink_to(outside)
    except OSError as exc:
        pytest.skip(f"symlinks are unavailable: {exc}")

    with pytest.raises(PathBoundaryError):
        create_plugin().ack_docs(
            repo_root=repo_root,
            task_id="TASK.path",
            doc_ids=["escape.md"],
            reader="tester",
        )


def test_stale_check_reports_tampered_external_receipt(tmp_path: Path) -> None:
    repo_root = tmp_path / "repo"
    repo_root.mkdir()
    store = JsonReceiptStore(path=".workflow-cache/tampered.json")
    store.save(
        repo_root=repo_root,
        receipts=[
            {
                "task_id": "TASK.path",
                "doc_id": "../outside.md",
                "version": "old",
                "reader": "tester",
            }
        ],
    )

    result = create_plugin(receipt_store=store).stale_check(
        repo_root=repo_root,
        task_id="TASK.path",
    )

    assert result.status == "stale"
    assert result.stale_reasons[0]["reason"] == "path_outside_repo"


def test_resolve_docs_warns_for_external_markdown_link(tmp_path: Path) -> None:
    repo_root = tmp_path / "repo"
    outside = tmp_path / "outside.md"
    _write(outside, "# Outside")
    _write(
        repo_root / "docs" / "tasks" / "task.md",
        """
---
task_id: TASK.link
---
# Task
- [outside](../../../outside.md)
""",
    )

    result = create_plugin().resolve_docs(repo_root=repo_root, task_id="TASK.link")

    assert result.warnings
    assert "outside repository root" in result.warnings[0]


def test_receipt_store_upsert_is_concurrency_safe(tmp_path: Path) -> None:
    repo_root = tmp_path / "repo"
    repo_root.mkdir()
    store = JsonReceiptStore(path=".workflow-cache/concurrent.json")
    records = [
        {"task_id": "TASK.concurrent", "doc_id": f"docs/{index}.md", "version": str(index)}
        for index in range(20)
    ]

    with ThreadPoolExecutor(max_workers=8) as executor:
        list(
            executor.map(
                lambda record: store.upsert(repo_root=repo_root, records=[record]),
                records,
            )
        )

    persisted = store.load(repo_root=repo_root)
    assert len(persisted) == len(records)
    assert {item["doc_id"] for item in persisted} == {item["doc_id"] for item in records}


def test_json_store_rejects_corrupt_payload(tmp_path: Path) -> None:
    repo_root = tmp_path / "repo"
    target = repo_root / ".workflow-cache" / "corrupt.json"
    _write(target, "{not-json")
    store = JsonReceiptStore(path=".workflow-cache/corrupt.json")

    with pytest.raises(JsonStoreCorruptionError):
        store.load(repo_root=repo_root)


def test_json_store_lock_timeout_is_explicit(tmp_path: Path) -> None:
    repo_root = tmp_path / "repo"
    target = repo_root / ".workflow-cache" / "locked.json"
    target.parent.mkdir(parents=True)
    lock = FileLock(str(target) + ".lock")
    store = JsonReceiptStore(path=".workflow-cache/locked.json", lock_timeout=0.01)

    with lock:
        with pytest.raises(JsonStoreLockTimeoutError):
            store.load(repo_root=repo_root)
