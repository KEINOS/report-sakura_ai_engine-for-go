# First-run versus second-run comparison

| Model | Run 1 pass | Run 2 pass | Δ pass | Run 1 median | Run 2 median | Δ time |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `preview/Qwen3-0.6B-cpu` | 0/19 | 0/19 | +0 | 11.88s | 10.17s | -1.70s |
| `preview/Phi-4-mini-instruct-cpu` | 9/19 | 8/19 | -1 | 43.83s | 44.94s | +1.12s |
| `preview/Phi-4-multimodal-instruct` | 10/19 | 7/19 | -3 | 2.04s | 2.04s | +0.00s |
| `preview/Qwen3-VL-30B-A3B-Instruct` | 3/19 | 3/19 | +0 | 4.09s | 4.09s | -0.00s |
| `gpt-oss-120b` | 18/19 | 7/19 | -11 | 8.88s | 8.91s | +0.03s |
| `Qwen3-Coder-30B-A3B-Instruct` | 0/19 | 0/19 | +0 | 3.68s | 3.64s | -0.04s |
| `llm-jp-3.1-8x13b-instruct4` | 0/19 | 0/19 | +0 | 7.13s | 7.73s | +0.60s |
| `preview/Qwen3.6-35B-A3B` | 18/19 | 18/19 | +0 | 50.68s | 50.14s | -0.54s |
| `Qwen3-Coder-480B-A35B-Instruct-FP8` | 17/19 | 17/19 | +0 | 8.56s | 8.51s | -0.05s |
| `preview/Kimi-K2.6` | 13/19 | 19/19 | +6 | 151.63s | 101.10s | -50.53s |

## Pareto fronts

- Run 1 quality front: gpt-oss-120b, preview/Qwen3.6-35B-A3B, Qwen3-Coder-480B-A35B-Instruct-FP8
- Run 2 quality front: preview/Qwen3.6-35B-A3B, Qwen3-Coder-480B-A35B-Instruct-FP8, preview/Kimi-K2.6

Run-to-run differences reflect stochastic inference, service load, and model-serving changes.
