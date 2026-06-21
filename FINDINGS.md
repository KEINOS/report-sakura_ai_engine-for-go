# Findings: second run

Date: 2026-06-21
Go: 1.26.4 darwin/arm64

## Outcome

The strongest pure-model result in the second run was
`preview/Kimi-K2.6`: all 19 functional checks passed, all three race-detector
runs passed, and the concurrency implementation passed 20 repeated race-test
runs.

Its trade-offs are substantial: ¥3.60 listed price, 101.10-second median
response time, and 26,895 completion tokens across the three requests.

`preview/Qwen3.6-35B-A3B` remained the most repeatable high-quality model
between runs, scoring 18/19 both times. Its recurring defect was failure to
convert a worker panic into the required error.

`Qwen3-Coder-480B-A35B-Instruct-FP8` again scored 17/19. Its concurrency
implementation consistently returned `context canceled` instead of preserving
the first worker error; in 20 repeated runs it failed that check 20 times and
the panic check twice.

`gpt-oss-120b` was highly unstable across runs: 18/19 in run 1, but only 7/19
in run 2 because its concurrency and options files did not compile. The top-k
file still passed all seven checks. This makes the first-run recommendation
insufficiently repeatable on the current evidence.

## Second-run table

| Model | Functional checks | Concurrency | Race tasks | Median response | Modern without `go fix` | Price |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `preview/Kimi-K2.6` | **19/19** | **6/6** | **3/3** | 101.10s | 1/3 | ¥3.60 |
| `preview/Qwen3.6-35B-A3B` | 18/19 | 5/6 | 2/3 | 50.14s | 1/3 | ¥1.80 |
| `Qwen3-Coder-480B-A35B-Instruct-FP8` | 17/19 | 5/6 | 1/3 | **8.51s** | 0/3 | ¥2.80 |
| `preview/Phi-4-mini-instruct-cpu` | 8/19 | 1/6 | 1/3 | 44.94s | 1/3 | ¥0.04 |
| `preview/Phi-4-multimodal-instruct` | 7/19 | 0/6 | 1/3 | **2.04s** | 1/3 | ¥0.40 |
| `gpt-oss-120b` | 7/19 | 0/6 | 1/3 | 8.91s | 1/3 | ¥0.90 |
| `preview/Qwen3-VL-30B-A3B-Instruct` | 3/19 | 0/6 | 0/3 | 4.09s | 0/3 | ¥0.40 |
| `preview/Qwen3-0.6B-cpu` | 0/19 | 0/6 | 0/3 | 10.17s | 0/3 | ¥0.04 |
| `Qwen3-Coder-30B-A3B-Instruct` | 0/19 | 0/6 | 0/3 | 3.64s | 0/3 | ¥0.90 |
| `llm-jp-3.1-8x13b-instruct4` | 0/19 | 0/6 | 0/3 | 7.73s | 0/3 | ¥0.90 |

## Pareto interpretation

Using all requested dimensions, the quality-gated Pareto front is:

- `preview/Qwen3.6-35B-A3B`
- `Qwen3-Coder-480B-A35B-Instruct-FP8`
- `preview/Kimi-K2.6`

Practical choices:

- Highest measured correctness: `preview/Kimi-K2.6`
- Best repeatability across both runs: `preview/Qwen3.6-35B-A3B`
- Fast high-quality response: `Qwen3-Coder-480B-A35B-Instruct-FP8`

## Reproducibility caveat

Run 1 raw responses were deleted with the previous temporary directory.
Therefore `run1-summary.json` contains the retained aggregate measurements,
while run 2 includes complete prompts, responses, source, and logs.

Large changes in Kimi and gpt-oss despite identical prompts and temperature
zero show that one generation per task is not enough for a production model
selection. A stronger next phase would run each task at least five times per
model and compare pass-rate distributions rather than single samples.
