# Five-run comparison

| Model | Run 1 | Run 2 | Run 3 | Run 4 | Run 5 | Mean pass | Pass range | Median response |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `preview/Qwen3-0.6B-cpu` | 0/19 | 0/19 | 0/19 | 0/19 | 0/19 | 0.0 | 0 | 10.75s |
| `preview/Phi-4-mini-instruct-cpu` | 9/19 | 8/19 | 9/19 | 9/19 | 8/19 | 8.6 | 1 | 43.83s |
| `preview/Phi-4-multimodal-instruct` | 10/19 | 7/19 | 10/19 | 10/19 | 10/19 | 9.4 | 3 | 2.04s |
| `preview/Qwen3-VL-30B-A3B-Instruct` | 3/19 | 3/19 | 3/19 | 3/19 | 3/19 | 3.0 | 0 | 4.09s |
| `gpt-oss-120b` | 18/19 | 7/19 | 13/19 | 18/19 | 19/19 | 15.0 | 12 | 8.60s |
| `Qwen3-Coder-30B-A3B-Instruct` | 0/19 | 0/19 | 0/19 | 0/19 | 0/19 | 0.0 | 0 | 3.68s |
| `llm-jp-3.1-8x13b-instruct4` | 0/19 | 0/19 | 0/19 | 0/19 | 0/19 | 0.0 | 0 | 7.14s |
| `preview/Qwen3.6-35B-A3B` | 18/19 | 18/19 | 18/19 | 18/19 | 18/19 | 18.0 | 0 | 50.56s |
| `Qwen3-Coder-480B-A35B-Instruct-FP8` | 17/19 | 17/19 | 17/19 | 16/19 | 17/19 | 16.8 | 1 | 8.51s |
| `preview/Kimi-K2.6` | 13/19 | 19/19 | 13/19 | 19/19 | 19/19 | 16.6 | 6 | 125.31s |

Median response is the median of the five per-run medians.

## Quality-gated Pareto fronts

- Run 1: `gpt-oss-120b`, `preview/Qwen3.6-35B-A3B`, `Qwen3-Coder-480B-A35B-Instruct-FP8`
- Run 2: `preview/Qwen3.6-35B-A3B`, `Qwen3-Coder-480B-A35B-Instruct-FP8`, `preview/Kimi-K2.6`
- Run 3: `preview/Qwen3.6-35B-A3B`, `Qwen3-Coder-480B-A35B-Instruct-FP8`
- Run 4: `gpt-oss-120b`, `preview/Qwen3.6-35B-A3B`, `Qwen3-Coder-480B-A35B-Instruct-FP8`, `preview/Kimi-K2.6`
- Run 5: `gpt-oss-120b`, `preview/Qwen3.6-35B-A3B`, `Qwen3-Coder-480B-A35B-Instruct-FP8`, `preview/Kimi-K2.6`

Qwen3.6 was the only high-quality model with an identical functional score across all five runs. Qwen3-Coder-480B varied by one test. Kimi reached 19/19 in three runs, while gpt-oss reached 19/19 in the fifth run, but both showed substantially larger run-to-run variation.
