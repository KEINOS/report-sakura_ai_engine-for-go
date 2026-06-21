#!/usr/bin/env python3
from __future__ import annotations

import difflib
import json
import os
import random
import re
import shutil
import statistics
import subprocess
import sys
import tempfile
import time
from pathlib import Path

import httpx

ROOT = Path(__file__).resolve().parent
FIXTURES = ROOT / "fixtures"
PROMPTS = ROOT / "prompts"
RESPONSES = ROOT / "responses"
LOGS = ROOT / "logs"
API = "https://api.ai.sakura.ad.jp/v1"
HERMES_PYTHON = Path("/opt/homebrew/Cellar/hermes-agent/2026.6.5/libexec/bin/python")

MODELS = [
    "preview/Qwen3-0.6B-cpu",
    "preview/Phi-4-mini-instruct-cpu",
    "preview/Phi-4-multimodal-instruct",
    "preview/Qwen3-VL-30B-A3B-Instruct",
    "gpt-oss-120b",
    "Qwen3-Coder-30B-A3B-Instruct",
    "llm-jp-3.1-8x13b-instruct4",
    "preview/Qwen3.6-35B-A3B",
    "Qwen3-Coder-480B-A35B-Instruct-FP8",
    "preview/Kimi-K2.6",
]

PRICE = {
    "preview/Qwen3-0.6B-cpu": (0.04, 0.01, 0.03),
    "preview/Phi-4-mini-instruct-cpu": (0.04, 0.01, 0.03),
    "preview/Phi-4-multimodal-instruct": (0.40, 0.10, 0.30),
    "preview/Qwen3-VL-30B-A3B-Instruct": (0.40, 0.10, 0.30),
    "gpt-oss-120b": (0.90, 0.15, 0.75),
    "Qwen3-Coder-30B-A3B-Instruct": (0.90, 0.15, 0.75),
    "llm-jp-3.1-8x13b-instruct4": (0.90, 0.15, 0.75),
    "preview/Qwen3.6-35B-A3B": (1.80, 0.30, 1.50),
    "Qwen3-Coder-480B-A35B-Instruct-FP8": (2.80, 0.30, 2.50),
    "preview/Kimi-K2.6": (3.60, 0.60, 3.00),
}

TASKS = {
    "topk": """Target Go 1.26. Return only one complete Go source file using package topk.

Implement:
    func TopKFrequent(words []string, k int) []string

Requirements:
- Return at most k unique words, ordered by descending frequency.
- Resolve equal frequencies by ascending lexical order.
- k <= 0 or empty input returns a non-nil empty slice.
- Do not mutate words.
- Optimize for large inputs and small k: avoid sorting all input occurrences.
- Use only the standard library.
- Produce concise, idiomatic, maintainable Go 1.26 code.""",
    "pmap": """Target Go 1.26. Return only one complete Go source file using package pmap.

Implement:
    func ParallelMapOrdered(
        ctx context.Context,
        input []int,
        workers int,
        fn func(context.Context, int) (int, error),
    ) ([]int, error)

Requirements:
- Preserve input order in successful output.
- Run no more than workers calls concurrently.
- workers <= 0 and nil fn return descriptive errors.
- Empty input returns a non-nil empty slice.
- On the first worker error or parent cancellation, cancel remaining work, wait for all started goroutines, and return nil plus an error preserving errors.Is.
- Convert a panic from fn into a descriptive error containing the panic value; never let it crash the process.
- Do not leak goroutines or race.
- Optimize for low overhead.
- Use only the standard library.
- Produce concise, idiomatic, maintainable Go 1.26 code.""",
    "options": """Target Go 1.26. Return only one complete Go source file using package options.

Define:
    type Options struct {
        Limit  *int
        Labels map[string]string
    }
    type ParseError struct {
        Field string
        Value string
        Err   error
    }

ParseError must implement error and Unwrap.

Implement:
    func ParseOptions(spec string, cause error) (Options, error)

Grammar and behavior:
- spec is comma-separated key=value fields; trim surrounding spaces around keys and values.
- Supported keys: limit and label.NAME, where NAME must be non-empty.
- limit is an integer from 1 through 1000.
- Default limit is a non-nil pointer to 100.
- Labels is always a non-nil map. Duplicate labels use the last value.
- Empty spec is valid.
- Invalid syntax, unknown fields, and invalid limits return the partially parsed Options plus *ParseError.
- If cause contains a *ParseError anywhere in its error tree, parse spec normally, then return the parsed Options and that same *ParseError.
- Use only the standard library.
- Prefer efficient, concise, idiomatic modern Go 1.26 syntax and APIs from the first response.""",
}


def safe(model: str) -> str:
    return re.sub(r"[^A-Za-z0-9_.-]+", "__", model)


def prepare() -> None:
    for directory in (PROMPTS, RESPONSES, LOGS):
        directory.mkdir(parents=True, exist_ok=True)
    for task, prompt in TASKS.items():
        (PROMPTS / f"{task}.md").write_text(prompt + "\n", encoding="utf-8")


def resolve_token() -> str:
    if token := os.getenv("SAKURA_AI_TOKEN"):
        return token
    env = os.environ.copy()
    env["PYTHONPATH"] = str(Path.home() / ".hermes/hermes-agent")
    code = """
from hermes_cli.runtime_provider import _try_resolve_from_custom_pool
r = _try_resolve_from_custom_pool("https://api.ai.sakura.ad.jp/v1", "custom")
if not r: raise SystemExit(1)
print(r["api_key"], end="")
"""
    return subprocess.run(
        [str(HERMES_PYTHON), "-c", code],
        env=env,
        check=True,
        capture_output=True,
        text=True,
    ).stdout


def extract_code(text: str) -> str:
    blocks = re.findall(r"```(?:go)?\s*(.*?)```", text, re.DOTALL | re.IGNORECASE)
    code = max(blocks, key=len) if blocks else text
    if (start := code.find("package ")) >= 0:
        code = code[start:]
    return code.strip() + "\n"


def max_tokens(model: str) -> int:
    if model == "preview/Kimi-K2.6":
        return 16384
    if model == "llm-jp-3.1-8x13b-instruct4":
        return 3000
    if model in {"preview/Qwen3-0.6B-cpu", "preview/Phi-4-mini-instruct-cpu"}:
        return 2048
    return 8192


def generate() -> None:
    prepare()
    token = resolve_token()
    jobs = [(model, task) for task in TASKS for model in MODELS]
    random.Random(20260621).shuffle(jobs)
    request_log = []
    with httpx.Client(timeout=httpx.Timeout(360.0, connect=30.0)) as client:
        for index, (model, task) in enumerate(jobs, 1):
            directory = RESPONSES / safe(model) / task
            directory.mkdir(parents=True, exist_ok=True)
            if (directory / "solution.go").exists():
                print(f"[{index:02d}/30] cached {model} :: {task}", flush=True)
                continue
            print(f"[{index:02d}/30] {model} :: {task}", flush=True)
            started = time.perf_counter()
            body = {
                "model": model,
                "messages": [
                    {
                        "role": "system",
                        "content": "You are an expert Go engineer. Follow the requested API exactly. Output source code only.",
                    },
                    {"role": "user", "content": TASKS[task]},
                ],
                "temperature": 0,
                "max_tokens": max_tokens(model),
                "stream": False,
            }
            try:
                response = client.post(
                    f"{API}/chat/completions",
                    headers={
                        "Authorization": f"Bearer {token}",
                        "Accept": "application/json",
                    },
                    json=body,
                )
                elapsed = time.perf_counter() - started
                payload = response.json()
                record = {
                    "model": model,
                    "task": task,
                    "status": response.status_code,
                    "elapsed_seconds": elapsed,
                    "usage": payload.get("usage", {}),
                }
                if response.is_error:
                    record["error"] = payload
                    (directory / "error.json").write_text(
                        json.dumps(record, ensure_ascii=False, indent=2), encoding="utf-8"
                    )
                else:
                    choice = payload["choices"][0]
                    message = choice["message"]
                    content = message.get("content")
                    record.update(
                        {
                            "finish_reason": choice.get("finish_reason"),
                            "message_keys": list(message),
                            "reasoning_chars": len(message.get("reasoning") or ""),
                        }
                    )
                    (directory / "response.json").write_text(
                        json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8"
                    )
                    if isinstance(content, str) and content.strip():
                        (directory / "solution.go").write_text(
                            extract_code(content), encoding="utf-8"
                        )
                    else:
                        record["error"] = "empty content"
                        (directory / "error.json").write_text(
                            json.dumps(record, ensure_ascii=False, indent=2),
                            encoding="utf-8",
                        )
                request_log.append(record)
                print(f"  {elapsed:.2f}s {record.get('usage', {})}", flush=True)
            except Exception as error:
                record = {
                    "model": model,
                    "task": task,
                    "elapsed_seconds": time.perf_counter() - started,
                    "error": repr(error),
                }
                request_log.append(record)
                (directory / "error.json").write_text(
                    json.dumps(record, ensure_ascii=False, indent=2), encoding="utf-8"
                )
                print(f"  ERROR {error}", flush=True)
    (LOGS / "requests.json").write_text(
        json.dumps(request_log, ensure_ascii=False, indent=2), encoding="utf-8"
    )


def command(args: list[str], cwd: Path, timeout: int = 60) -> dict:
    try:
        result = subprocess.run(
            args, cwd=cwd, capture_output=True, text=True, timeout=timeout
        )
        return {
            "ok": result.returncode == 0,
            "returncode": result.returncode,
            "stdout": result.stdout,
            "stderr": result.stderr,
        }
    except subprocess.TimeoutExpired as error:
        return {
            "ok": False,
            "returncode": 124,
            "stdout": error.stdout or "",
            "stderr": (error.stderr or "") + f"\ntimeout after {timeout}s",
        }


def test_results(output: str) -> dict:
    actions: dict[str, str] = {}
    for line in output.splitlines():
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        if name := event.get("Test"):
            if event.get("Action") in {"pass", "fail"}:
                actions[name] = event["Action"]
    leaves = {
        name
        for name in actions
        if not any(other.startswith(name + "/") for other in actions)
    }
    passed = sorted(name for name in leaves if actions[name] == "pass")
    failed = sorted(name for name in leaves if actions[name] == "fail")
    return {"passed": passed, "failed": failed, "passed_count": len(passed)}


def benchmark_metrics(output: str) -> dict:
    matches = re.findall(
        r"^(Benchmark\S+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"
        r"(?:\s+([\d.]+)\s+B/op)?(?:\s+([\d.]+)\s+allocs/op)?",
        output,
        re.MULTILINE,
    )
    if not matches:
        return {}
    name, ns, bytes_op, allocs = matches[-1]
    return {
        "name": name,
        "ns_per_op": float(ns),
        "bytes_per_op": float(bytes_op or 0),
        "allocs_per_op": float(allocs or 0),
    }


def evaluate_one(model: str, task: str) -> dict:
    directory = RESPONSES / safe(model) / task
    source = directory / "solution.go"
    result = {"model": model, "task": task, "available": source.exists()}
    metadata = directory / "response.json"
    error = directory / "error.json"
    if metadata.exists():
        payload = json.loads(metadata.read_text(encoding="utf-8"))
        choice = payload["choices"][0]
        result["usage"] = payload.get("usage", {})
        result["finish_reason"] = choice.get("finish_reason")
    if error.exists():
        result["generation_error"] = json.loads(error.read_text(encoding="utf-8"))
    requests = json.loads((LOGS / "requests.json").read_text(encoding="utf-8"))
    for request in requests:
        if request["model"] == model and request["task"] == task:
            result["elapsed_seconds"] = request["elapsed_seconds"]
            result.setdefault("usage", request.get("usage", {}))
            break
    if not source.exists():
        return result

    code = source.read_text(encoding="utf-8")
    result["loc"] = sum(
        1
        for line in code.splitlines()
        if line.strip() and not line.lstrip().startswith("//")
    )
    with tempfile.TemporaryDirectory(prefix=f"go-model-{task}-") as temp:
        work = Path(temp)
        shutil.copytree(FIXTURES / task, work, dirs_exist_ok=True)
        shutil.copy2(source, work / "solution.go")
        formatted = command(["gofmt", "-d", "solution.go"], work)
        result["gofmt_clean"] = formatted["stdout"] == ""
        result["gofmt_diff"] = formatted["stdout"]
        result["build"] = command(["go", "build", "./..."], work)
        if not result["build"]["ok"]:
            return result
        test = command(
            ["go", "test", "-json", "-count=1", "-timeout=15s", "./..."],
            work,
            timeout=20,
        )
        result["test"] = {**test, **test_results(test["stdout"])}
        result["vet"] = command(["go", "vet", "./..."], work)
        result["lint"] = command(
            [
                "golangci-lint",
                "run",
                "--no-config",
                "--default=standard",
                "--timeout=30s",
                "--output.text.colors=false",
            ],
            work,
            timeout=40,
        )
        before = (work / "solution.go").read_text(encoding="utf-8")
        fix = command(["go", "fix", "./..."], work)
        after = (work / "solution.go").read_text(encoding="utf-8")
        diff = "\n".join(
            difflib.unified_diff(
                before.splitlines(),
                after.splitlines(),
                fromfile="before.go",
                tofile="after.go",
                lineterm="",
            )
        )
        result["go_fix"] = {**fix, "changed": before != after, "diff": diff}
        if test["ok"]:
            result["race"] = command(
                ["go", "test", "-race", "-count=1", "-timeout=30s", "./..."],
                work,
                timeout=40,
            )
        else:
            result["race"] = {"ok": False, "stderr": "skipped: tests failed"}
        if test["ok"] and result["race"]["ok"]:
            bench = command(
                ["go", "test", "-run=^$", "-bench=.", "-benchmem", "-count=3", "./..."],
                work,
                timeout=240,
            )
            result["benchmark"] = {**bench, "metrics": benchmark_metrics(bench["stdout"])}
    return result


def evaluate() -> None:
    rows = []
    for model in MODELS:
        for task in TASKS:
            print(f"validate {model} :: {task}", flush=True)
            rows.append(evaluate_one(model, task))
            (LOGS / "evaluation.json").write_text(
                json.dumps(rows, ensure_ascii=False, indent=2), encoding="utf-8"
            )


def repeat_reliability() -> None:
    rows = json.loads((LOGS / "evaluation.json").read_text(encoding="utf-8"))
    candidates = sorted(
        {
            row["model"]
            for row in rows
            if row["task"] == "pmap" and row.get("build", {}).get("ok")
        }
    )
    repeats = {}
    for model in candidates:
        with tempfile.TemporaryDirectory(prefix="repeat-pmap-") as temp:
            work = Path(temp)
            shutil.copytree(FIXTURES / "pmap", work, dirs_exist_ok=True)
            shutil.copy2(RESPONSES / safe(model) / "pmap" / "solution.go", work / "solution.go")
            outcome = command(
                ["go", "test", "-race", "-v", "-count=20", "-timeout=90s", "./..."],
                work,
                timeout=100,
            )
            repeats[model] = outcome
            (LOGS / f"repeat-{safe(model)}.log").write_text(
                outcome["stdout"] + outcome["stderr"], encoding="utf-8"
            )
    (LOGS / "repeat-summary.json").write_text(
        json.dumps(repeats, ensure_ascii=False, indent=2), encoding="utf-8"
    )


def summarize() -> None:
    rows = json.loads((LOGS / "evaluation.json").read_text(encoding="utf-8"))
    summary = {}
    for model in MODELS:
        items = [row for row in rows if row["model"] == model]
        passed = sum(item.get("test", {}).get("passed_count", 0) for item in items)
        pmap = next(item for item in items if item["task"] == "pmap")
        elapsed = [item.get("elapsed_seconds", 360.0) for item in items]
        functionality = passed / 19
        reliability = (
            pmap.get("test", {}).get("passed_count", 0) / 6
            + sum(bool(item.get("race", {}).get("ok")) for item in items) / 3
        ) / 2
        usability = (
            sum(bool(item.get("gofmt_clean")) for item in items) / 3
            + sum(bool(item.get("vet", {}).get("ok")) for item in items) / 3
            + sum(bool(item.get("lint", {}).get("ok")) for item in items) / 3
        ) / 3
        modern = (
            sum(
                bool(item.get("build", {}).get("ok"))
                and item.get("go_fix", {}).get("changed") is False
                for item in items
            )
            / 3
        )
        successful = [item for item in items if item.get("test", {}).get("ok")]
        effectiveness = len(successful) / 3
        summary[model] = {
            "functionality": functionality,
            "effectiveness": effectiveness,
            "reliability": reliability,
            "usability": usability,
            "response_seconds": statistics.median(elapsed),
            "modern": modern,
            "price": PRICE[model][0],
            "passed": passed,
            "requirements": 19,
            "pmap_passed": pmap.get("test", {}).get("passed_count", 0),
            "race_tasks": sum(bool(item.get("race", {}).get("ok")) for item in items),
            "modern_tasks": round(modern * 3),
            "loc": sum(item.get("loc", 0) for item in items),
            "completion_tokens": sum(
                item.get("usage", {}).get("completion_tokens", 0) for item in items
            ),
        }

    def dominates(a: dict, b: dict) -> bool:
        larger = ["functionality", "effectiveness", "reliability", "usability", "modern"]
        smaller = ["response_seconds", "price"]
        no_worse = all(a[key] >= b[key] for key in larger) and all(
            a[key] <= b[key] for key in smaller
        )
        better = any(a[key] > b[key] for key in larger) or any(
            a[key] < b[key] for key in smaller
        )
        return no_worse and better

    front = [
        model
        for model in MODELS
        if not any(
            dominates(summary[other], summary[model])
            for other in MODELS
            if other != model
        )
    ]
    quality = [model for model in MODELS if summary[model]["functionality"] >= 0.8]
    quality_front = [
        model
        for model in quality
        if not any(
            dominates(summary[other], summary[model])
            for other in quality
            if other != model
        )
    ]
    output = {
        "date": "2026-06-21",
        "models": summary,
        "pareto_front": front,
        "quality_pareto_front": quality_front,
    }
    (ROOT / "run2-summary.json").write_text(
        json.dumps(output, ensure_ascii=False, indent=2), encoding="utf-8"
    )

    baseline = json.loads((ROOT / "baseline_run1.json").read_text(encoding="utf-8"))
    lines = [
        "# First-run versus second-run comparison",
        "",
        "| Model | Run 1 pass | Run 2 pass | Δ pass | Run 1 median | Run 2 median | Δ time |",
        "|---|---:|---:|---:|---:|---:|---:|",
    ]
    for model in MODELS:
        old = baseline["models"][model]
        new = summary[model]
        lines.append(
            f"| `{model}` | {old['passed']}/19 | {new['passed']}/19 | "
            f"{new['passed'] - old['passed']:+d} | {old['median_seconds']:.2f}s | "
            f"{new['response_seconds']:.2f}s | "
            f"{new['response_seconds'] - old['median_seconds']:+.2f}s |"
        )
    lines.extend(
        [
            "",
            "## Pareto fronts",
            "",
            f"- Run 1 quality front: {', '.join(baseline['quality_pareto_front'])}",
            f"- Run 2 quality front: {', '.join(quality_front)}",
            "",
            "Run-to-run differences reflect stochastic inference, service load, and model-serving changes.",
        ]
    )
    (ROOT / "comparison.md").write_text("\n".join(lines) + "\n", encoding="utf-8")


def main() -> None:
    action = sys.argv[1] if len(sys.argv) > 1 else "all"
    if action in {"generate", "all"}:
        generate()
    if action in {"evaluate", "all"}:
        evaluate()
    if action in {"repeat", "all"}:
        repeat_reliability()
    if action in {"summarize", "all"}:
        summarize()


if __name__ == "__main__":
    main()
