from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True)
class DocsSelectionPolicy:
    core_docs: tuple[str, ...] = ("README.md", "BLUEPRINT.md", "RUNBOOK.md", "EVALUATION.md")
    recommended_docs: tuple[str, ...] = ("CHECKLISTS.md",)
    acceptance_glob: str = "docs/acceptance/AC-*.md"
    required_reason: str = "task dependency"
    recommended_reason: str = "workflow reference"

    def core_paths(self, *, repo_root: Path) -> list[Path]:
        return [repo_root / relative for relative in self.core_docs]

    def recommended_paths(self, *, repo_root: Path) -> list[Path]:
        return [repo_root / relative for relative in self.recommended_docs]

    def acceptance_paths(self, *, repo_root: Path) -> list[Path]:
        return sorted(repo_root.glob(self.acceptance_glob))
