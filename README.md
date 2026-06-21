---
title: "Sakura.AI Go コーディングモデル比較調査"
description: "A benchmark-based comparison of Go coding models on Sakura.AI OpenAI-compatible endpoints."
date: 2026-06-21
lastmod: 2026-06-22
type: "blog"
draft: false
tags:
	- go
	- llm
	- benchmark
	- sakura-ai
	- sakura-ai-engine
categories:
	- research
---

# Sakura AI Engine：AI モデルの Go コーディング能力比較

## 概要

AI Agent で利用する Go コーディングモデルを選定するため、**[さくらの AI Engine](https://manual.sakura.ad.jp/cloud/manual-ai-engine.html)（以下 Sakura AI）が提供する 10 モデルを比較・評価した**。

AI Agent やハーネスによる会話履歴、メモリー、自己学習コンテキストの影響を避けるため、Sakura AI の OpenAI API 互換エンドポイントへ各モデルを直接リクエストし、初回の Go コード生成能力を測定した。

各モデルに同一の system prompt、user prompt、`temperature: 0` を設定し、3 種類の課題をそれぞれ 1 回ずつ生成させた。回答は人手で修正せず、Go 1.26.4 環境で同じ調査を 3 回実施した。課題の作成と調査の実施には Codex を利用している。

このリポジトリには、第 2 回・第 3 回調査の各 30 件の API 応答、抽出コード、検証ログを保存している。第 1 回の生データは保存されていないが、集計値は `baseline_run1.json` に残っている。

- 調査日：2026 年 6 月 21 日
- 調査員：Codex (GPT-5.5, 推論: Medium)
- 監修・編集：[KEINOS](https://github.com/KEINOS/)
- 実行環境：Go 1.26.4、darwin/arm64
- API エンドポイント：`https://api.ai.sakura.ad.jp/v1/chat/completions`

## 目次

- [概要](#概要)
- [結論](#結論)
- [評価対象](#評価対象)
- [評価課題](#評価課題)
- [評価方法](#評価方法)
- [調査結果](#調査結果)
  - [第 1 回結果](#第-1-回結果)
  - [第 2 回結果](#第-2-回結果)
  - [第 3 回結果](#第-3-回結果)
- [Pareto front による候補選定](#pareto-front-による候補選定)
- [3 回の結果比較](#3-回の結果比較)
- [モデル別所見](#モデル別所見)
- [制約と次の検証案](#制約と次の検証案)
- [再実行方法](#再実行方法)
- [検証ファイル一覧](#検証ファイル一覧)

## 結論

3 回の結果を通した用途別の推奨は次のとおり。

- 正確性と安定性：`preview/Qwen3.6-35B-A3B`
- 正確性を保った応答速度：`Qwen3-Coder-480B-A35B-Instruct-FP8`
- 低価格・超高速の補助用途：`preview/Phi-4-multimodal-instruct`

`preview/Qwen3.6-35B-A3B` は 3 回すべてで 18/19 を記録し、最も安定した高品質モデルだった。再現性を重視する場合の第一候補である。

`Qwen3-Coder-480B-A35B-Instruct-FP8` も 3 回すべてで 17/19 を記録し、応答時間中央値は各回とも約 8.5 秒だった。並行処理とエラー処理にはレビューが必要だが、品質と速度のバランスに優れる。

`preview/Kimi-K2.6` は第 2 回に唯一 19/19 を達成した一方、3 回の推移は 13 → 19 → 13 だった。最高到達点は高いが、価格、待ち時間、出力安定性が課題となる。

`gpt-oss-120b` は 18 → 7 → 13 と大きく変動した。Kimi とあわせ、単発の好結果だけでは既定モデルを選定できないことを示している。

## 評価対象

| モデル | 価格（入力 / 出力） |
| :----- | :----------: |
| `preview/Qwen3-0.6B-cpu` | ¥0.04（¥0.01 / ¥0.03） |
| `preview/Phi-4-mini-instruct-cpu` | ¥0.04（¥0.01 / ¥0.03） |
| `preview/Phi-4-multimodal-instruct` | ¥0.40（¥0.10 / ¥0.30） |
| `preview/Qwen3-VL-30B-A3B-Instruct` | ¥0.40（¥0.10 / ¥0.30） |
| `gpt-oss-120b` | ¥0.90（¥0.15 / ¥0.75） |
| `Qwen3-Coder-30B-A3B-Instruct` | ¥0.90（¥0.15 / ¥0.75） |
| `llm-jp-3.1-8x13b-instruct4` | ¥0.90（¥0.15 / ¥0.75） |
| `preview/Qwen3.6-35B-A3B` | ¥1.80（¥0.30 / ¥1.50） |
| `Qwen3-Coder-480B-A35B-Instruct-FP8` | ¥2.80（¥0.30 / ¥2.50） |
| `preview/Kimi-K2.6` | ¥3.60（¥0.60 / ¥3.00） |

> 価格は 10,000 トークンあたりの入力・出力料金である（2026/06/21 検証日時点）。
> 検証は、[基盤モデル無償プラン](https://ai.sakura.ad.jp/sakura-ai/ai-engine/)で行い、無料枠のリクエスト数 3,000 回/月のうち合計 100 程度を消費した。
> コード生成では、入力より出力のトークン数が大幅に多くなる（本調査では 6〜40 倍）。従量制課金プランで追試する場合は、出力料金の影響に注意が必要である。

## 評価課題

### TopKFrequent

単語の出現頻度を集計し、頻度降順・辞書順で上位 `k` 件を返す課題。正確性に加えて、大規模入力かつ小さな `k` に対するアルゴリズム効率、入力非破壊、非 `nil` の空スライスを検証した。

### ParallelMapOrdered

順序を維持する並行 map 処理。worker 数制限、親 context のキャンセル、最初のエラーの保持、panic のエラー化、goroutine leak、race の有無を検証した。本調査で Reliability の差を最も強く判別した課題である。

### ParseOptions

文字列オプションを解析する課題。部分結果、型付きエラー、`errors.Is`、デフォルト値、重複ラベル、Go 1.26 の構文・API を検証した。

## 評価方法

次の観測値を評価ベクトルに対応させ、[Pareto front](https://en.wikipedia.org/wiki/Pareto_front) を求めた。ここでは「その候補より全評価軸で同等以上、かつ少なくとも 1 軸で優れる別の候補が存在しない候補の集合」を指す。その中から、選定方針に応じて最終候補を絞り込んだ。

| 評価軸 | 主な観測方法 |
| :----- | :----------- |
| Functionality | 3 課題、合計 19 件のブラックボックス・テスト |
| Effectiveness | テスト合格、生成行数、benchmark の実行時間・allocation |
| Reliability | 並行処理 6 テスト、race detector、20 回反復試験 |
| Usability | `gofmt`、`go vet`、`golangci-lint`、生成行数 |
| Responsiveness | API リクエスト開始から完了までの経過秒数の中央値 |
| Modern Syntax | 初回回答へ `go fix ./...` を適用した際の変更有無 |
| Cost | 測定時の入力・出力料金と、API 応答のトークン数 |

各生成コードに対して、原則として次の順で検証した。

```text
go build ./...
go test -json -count=1 -timeout=15s ./...
go vet ./...
golangci-lint run --no-config --default=standard
go fix ./...
go test -race -count=1 -timeout=30s ./...
go test -run=^$ -bench=. -benchmem -count=3 ./...
```

通常テストまたは race test に失敗したコードは benchmark の対象外とした。生成コードは採点前に修正していない。

## 調査結果

### 第 1 回結果

| モデル | 機能 | 並行処理 | Race | 中央応答時間 | 初回から Modern | 価格 |
| :----- | :--: | :------: | :--: | :----------: | :------------: | :--: |
| `gpt-oss-120b` | **18/19** | 未記録 | 未記録 | 8.88秒 | 0/3 | ¥0.90 |
| `preview/Qwen3.6-35B-A3B` | **18/19** | 未記録 | 未記録 | 50.68秒 | 1/3 | ¥1.80 |
| `Qwen3-Coder-480B-A35B-Instruct-FP8` | 17/19 | 未記録 | 未記録 | 8.56秒 | 0/3 | ¥2.80 |
| `preview/Kimi-K2.6` | 13/19 | 未記録 | 未記録 | 151.63秒 | 2/3 | ¥3.60 |
| `preview/Phi-4-multimodal-instruct` | 10/19 | 未記録 | 未記録 | **2.04秒** | 1/3 | ¥0.40 |
| `preview/Phi-4-mini-instruct-cpu` | 9/19 | 未記録 | 未記録 | 43.83秒 | 1/3 | ¥0.04 |
| `preview/Qwen3-VL-30B-A3B-Instruct` | 3/19 | 未記録 | 未記録 | 4.09秒 | 0/3 | ¥0.40 |
| `preview/Qwen3-0.6B-cpu` | 0/19 | 未記録 | 未記録 | 11.88秒 | 0/3 | ¥0.04 |
| `Qwen3-Coder-30B-A3B-Instruct` | 0/19 | 未記録 | 未記録 | 3.68秒 | 0/3 | ¥0.90 |
| `llm-jp-3.1-8x13b-instruct4` | 0/19 | 未記録 | 未記録 | 7.13秒 | 0/3 | ¥0.90 |

第 1 回は集計値のみ保存されており、並行処理テストと race detector の内訳、API の生レスポンス、生成コード、検証ログは残っていない。

### 第 2 回結果

| モデル | 機能 | 並行処理 | Race | 中央応答時間 | 初回から Modern | 価格 |
| :----- | :--: | :------: | :--: | :----------: | :------------: | :--: |
| `preview/Kimi-K2.6` | **19/19** | **6/6** | **3/3** | 101.10秒 | 1/3 | ¥3.60 |
| `preview/Qwen3.6-35B-A3B` | 18/19 | 5/6 | 2/3 | 50.14秒 | 1/3 | ¥1.80 |
| `Qwen3-Coder-480B-A35B-Instruct-FP8` | 17/19 | 5/6 | 1/3 | 8.51秒 | 0/3 | ¥2.80 |
| `preview/Phi-4-mini-instruct-cpu` | 8/19 | 1/6 | 1/3 | 44.94秒 | 1/3 | ¥0.04 |
| `preview/Phi-4-multimodal-instruct` | 7/19 | 0/6 | 1/3 | **2.04秒** | 1/3 | ¥0.40 |
| `gpt-oss-120b` | 7/19 | 0/6 | 1/3 | 8.91秒 | 1/3 | ¥0.90 |
| `preview/Qwen3-VL-30B-A3B-Instruct` | 3/19 | 0/6 | 0/3 | 4.09秒 | 0/3 | ¥0.40 |
| `preview/Qwen3-0.6B-cpu` | 0/19 | 0/6 | 0/3 | 10.17秒 | 0/3 | ¥0.04 |
| `Qwen3-Coder-30B-A3B-Instruct` | 0/19 | 0/6 | 0/3 | 3.64秒 | 0/3 | ¥0.90 |
| `llm-jp-3.1-8x13b-instruct4` | 0/19 | 0/6 | 0/3 | 7.73秒 | 0/3 | ¥0.90 |

### 第 3 回結果

| モデル | 機能 | 並行処理 | Race | 中央応答時間 | 初回から Modern | 価格 |
| :----- | :--: | :------: | :--: | :----------: | :------------: | :--: |
| `preview/Qwen3.6-35B-A3B` | **18/19** | **5/6** | **2/3** | 50.56秒 | 1/3 | ¥1.80 |
| `Qwen3-Coder-480B-A35B-Instruct-FP8` | 17/19 | **5/6** | 1/3 | 8.46秒 | 0/3 | ¥2.80 |
| `gpt-oss-120b` | 13/19 | 0/6 | **2/3** | 8.60秒 | 1/3 | ¥0.90 |
| `preview/Kimi-K2.6` | 13/19 | 0/6 | **2/3** | 103.59秒 | 0/3 | ¥3.60 |
| `preview/Phi-4-multimodal-instruct` | 10/19 | 3/6 | 1/3 | **2.04秒** | 1/3 | ¥0.40 |
| `preview/Phi-4-mini-instruct-cpu` | 9/19 | 2/6 | 1/3 | 37.39秒 | 1/3 | ¥0.04 |
| `preview/Qwen3-VL-30B-A3B-Instruct` | 3/19 | 0/6 | 0/3 | 4.06秒 | 0/3 | ¥0.40 |
| `preview/Qwen3-0.6B-cpu` | 0/19 | 0/6 | 0/3 | 10.28秒 | 0/3 | ¥0.04 |
| `Qwen3-Coder-30B-A3B-Instruct` | 0/19 | 0/6 | 0/3 | 4.04秒 | 0/3 | ¥0.90 |
| `llm-jp-3.1-8x13b-instruct4` | 0/19 | 0/6 | 0/3 | 7.75秒 | 0/3 | ¥0.90 |

Kimi の ParallelMapOrdered は API が `finish_reason: length` で終了し、回答本文が空だったため、機能・並行処理・Modern の採点対象となるコードを取得できなかった。Race の 2/3 は生成できた TopKFrequent と ParseOptions の結果である。

「初回から Modern」は、各 3 課題のうち、コンパイル可能かつ `go fix` でソースが変更されなかった件数を示す。

## Pareto front による候補選定

全評価軸を同時に比較した第 3 回の Pareto front は次の 6 モデルだった。

- `preview/Qwen3-0.6B-cpu`
- `preview/Phi-4-mini-instruct-cpu`
- `preview/Phi-4-multimodal-instruct`
- `gpt-oss-120b`
- `preview/Qwen3.6-35B-A3B`
- `Qwen3-Coder-480B-A35B-Instruct-FP8`

価格が低いだけのモデルも Pareto front に残り得るため、Functionality 80% 以上を品質ゲートに設定した。第 3 回の実用的な Pareto front は次の 2 モデルである。

- `preview/Qwen3.6-35B-A3B`
- `Qwen3-Coder-480B-A35B-Instruct-FP8`

Qwen3.6 は正確性と再現性、Qwen3-Coder-480B は応答速度に強みがある。第 2 回の品質 front に含まれた Kimi は、第 3 回では Functionality 80% の品質ゲートを下回った。

## 3 回の結果比較

| モデル | 第1回 | 第2回 | 第3回 | 合格数の幅 |
| :----- | :---: | :---: | :---: | :-------: |
| `preview/Qwen3-0.6B-cpu` | 0/19 | 0/19 | 0/19 | 0 |
| `preview/Phi-4-mini-instruct-cpu` | 9/19 | 8/19 | 9/19 | 1 |
| `preview/Phi-4-multimodal-instruct` | 10/19 | 7/19 | 10/19 | 3 |
| `preview/Qwen3-VL-30B-A3B-Instruct` | 3/19 | 3/19 | 3/19 | 0 |
| `gpt-oss-120b` | 18/19 | 7/19 | 13/19 | **11** |
| `Qwen3-Coder-30B-A3B-Instruct` | 0/19 | 0/19 | 0/19 | 0 |
| `llm-jp-3.1-8x13b-instruct4` | 0/19 | 0/19 | 0/19 | 0 |
| `preview/Qwen3.6-35B-A3B` | 18/19 | 18/19 | 18/19 | **0** |
| `Qwen3-Coder-480B-A35B-Instruct-FP8` | 17/19 | 17/19 | 17/19 | **0** |
| `preview/Kimi-K2.6` | 13/19 | 19/19 | 13/19 | **6** |

「合格数の幅」は、3 回の最大合格数と最小合格数の差である。Qwen3.6、Qwen3-Coder-480B、Qwen3-VL、Qwen3-0.6B、Qwen3-Coder-30B、llm-jp は 3 回とも同じ合格数だった。ただし、0/19 や 3/19 の安定は実用上の品質を意味しない。高品質かつ安定していたのは Qwen3.6 と Qwen3-Coder-480B の 2 モデルである。

Kimi と gpt-oss は引き続き大きく変動した。`temperature: 0` でも、推論基盤、内部 reasoning、出力上限、サービス側の状態などにより完全な決定性は得られていない。

第 1 回の生レスポンスは保存されていないため、比較には集計値を使用している。第 2 回と第 3 回は API レスポンスを含む完全な証跡を保存している。

## モデル別所見

### preview/Qwen3.6-35B-A3B

3 回とも 18/19 で、候補中もっとも再現性が高い。毎回、worker panic を要求どおりのエラーへ変換できない同じ欠陥が残った。第 3 回の応答時間中央値は 50.56 秒、出力は 21,140 トークンと重いが、AI Agent の主モデルとしては安定性が魅力である。

### Qwen3-Coder-480B-A35B-Instruct-FP8

3 回とも 17/19 で、第 3 回も中央値 8.46 秒と高品質群では最速だった。第 3 回は ParallelMapOrdered のキャンセル待機テストと ParseOptions の型付き cause テストを落とした。高速だが、並行処理とエラー伝播には慎重なレビューが必要である。

### preview/Kimi-K2.6

第 2 回では唯一 19/19 を達成したが、第 1 回と第 3 回は 13/19 だった。第 3 回の ParallelMapOrdered は 16,384 出力トークンの上限に達し、コードを取得できなかった。第 3 回は 3 課題で合計 30,340 出力トークン、応答時間中央値 103.59 秒を要しており、価格・待ち時間・出力安定性の負担が最も大きい。

### gpt-oss-120b

18 → 7 → 13 と推移した。第 3 回は TopKFrequent と ParseOptions が全テストを通過した一方、ParallelMapOrdered は未定義変数 `err` によりコンパイルできなかった。速度と価格は魅力的だが、3 標本でも品質分散が大きい。

### 小型・マルチモーダルモデル

`preview/Phi-4-multimodal-instruct` は 3 回とも応答時間中央値が約 2.04 秒と非常に高速だったが、複雑な並行処理には対応できなかった。

`preview/Phi-4-mini-instruct-cpu` は最低価格だが、応答時間中央値は 37.39〜44.94 秒で、価格以外の優位性は限定的だった。VL モデルと 0.6B モデルはコンパイルエラーが多く、Go 実装の主モデルには適さない。

## 制約と次の検証案

各調査では各モデル・各課題を 1 回だけ生成しているため、これは能力の確定順位ではなく、3 回の独立した初回成功率の比較である。Kimi と gpt-oss の変動が示すとおり、本番選定にはさらに多い反復試験が必要である。

次の段階では、Qwen3.6、Qwen3-Coder-480B、Kimi、gpt-oss の 4 モデルを対象に各課題を最低 5 回生成し、次を比較するのが望ましい。

- コンパイル成功率
- 全テスト成功率
- 並行処理テストごとの失敗分布
- median・p90 応答時間
- トークン数と実測費用
- `go fix` の変更回数と変更内容
- AI Agent に組み込んだ状態での end-to-end 成功率

## 再実行方法

AI Agent（検証では [Hermes Agent](https://hermes-agent.nousresearch.com/)）の credential pool または環境変数 `SAKURA_AI_TOKEN` から API トークンを解決する。トークン値は成果物へ保存していない。

```bash
PYTHON=/opt/homebrew/Cellar/hermes-agent/2026.6.5/libexec/bin/python
RESEARCH_RUN=3 "$PYTHON" research.py all
```

段階別にも実行できる。

```bash
RESEARCH_RUN=3 "$PYTHON" research.py generate
RESEARCH_RUN=3 "$PYTHON" research.py evaluate
RESEARCH_RUN=3 "$PYTHON" research.py repeat
RESEARCH_RUN=3 "$PYTHON" research.py summarize
```

`RESEARCH_RUN` の既定値は `2` で、第 2 回と互換性のある `responses/`・`logs/` を使用する。`RESEARCH_RUN=3` では `responses-run3/`・`logs-run3/`・`run3-summary.json` に分離して保存する。同じ調査番号に既存の `solution.go` がある場合はキャッシュとして再利用される。

## 検証ファイル一覧

### 総括・集計

- [README.md](README.md)：本調査の日本語総括レポートとファイル索引。
- [FINDINGS.md](FINDINGS.md)：第 2 回結果を中心にまとめた英語版の短い所見。
- [comparison.md](comparison.md)：3 回のモデル別合格数、応答時間、品質 Pareto front の比較表。
- [baseline_run1.json](baseline_run1.json)：第 1 回の保存済み集計値。API の生レスポンスは含まない。
- [run2-summary.json](run2-summary.json)：第 2 回の評価ベクトル、合格数、Pareto front。
- [run3-summary.json](run3-summary.json)：第 3 回の評価ベクトル、合格数、Pareto front。
- [research.py](research.py)：API 呼び出し、コード抽出、Go 検証、反復試験、集計を行う再実行ハーネス。

### 入力プロンプト

- [prompts/topk.md](prompts/topk.md)：TopKFrequent 課題の完全な user prompt。
- [prompts/pmap.md](prompts/pmap.md)：ParallelMapOrdered 課題の完全な user prompt。
- [prompts/options.md](prompts/options.md)：ParseOptions 課題の完全な user prompt。

### Go テスト fixture

- [fixtures/topk/go.mod](fixtures/topk/go.mod)：TopKFrequent 検証用の Go 1.26 module。
- [fixtures/topk/solution_test.go](fixtures/topk/solution_test.go)：TopKFrequent の機能テストと benchmark。
- [fixtures/pmap/go.mod](fixtures/pmap/go.mod)：ParallelMapOrdered 検証用の Go 1.26 module。
- [fixtures/pmap/solution_test.go](fixtures/pmap/solution_test.go)：順序、キャンセル、panic、leak、race を検査するテストと benchmark。
- [fixtures/options/go.mod](fixtures/options/go.mod)：ParseOptions 検証用の Go 1.26 module。
- [fixtures/options/solution_test.go](fixtures/options/solution_test.go)：パース、部分結果、型付きエラー、modern API を検査するテストと benchmark。

### 実行・検証ログ

- [logs/requests.json](logs/requests.json)：30 API リクエストのモデル、課題、HTTP status、経過時間、トークン使用量。
- [logs/evaluation.json](logs/evaluation.json)：各生成コードの build、test、race、vet、lint、`go fix`、benchmark 結果。
- [logs/repeat-summary.json](logs/repeat-summary.json)：コンパイル可能な並行処理回答を 20 回反復した結果の集計。
- [logs/repeat-Qwen3-Coder-30B-A3B-Instruct.log](logs/repeat-Qwen3-Coder-30B-A3B-Instruct.log)：Qwen3-Coder-30B の並行処理反復ログ。
- [logs/repeat-Qwen3-Coder-480B-A35B-Instruct-FP8.log](logs/repeat-Qwen3-Coder-480B-A35B-Instruct-FP8.log)：Qwen3-Coder-480B の並行処理反復ログ。
- [logs/repeat-preview__Qwen3.6-35B-A3B.log](logs/repeat-preview__Qwen3.6-35B-A3B.log)：Qwen3.6-35B の並行処理反復ログ。
- [logs/repeat-preview__Kimi-K2.6.log](logs/repeat-preview__Kimi-K2.6.log)：Kimi-K2.6 の並行処理反復ログ。
- [logs/repeat-preview__Phi-4-mini-instruct-cpu.log](logs/repeat-preview__Phi-4-mini-instruct-cpu.log)：Phi-4-mini の並行処理反復ログ。
- [logs/file-manifest.txt](logs/file-manifest.txt)：調査成果物のファイル一覧。
- [logs-run3/requests.json](logs-run3/requests.json)：第 3 回の 30 API リクエストの記録。
- [logs-run3/evaluation.json](logs-run3/evaluation.json)：第 3 回の build、test、race、vet、lint、`go fix`、benchmark 結果。
- [logs-run3/repeat-summary.json](logs-run3/repeat-summary.json)：第 3 回の並行処理回答を 20 回反復した結果の集計。

### モデル別 API 応答・生成コード

第 2 回は `responses/`、第 3 回は `responses-run3/` に保存している。各モデルのディレクトリには `topk/`、`pmap/`、`options/` があり、原則として API の生レスポンス `response.json` と、そこから抽出した未修正の `solution.go` を格納している。

| モデル | 第 2 回 | 第 3 回 |
| :----- | :------ | :------ |
| `preview/Qwen3-0.6B-cpu` | [成果物](responses/preview__Qwen3-0.6B-cpu/) | [成果物](responses-run3/preview__Qwen3-0.6B-cpu/) |
| `preview/Phi-4-mini-instruct-cpu` | [成果物](responses/preview__Phi-4-mini-instruct-cpu/) | [成果物](responses-run3/preview__Phi-4-mini-instruct-cpu/) |
| `preview/Phi-4-multimodal-instruct` | [成果物](responses/preview__Phi-4-multimodal-instruct/) | [成果物](responses-run3/preview__Phi-4-multimodal-instruct/) |
| `preview/Qwen3-VL-30B-A3B-Instruct` | [成果物](responses/preview__Qwen3-VL-30B-A3B-Instruct/) | [成果物](responses-run3/preview__Qwen3-VL-30B-A3B-Instruct/) |
| `gpt-oss-120b` | [成果物](responses/gpt-oss-120b/) | [成果物](responses-run3/gpt-oss-120b/) |
| `Qwen3-Coder-30B-A3B-Instruct` | [成果物](responses/Qwen3-Coder-30B-A3B-Instruct/) | [成果物](responses-run3/Qwen3-Coder-30B-A3B-Instruct/) |
| `llm-jp-3.1-8x13b-instruct4` | [成果物](responses/llm-jp-3.1-8x13b-instruct4/) | [成果物](responses-run3/llm-jp-3.1-8x13b-instruct4/) |
| `preview/Qwen3.6-35B-A3B` | [成果物](responses/preview__Qwen3.6-35B-A3B/) | [成果物](responses-run3/preview__Qwen3.6-35B-A3B/) |
| `Qwen3-Coder-480B-A35B-Instruct-FP8` | [成果物](responses/Qwen3-Coder-480B-A35B-Instruct-FP8/) | [成果物](responses-run3/Qwen3-Coder-480B-A35B-Instruct-FP8/) |
| `preview/Kimi-K2.6` | [成果物](responses/preview__Kimi-K2.6/) | [成果物](responses-run3/preview__Kimi-K2.6/) |

第 3 回の Kimi の ParallelMapOrdered は出力上限に達して本文が空だったため、`response.json` と `solution.go` の代わりに `error.json` を保存している。
