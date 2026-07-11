from __future__ import annotations

import json
import os
import signal
import socket
import subprocess
import time
import urllib.error
import urllib.request
from pathlib import Path

import yaml
from jsonschema import Draft202012Validator


REPO_ROOT = Path(__file__).resolve().parents[1]
GO_ROOT = REPO_ROOT
CLI_SCHEMA_PATH = REPO_ROOT / "docs" / "memx_spec_v3" / "docs" / "contracts" / "cli-json.schema.json"
OPENAPI_PATH = REPO_ROOT / "docs" / "memx_spec_v3" / "docs" / "contracts" / "openapi.yaml"


def _process_group_kwargs() -> dict[str, object]:
    if os.name == "nt":
        return {"creationflags": subprocess.CREATE_NEW_PROCESS_GROUP}
    return {"start_new_session": True}


def _terminate_process_tree(proc: subprocess.Popen[str]) -> None:
    if proc.poll() is not None:
        return
    if os.name == "nt":
        subprocess.run(
            ["taskkill", "/PID", str(proc.pid), "/T", "/F"],
            capture_output=True,
            text=True,
            timeout=15,
            check=False,
        )
    else:
        os.killpg(proc.pid, signal.SIGTERM)
        try:
            proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            os.killpg(proc.pid, signal.SIGKILL)

def _run(args: list[str], *, cwd: Path, env: dict[str, str]) -> str:
    proc = subprocess.Popen(
        args,
        cwd=cwd,
        env=env,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        **_process_group_kwargs(),
    )
    try:
        stdout, stderr = proc.communicate(timeout=180)
    except subprocess.TimeoutExpired as exc:
        _terminate_process_tree(proc)
        stdout, stderr = proc.communicate()
        raise AssertionError(f"command timed out after 180s: {' '.join(args)}; stdout: {stdout}; stderr: {stderr}") from exc
    if proc.returncode != 0:
        raise AssertionError(
            f"command failed: {' '.join(args)}\nstdout:\n{stdout}\nstderr:\n{stderr}"
        )
    return stdout


def _build_mem(tmp_path: Path) -> tuple[Path, dict[str, str]]:
    env = os.environ.copy()
    env.update(
        {
            "GOCACHE": str(REPO_ROOT / ".tmp" / "go-build"),
            "OPENAI_API_KEY": "",
            "MEMX_OPENAI_API_KEY": "",
            "DASHSCOPE_API_KEY": "",
            "MEMX_ALIBABA_API_KEY": "",
            "MEMX_DASHSCOPE_API_KEY": "",
            "MEMX_LLM_PROVIDER": "",
        }
    )
    bin_path = tmp_path / ("mem.exe" if os.name == "nt" else "mem")
    _run(["go", "build", "-o", str(bin_path), "./cmd/mem"], cwd=GO_ROOT, env=env)
    return bin_path, env


def _free_addr() -> str:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        sock.bind(("127.0.0.1", 0))
        return f"127.0.0.1:{sock.getsockname()[1]}"


def _post_json(base_url: str, path: str, payload: dict) -> dict:
    body = json.dumps(payload).encode("utf-8")
    request = urllib.request.Request(
        f"{base_url}{path}",
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=5) as response:
            return json.loads(response.read().decode("utf-8"))
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode("utf-8")
        raise AssertionError(f"POST {path} failed: {exc.code}\n{detail}") from exc


def _wait_for_healthz(base_url: str, proc: subprocess.Popen[str]) -> None:
    deadline = time.monotonic() + 10
    while time.monotonic() < deadline:
        if proc.poll() is not None:
            stdout, stderr = proc.communicate()
            raise AssertionError(f"api server exited early\nstdout:\n{stdout}\nstderr:\n{stderr}")
        try:
            with urllib.request.urlopen(f"{base_url}/healthz", timeout=1) as response:
                if response.read().decode("utf-8") == "ok":
                    return
        except OSError:
            time.sleep(0.1)
    proc.terminate()
    stdout, stderr = proc.communicate(timeout=5)
    raise AssertionError(f"api server did not become healthy\nstdout:\n{stdout}\nstderr:\n{stderr}")


def _openapi_validators() -> dict[str, Draft202012Validator]:
    openapi = yaml.safe_load(OPENAPI_PATH.read_text(encoding="utf-8"))
    schemas = openapi["components"]["schemas"]
    return {
        name: Draft202012Validator(_component_schema(schemas, name))
        for name in [
            "DocsIngestResponse",
            "DocsResolveResponse",
            "ChunksGetResponse",
            "DocsSearchResponse",
            "CardsSearchResponse",
            "CardFeedbackResponse",
            "PromptBundleResponse",
            "TaskStateExportResponse",
            "ReadsAckResponse",
            "DocsStaleCheckResponse",
            "ContractsResolveResponse",
        ]
    }


def _component_schema(schemas: dict, name: str) -> dict:
    bundled = {"$defs": schemas, **schemas[name]}
    return json.loads(json.dumps(bundled).replace("#/components/schemas/", "#/$defs/"))


def test_docs_cli_json_outputs_match_schema(tmp_path: Path) -> None:
    schema = json.loads(CLI_SCHEMA_PATH.read_text(encoding="utf-8"))
    validator = Draft202012Validator(schema)
    bin_path, env = _build_mem(tmp_path)
    workdir = tmp_path / "work"
    workdir.mkdir()
    resolver_db = workdir / "resolver.db"

    def mem(*args: str) -> dict:
        out = _run([str(bin_path), *args], cwd=workdir, env=env)
        payload = json.loads(out)
        validator.validate(payload)
        return payload

    ingest = mem(
        "docs",
        "ingest",
        "--json",
        "--resolver",
        str(resolver_db),
        "--title",
        "Schema Contract Spec",
        "--body",
        "# Schema Contract Spec\n\n## Acceptance Criteria\n- resolver cards schema validation works",
        "--doc-type",
        "spec",
        "--version",
        "2026-03-10",
        "--feature",
        "schema-contract",
        "--tag",
        "memory",
    )
    doc_id = ingest["doc_id"]

    resolve = mem("docs", "resolve", "--json", "--resolver", str(resolver_db), "--feature", "schema-contract")
    chunk_id = resolve["required"][0]["top_chunks"][0]

    mem("docs", "chunks", "--json", "--resolver", str(resolver_db), "--chunk-id", chunk_id)
    mem(
        "docs",
        "search",
        "--json",
        "--resolver",
        str(resolver_db),
        "--doc-type",
        "spec",
        "--feature",
        "schema-contract",
        "schema validation",
    )
    cards = mem(
        "docs",
        "cards",
        "--json",
        "--resolver",
        str(resolver_db),
        "--query",
        "schema validation",
        "--feature",
        "schema-contract",
        "--memory-type",
        "acceptance",
        "--token-budget",
        "80",
    )
    card = cards["cards"][0]
    mem(
        "docs",
        "cards-feedback",
        "--json",
        "--resolver",
        str(resolver_db),
        "--card-id",
        card["card_id"],
        "--doc-id",
        card["doc_id"],
        "--chunk-id",
        card["chunk_id"],
        "--memory-type",
        card["memory_type"],
        "--signal",
        "helpful",
    )
    mem(
        "docs",
        "bundle",
        "--json",
        "--resolver",
        str(resolver_db),
        "--query",
        "schema validation",
        "--feature",
        "schema-contract",
    )
    mem("docs", "ack", "--json", "--resolver", str(resolver_db), "--task-id", "TASK.schema", "--doc-id", doc_id)
    mem("docs", "stale", "--json", "--resolver", str(resolver_db), "--task-id", "TASK.schema")
    mem(
        "docs",
        "taskstate-export",
        "--json",
        "--resolver",
        str(resolver_db),
        "--task-id",
        "TASK.schema",
        "--feature",
        "schema-contract",
    )
    mem("docs", "contract", "--json", "--resolver", str(resolver_db), "--feature", "schema-contract")


def test_docs_http_outputs_match_openapi_schema(tmp_path: Path) -> None:
    validators = _openapi_validators()
    bin_path, env = _build_mem(tmp_path)
    workdir = tmp_path / "server"
    workdir.mkdir()
    addr = _free_addr()
    base_url = f"http://{addr}"
    proc = subprocess.Popen(
        [
            str(bin_path),
            "api",
            "serve",
            "--addr",
            addr,
            "--resolver",
            str(workdir / "resolver.db"),
            "--short",
            str(workdir / "short.db"),
            "--journal",
            str(workdir / "journal.db"),
            "--knowledge",
            str(workdir / "knowledge.db"),
            "--archive",
            str(workdir / "archive.db"),
        ],
        cwd=workdir,
        env=env,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        **_process_group_kwargs(),
    )
    try:
        _wait_for_healthz(base_url, proc)

        ingest = _post_json(
            base_url,
            "/v1/docs:ingest",
            {
                "title": "OpenAPI Contract Spec",
                "body": "# OpenAPI Contract Spec\n\n## Acceptance Criteria\n- resolver cards schema validation works",
                "doc_type": "spec",
                "version": "2026-03-10",
                "feature_keys": ["schema-contract"],
                "tags": ["memory"],
            },
        )
        validators["DocsIngestResponse"].validate(ingest)
        doc_id = ingest["doc_id"]

        resolve = _post_json(base_url, "/v1/docs:resolve", {"feature": "schema-contract"})
        validators["DocsResolveResponse"].validate(resolve)
        chunk_id = resolve["required"][0]["top_chunks"][0]

        validators["ChunksGetResponse"].validate(_post_json(base_url, "/v1/chunks:get", {"chunk_ids": [chunk_id]}))
        validators["DocsSearchResponse"].validate(
            _post_json(
                base_url,
                "/v1/docs:search",
                {"query": "schema validation", "doc_types": ["spec"], "feature_keys": ["schema-contract"]},
            )
        )
        cards = _post_json(
            base_url,
            "/v1/cards:search",
            {
                "query": "schema validation",
                "feature_keys": ["schema-contract"],
                "memory_types": ["acceptance"],
                "token_budget": 80,
            },
        )
        validators["CardsSearchResponse"].validate(cards)
        card = cards["cards"][0]
        validators["CardFeedbackResponse"].validate(
            _post_json(
                base_url,
                "/v1/cards:feedback",
                {
                    "card_id": card["card_id"],
                    "doc_id": card["doc_id"],
                    "chunk_id": card["chunk_id"],
                    "memory_type": card["memory_type"],
                    "signal": "helpful",
                },
            )
        )
        validators["PromptBundleResponse"].validate(
            _post_json(
                base_url,
                "/v1/cards:bundle",
                {"query": "schema validation", "feature": "schema-contract", "token_budget": 80},
            )
        )
        validators["ReadsAckResponse"].validate(
            _post_json(base_url, "/v1/reads:ack", {"task_id": "TASK.openapi", "doc_id": doc_id})
        )
        validators["DocsStaleCheckResponse"].validate(
            _post_json(base_url, "/v1/docs:stale-check", {"task_id": "TASK.openapi"})
        )
        validators["TaskStateExportResponse"].validate(
            _post_json(base_url, "/v1/taskstate:export", {"task_id": "TASK.openapi", "feature": "schema-contract"})
        )
        validators["ContractsResolveResponse"].validate(
            _post_json(base_url, "/v1/contracts:resolve", {"feature": "schema-contract"})
        )
    finally:
        _terminate_process_tree(proc)
        proc.communicate(timeout=5)
