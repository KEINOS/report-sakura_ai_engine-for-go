# Three-run comparison

| Model | Run 1 pass | Run 2 pass | Run 3 pass | Pass range | Run 1 median | Run 2 median | Run 3 median |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `preview/Qwen3-0.6B-cpu` | 0/19 | 0/19 | 0/19 | 0 | 11.88s | 10.17s | 10.28s |
| `preview/Phi-4-mini-instruct-cpu` | 9/19 | 8/19 | 9/19 | 1 | 43.83s | 44.94s | 37.39s |
| `preview/Phi-4-multimodal-instruct` | 10/19 | 7/19 | 10/19 | 3 | 2.04s | 2.04s | 2.04s |
| `preview/Qwen3-VL-30B-A3B-Instruct` | 3/19 | 3/19 | 3/19 | 0 | 4.09s | 4.09s | 4.06s |
| `gpt-oss-120b` | 18/19 | 7/19 | 13/19 | 11 | 8.88s | 8.91s | 8.60s |
| `Qwen3-Coder-30B-A3B-Instruct` | 0/19 | 0/19 | 0/19 | 0 | 3.68s | 3.64s | 4.04s |
| `llm-jp-3.1-8x13b-instruct4` | 0/19 | 0/19 | 0/19 | 0 | 7.13s | 7.73s | 7.75s |
| `preview/Qwen3.6-35B-A3B` | 18/19 | 18/19 | 18/19 | 0 | 50.68s | 50.14s | 50.56s |
| `Qwen3-Coder-480B-A35B-Instruct-FP8` | 17/19 | 17/19 | 17/19 | 0 | 8.56s | 8.51s | 8.46s |
| `preview/Kimi-K2.6` | 13/19 | 19/19 | 13/19 | 6 | 151.63s | 101.10s | 103.59s |

## Quality-gated Pareto fronts

- Run 1: `gpt-oss-120b`, `preview/Qwen3.6-35B-A3B`, `Qwen3-Coder-480B-A35B-Instruct-FP8`
- Run 2: `preview/Qwen3.6-35B-A3B`, `Qwen3-Coder-480B-A35B-Instruct-FP8`, `preview/Kimi-K2.6`
- Run 3: `preview/Qwen3.6-35B-A3B`, `Qwen3-Coder-480B-A35B-Instruct-FP8`

Qwen3.6 and Qwen3-Coder-480B were the only high-quality models with identical functional scores across all three runs. Kimi and gpt-oss showed the largest run-to-run variation.
