# Sakura AI Model Routing

This routing is derived from the five Go code-generation benchmark runs in this repository. Prices are the recorded input-plus-output unit price per 10,000 tokens as of June 21, 2026.

| Profile | Model | Five-run functional result | Median response | Unit price | Suggested use |
| :------ | :---- | :------------------------- | :-------------- | :--------- | :------------ |
| `reliable` | `preview/Qwen3.6-35B-A3B` | 18/19 in all five runs | 50.56 s | ¥1.80 | Correctness-sensitive review and coding analysis |
| `balanced` | `Qwen3-Coder-480B-A35B-Instruct-FP8` | 17, 17, 17, 16, 17 | 8.51 s | ¥2.80 | Strong coding work with lower latency |
| `rapid` | `preview/Phi-4-multimodal-instruct` | 10, 7, 10, 10, 10 | 2.04 s | ¥0.40 | Reversible triage and candidate generation |
| `opportunistic` | `gpt-oss-120b` | 18, 7, 13, 18, 19 | 8.60 s | ¥0.90 | Cheap independent ideas when variance is acceptable |
| `peak` | `preview/Kimi-K2.6` | 13, 19, 13, 19, 19 | 125.31 s | ¥3.60 | High-upside experiments tolerant of delay and variance |

## Selection Order

1. Choose `reliable` when a wrong answer would waste implementation or review time.
2. Choose `balanced` when interactive latency matters and coding quality must remain strong.
3. Choose `rapid` only when a weak answer is easy to reject mechanically or with a stronger reviewer.
4. Choose `opportunistic` for diversity rather than reproducibility.
5. Choose `peak` only when its best-case quality justifies the slowest and most expensive profile.

## Evidence Limits

- The benchmark covers three Go tasks, not general knowledge, prose, vision, or every programming language.
- Each model-task pair was sampled once per run at `temperature: 0`; service-side behavior still varied.
- Latency, price, availability, and model behavior can change.
- Unit price is not actual request cost because token counts vary.
- Treat every profile as a routing prior, not permission to trust an unverified response.
