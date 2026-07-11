from __future__ import annotations

import os
from pathlib import Path
from typing import Union


class PathBoundaryError(ValueError):
    """Raised when a workflow document path escapes the repository root."""


def repository_root(repo_root: Path) -> Path:
    return repo_root.resolve(strict=True)


def _comparison_path(path: Path) -> str:
    value = str(path)
    if os.name == "nt" and value.startswith("\\\\?\\"):
        value = value[4:]
    return os.path.normcase(os.path.normpath(value))


def is_within_repository(*, repo_root: Path, target: Path) -> bool:
    root = repository_root(repo_root)
    try:
        root_value = _comparison_path(root)
        target_value = _comparison_path(target.resolve(strict=False))
        return os.path.commonpath([root_value, target_value]) == root_value
    except ValueError:
        return False


def resolve_repository_path(
    *,
    repo_root: Path,
    value: Union[str, Path],
    require_file: bool = True,
    reject_absolute: bool = False,
) -> Path:
    root = repository_root(repo_root)
    raw = Path(value)
    if reject_absolute and raw.is_absolute():
        raise PathBoundaryError(f"absolute path is not allowed: {value}")
    candidate = raw if raw.is_absolute() else root / raw
    resolved = candidate.resolve(strict=False)
    if not is_within_repository(repo_root=root, target=resolved):
        raise PathBoundaryError(f"path escapes repository root: {value}")
    if os.name == "nt" and str(resolved).startswith("\\\\?\\"):
        resolved = Path(str(resolved)[4:])
    if require_file and not resolved.is_file():
        raise FileNotFoundError(f"document does not exist: {value}")
    return resolved
