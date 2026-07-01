from __future__ import annotations

from pathlib import Path

from memx_resolver_workflow_plugin.plugin import create_plugin
from memx_resolver_workflow_plugin.policy import DocsSelectionPolicy
from memx_resolver_workflow_plugin.results import DocsResolveResult
from memx_resolver_workflow_plugin.store import JsonReceiptStore, JsonResolveCacheStore
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
