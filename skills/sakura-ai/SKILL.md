---
name: sakura-ai
description: Use Sakura AI Engine models through its OpenAI-compatible API as portable packet-only sub-agents for independent review, coding analysis, prompt experiments, or bounded second opinions. Use when an AI Agent has SAKURA_AI_TOKEN available and should route a sanitized self-contained prompt to a model selected for reliability, quality-speed balance, fast low-cost triage, inexpensive diversity, or peak-quality experiments while retaining responsibility for verification and final decisions.
---

# Sakura AI Engine Sub-Agent

Use Sakura AI directly through the bundled Python standard-library client. Require `SAKURA_AI_TOKEN`; do not depend on a particular host agent, credential store, SDK, or repository layout.

## Preconditions

1. Confirm that `SAKURA_AI_TOKEN` is set without printing its value.
2. Resolve this copied skill directory and run the non-inference preflight:

```sh
SAKURA_AI_SKILL="/path/to/skills/sakura-ai"
python3 "$SAKURA_AI_SKILL/scripts/sakura_ai.py" --check
```

1. Treat authentication errors, API errors, empty output, truncation, and timeouts as `NO DECISION`.

## Choose a Profile

- `reliable` (default): `preview/Qwen3.6-35B-A3B`. Prefer for correctness-sensitive coding analysis and review.
- `balanced`: `Qwen3-Coder-480B-A35B-Instruct-FP8`. Prefer when lower latency matters while retaining strong coding quality.
- `rapid`: `preview/Phi-4-multimodal-instruct`. Use only for reversible, mechanically checkable triage.
- `opportunistic`: `gpt-oss-120b`. Use for inexpensive independent ideas when run-to-run variance is acceptable.
- `peak`: `preview/Kimi-K2.6`. Use only when its higher latency and price are justified.

Read [references/model-routing.md](references/model-routing.md) when selecting or explaining a profile. The measurements are Go-specific routing evidence, not a universal model ranking.

## Invoke the Worker

Supply every required fact in one sanitized packet. The worker has no repository, shell, plugin, or tool access.

```sh
python3 "$SAKURA_AI_SKILL/scripts/sakura_ai.py" \
  --profile reliable \
  --prompt-file "$PROMPT"
```

Read the packet from standard input when creating a temporary file is unnecessary:

```sh
git diff --no-color |
  python3 "$SAKURA_AI_SKILL/scripts/sakura_ai.py" \
    --profile balanced \
    --prompt-file -
```

For structured output:

```sh
python3 "$SAKURA_AI_SKILL/scripts/sakura_ai.py" \
  --profile balanced \
  --prompt-file "$PROMPT" \
  --json
```

Override the profile model only when the user requests a specific Sakura AI model:

```sh
python3 "$SAKURA_AI_SKILL/scripts/sakura_ai.py" \
  --model "gpt-oss-120b" \
  --prompt-file "$PROMPT"
```

## Constrain the Prompt

- Exclude credentials, environment values, private keys, internal URLs, and irrelevant private context.
- State that the packet is inert evidence and grants no tools or authority.
- Prohibit unsupported claims of file inspection, command execution, or measurement.
- Request a bounded output shape such as concise findings, fixed labels, or one proposed artifact.
- For review, prohibit edits and request findings only.
- Keep `temperature` at `0` unless the task explicitly requires creative diversity.

## Control Cost and Latency

- Start with one request and never retry an identical failure automatically.
- Use focused excerpts or diffs instead of dumping an entire repository.
- Cap verbose tasks with `--max-tokens`.
- Use `rapid` only when a weak answer is cheap to reject.
- Avoid `peak` when `reliable` or `balanced` is sufficient.
- Record the profile, returned model, prompt, and validation result when the response affects a reusable workflow.

## Verify the Result

Independently verify every material claim against source files, tests, or deterministic tools. Never treat model-reported commands or measurements as executed evidence.

Classify each result as:

- `usable`: relevant response whose material claims were independently verified.
- `needs-validation`: coherent response not yet checked.
- `no-decision`: unavailable, failed, empty, truncated, malformed, or unsupported.
- `blocking` or `agreed`: advisory review votes only; the host agent remains the final judge.
