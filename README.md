---
title: "Sakura.AI Go コーディングモデル比較調査"
description: "A benchmark-based comparison of Go coding models on Sakura.AI OpenAI-compatible endpoints."
date: 2026-06-21
lastmod: 2026-06-21
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

# Sakura AI Engine - AI モデルの Go コーディング能力比較調査

## 概要

AI Agent で利用する Go コーディング向け学習モデル（以下モデル）を選定する候補の 1 つとして、**[さくらの AI Engine](https://manual.sakura.ad.jp/cloud/manual-ai-engine.html)（以下 Sakura AI）で提供されているモデルの性能を比較・評価した**。

純粋なモデル性能を比較するため、AI Agent/Harness の会話履歴・メモリー・自己学習コンテキストを介さないように、Sakura AI の OpenAI API 互換エンドポイントへ各モデルに直接リクエストし、初回 Go コード生成の性能を比較した。

評価対象は 10 モデル、課題は 3 種類である。各モデルに同一の system prompt、user prompt、`temperature: 0` を与え、生成された初回回答を人手で修正せずに Go 1.26.4 環境で 2 回にわたり検証した。課題の作成と調査の実施には Codex を利用している。

このリポジトリには、第 2 回調査の 30 件すべての API 応答、抽出コード、検証ログをこのディレクトリに保存している。同日に行った初回調査のデータは筆者のポカにより保存していないが、集計値は `baseline_run1.json` に保存している。

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
- [第2回結果](#第2回結果)
- [Pareto comparison](#pareto-comparison)
- [第1回との比較](#第1回との比較)
- [モデル別所見](#モデル別所見)
- [制約と次の検証案](#制約と次の検証案)
- [再実行方法](#再実行方法)
- [検証ファイル一覧](#検証ファイル一覧)

## 結論

用途別の推奨は次のとおり。

- 最高の測定正確性：`preview/Kimi-K2.6`
- 2 回の調査を通した安定性：`preview/Qwen3.6-35B-A3B`
- 高品質モデル内での応答速度：`Qwen3-Coder-480B-A35B-Instruct-FP8`
- 低価格・超高速の補助用途：`preview/Phi-4-multimodal-instruct`

第 2 回の単発測定で最も高い正確性を示したのは `preview/Kimi-K2.6` であり、機能テスト19/19、並行処理テスト 6/6、race detector 3/3を通過した。並行処理コードを 20 回繰り返した race test もすべて成功している。

一方、2 回の調査を通じて最も安定していた高品質モデルは `preview/Qwen3.6-35B-A3B` だった。両方の調査で 18/19 を記録しており、再現性を重視するなら現時点で最も堅実な候補である。

第 1 回の検証で推奨された `gpt-oss-120b` は、18/19 から第 2 回は 7/19 まで低下した。第 2 回では並行処理とオプション解析の回答がコンパイルできず、単発結果の振れが大きい。

そのため、現時点では AI Agent の Go コーディング・モデルの既定モデルとして即決するには再現性の根拠が足りない。

## 評価対象

| モデル | 公称価格（入力と出力の合計料金） |
| :----- | :------: |
| `preview/Qwen3-0.6B-cpu` | ¥0.04 |
| `preview/Phi-4-mini-instruct-cpu` | ¥0.04 |
| `preview/Phi-4-multimodal-instruct` | ¥0.40 |
| `preview/Qwen3-VL-30B-A3B-Instruct` | ¥0.40 |
| `gpt-oss-120b` | ¥0.90 |
| `Qwen3-Coder-30B-A3B-Instruct` | ¥0.90 |
| `llm-jp-3.1-8x13b-instruct4` | ¥0.90 |
| `preview/Qwen3.6-35B-A3B` | ¥1.80 |
| `Qwen3-Coder-480B-A35B-Instruct-FP8` | ¥2.80 |
| `preview/Kimi-K2.6` | ¥3.60 |

> 価格は検証日（2026/06/21）時点の入力・出力の合計料金を基にしている。検証は、[基盤モデル無償プラン](https://ai.sakura.ad.jp/sakura-ai/ai-engine/)の無料枠の 3,000 回/月のうち 80 程度を消費した。従量制課金の場合、実際の請求額は各リクエストの token 数に依存するので注意。

## 評価課題

### TopKFrequent

単語の出現頻度を集計し、頻度降順・辞書順で上位 k 件を返す課題。正確性に加えて、大規模入力かつ小さな k に対するアルゴリズム効率、入力非破壊、非 nil 空スライスを検証した。

### ParallelMapOrdered

順序を維持する並行 map 処理。worker 数制限、親 context のキャンセル、最初のエラーの保持、panic のエラー化、goroutine leak、race の有無を検証した。本調査で Reliability を最も強く判別した課題である。

### ParseOptions

文字列オプションを解析する課題。部分結果、型付きエラー、`errors.Is`、デフォルト値、重複ラベル、Go 1.26 の modern syntax・API を検証した。

## 評価方法

次の観測値を評価ベクトルへ対応させ、[Pareto front](https://en.wikipedia.org/wiki/Pareto_front)（パレート解の最適バランスとなる境界線）を求めた。

| 評価軸 | 主な観測方法 |
| :----- | :----------- |
| Functionality | 3 課題、合計 19 件のブラックボックス・テスト |
| Effectiveness | テスト合格、生成行数、benchmark の実行時間・allocation |
| Reliability | 並行処理 6 テスト、race detector、20 回反復試験 |
| Usability | `gofmt`、`go vet`、`golangci-lint`、生成行数 |
| Responsiveness | API リクエスト開始から完了までの経過秒数の中央値 |
| Modern Syntax | 初回回答へ `go fix ./...` を適用した際の変更有無 |
| Cost | 測定時の入力と出力の合計料金と、API usage の token 数 |

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

## 第2回結果

| モデル | 機能 | 並行処理 | Race | 中央応答時間 | 初回から Modern | 価格 |
| :----- | :--: | :------: | :--: | :----------: | :------------: | :--: |
| `preview/Kimi-K2.6` | **19/19** | **6/6** | **3/3** | 101.10秒 | 1/3 | ¥3.60 |
| `preview/Qwen3.6-35B-A3B` | 18/19 | 5/6 | 2/3 | 50.14秒 | 1/3 | ¥1.80 |
| `Qwen3-Coder-480B-A35B-Instruct-FP8` | 17/19 | 5/6 | 1/3 | **8.51秒** | 0/3 | ¥2.80 |
| `preview/Phi-4-mini-instruct-cpu` | 8/19 | 1/6 | 1/3 | 44.94秒 | 1/3 | ¥0.04 |
| `preview/Phi-4-multimodal-instruct` | 7/19 | 0/6 | 1/3 | **2.04秒** | 1/3 | ¥0.40 |
| `gpt-oss-120b` | 7/19 | 0/6 | 1/3 | 8.91秒 | 1/3 | ¥0.90 |
| `preview/Qwen3-VL-30B-A3B-Instruct` | 3/19 | 0/6 | 0/3 | 4.09秒 | 0/3 | ¥0.40 |
| `preview/Qwen3-0.6B-cpu` | 0/19 | 0/6 | 0/3 | 10.17秒 | 0/3 | ¥0.04 |
| `Qwen3-Coder-30B-A3B-Instruct` | 0/19 | 0/6 | 0/3 | 3.64秒 | 0/3 | ¥0.90 |
| `llm-jp-3.1-8x13b-instruct4` | 0/19 | 0/6 | 0/3 | 7.73秒 | 0/3 | ¥0.90 |

「初回から Modern」は、各 3 課題のうち、コンパイル可能かつ `go fix` でソースが変更されなかった件数を示す。

## Pareto comparison

全評価軸を同時に比較した第 2 回の Pareto front は次の 6 モデルだった。

- `preview/Qwen3-0.6B-cpu`
- `preview/Phi-4-mini-instruct-cpu`
- `preview/Phi-4-multimodal-instruct`
- `preview/Qwen3.6-35B-A3B`
- `Qwen3-Coder-480B-A35B-Instruct-FP8`
- `preview/Kimi-K2.6`

低価格だけが突出したモデルも非劣解になり得るため、Functionality 80% 以上を品質ゲートにした実用的な Pareto front は次の 3 モデルとなった。

- `preview/Qwen3.6-35B-A3B`
- `Qwen3-Coder-480B-A35B-Instruct-FP8`
- `preview/Kimi-K2.6`

3 モデルの性質は明確に異なる。Kimi は正確性、Qwen3.6 は 2 回を通した安定性、Qwen3-Coder-480B は応答速度に強みがある。

## 第1回との比較

| モデル | 第1回 | 第2回 | 合格数差 | 応答時間差 |
| :----- | :---: | :---: | :------: | :--------: |
| `preview/Qwen3-0.6B-cpu` | 0/19 | 0/19 | 0 | -1.70秒 |
| `preview/Phi-4-mini-instruct-cpu` | 9/19 | 8/19 | -1 | +1.12秒 |
| `preview/Phi-4-multimodal-instruct` | 10/19 | 7/19 | -3 | ±0.00秒 |
| `preview/Qwen3-VL-30B-A3B-Instruct` | 3/19 | 3/19 | 0 | ±0.00秒 |
| `gpt-oss-120b` | 18/19 | 7/19 | **-11** | +0.03秒 |
| `Qwen3-Coder-30B-A3B-Instruct` | 0/19 | 0/19 | 0 | -0.04秒 |
| `llm-jp-3.1-8x13b-instruct4` | 0/19 | 0/19 | 0 | +0.60秒 |
| `preview/Qwen3.6-35B-A3B` | 18/19 | 18/19 | 0 | -0.54秒 |
| `Qwen3-Coder-480B-A35B-Instruct-FP8` | 17/19 | 17/19 | 0 | -0.05秒 |
| `preview/Kimi-K2.6` | 13/19 | 19/19 | **+6** | -50.53秒 |

第1回と第2回で大きく変動したのは Kimi と gpt-oss である。`temperature: 0` でも、推論基盤、内部 reasoning、サービス側の状態などにより完全な決定性は得られていない。

第 1 回の raw response は前回の一時ディレクトリ削除とともに失われたため、比較には保存していた集計値を使用している。第 2 回は raw response を含む完全な証跡を保存している。

## モデル別所見

### preview/Kimi-K2.6

第 2 回では唯一、19/19 を達成し、20 回反復した並行処理 race test も全成功した。ただし 3 課題で 26,895 completion tokens を消費し、応答時間中央値は 101.10 秒だった。品質最優先の候補だが、価格と待ち時間の負担は最も大きい。

### preview/Qwen3.6-35B-A3B

2 回とも 18/19 で、今回の候補中もっとも再現性が高い。毎回、worker panic を要求どおりのエラーへ変換できない同じ欠陥が残った。中央値 50.14 秒、completion tokens は 21,140 と重いが、AI Agent の Go コーディングの主モデル候補としては安定性が魅力である。

### Qwen3-Coder-480B-A35B-Instruct-FP8

2 回とも 17/19 で、中央値 8.51 秒と高品質群では最速だった。最初の worker error より `context canceled` を優先して返す問題が 20 回すべてで再現し、panic test も 20 回中 2 回失敗した。高速だが、並行処理のエラー処理には慎重なレビューが必要である。

### gpt-oss-120b

第 1 回は 18/19 だったが、第 2 回は並行処理コードに未定義変数、options コードに戻り値型の誤りがあり 7/19 となった。速度と価格は魅力的だが、今回の 2 標本では品質分散が大きい。

### 小型・マルチモーダルモデル

`preview/Phi-4-multimodal-instruct` は 2.04 秒と非常に高速だが、複雑な並行処理には対応できなかった。

`preview/Phi-4-mini-instruct-cpu` は最低価格だが 44.94 秒を要し、価格以外の優位性は限定的だった。VL モデルと 0.6B モデルはコンパイルエラーが多く、Go 実装の主モデルには適さない。

## 制約と次の検証案

各モデル・各課題を 1 回だけ生成した結果なので、これは能力の確定順位ではなく、観測された初回成功率の比較である。Kimi と gpt-oss の変動が示すとおり、本番選定には反復試験が必要である。

次の段階では、上位 4 モデルを対象に各課題を最低 5 回生成し、次を比較するのが望ましい。

- コンパイル成功率
- 全テスト成功率
- 並行処理テストごとの失敗分布
- median・p90 応答時間
- token 数と実測費用
- `go fix` の変更回数と変更内容
- AI Agent に組み込んだ状態での end-to-end 成功率

## 再実行方法

AI Agent（検証では [Hermes Agent](https://hermes-agent.nousresearch.com/)）の credential pool または `SAKURA_AI_TOKEN` の環境変数から API token を解決する。token 値は成果物へ保存していない。

```bash
cd .temp_ai_research
/opt/homebrew/Cellar/hermes-agent/2026.6.5/libexec/bin/python research.py all
```

段階別にも実行できる。

```bash
python research.py generate
python research.py evaluate
python research.py repeat
python research.py summarize
```

既存の `solution.go` がある生成結果はキャッシュとして再利用される。完全に再生成する場合は、対象モデルの `responses/` 配下を別途退避または削除してから `generate` を実行する。

## 検証ファイル一覧

### 総括・集計

- [README.md](README.md)：本調査の日本語総括レポートとファイルINDEX。
- [FINDINGS.md](FINDINGS.md)：第 2 回結果を中心にまとめた英語版の短い所見。
- [comparison.md](comparison.md)：第 1 回と第 2 回のモデル別差分表。
- [baseline_run1.json](baseline_run1.json)：第 1 回の保存済み集計値。raw responseは含まない。
- [run2-summary.json](run2-summary.json)：第 2 回の評価ベクトル、合格数、Pareto front。
- [research.py](research.py)：API 呼び出し、コード抽出、Go 検証、反復試験、集計を行う再実行ハーネス。

### 入力プロンプト

- [prompts/topk.md](prompts/topk.md)：TopKFrequent 課題の完全なuser prompt。
- [prompts/pmap.md](prompts/pmap.md)：ParallelMapOrdered 課題の完全なuser prompt。
- [prompts/options.md](prompts/options.md)：ParseOptions 課題の完全なuser prompt。

### Goテストfixture

- [fixtures/topk/go.mod](fixtures/topk/go.mod)：TopKFrequent 検証用の Go 1.26 module。
- [fixtures/topk/solution_test.go](fixtures/topk/solution_test.go)：TopKFrequent の機能テストと benchmark。
- [fixtures/pmap/go.mod](fixtures/pmap/go.mod)：ParallelMapOrdered 検証用の Go 1.26 module。
- [fixtures/pmap/solution_test.go](fixtures/pmap/solution_test.go)：順序、キャンセル、panic、leak、race を検査するテストと benchmark。
- [fixtures/options/go.mod](fixtures/options/go.mod)：ParseOptions 検証用の Go 1.26 module。
- [fixtures/options/solution_test.go](fixtures/options/solution_test.go)：パース、部分結果、型付きエラー、modern API を検査するテストと benchmark。

### 実行・検証ログ

- [logs/requests.json](logs/requests.json)：30 API リクエストのモデル、課題、HTTP status、経過時間、token usage。
- [logs/evaluation.json](logs/evaluation.json)：各生成コードの build、test、race、vet、lint、`go fix`、benchmark 結果。
- [logs/repeat-summary.json](logs/repeat-summary.json)：コンパイル可能な並行処理回答を 20 回反復した結果の集計。
- [logs/repeat-Qwen3-Coder-30B-A3B-Instruct.log](logs/repeat-Qwen3-Coder-30B-A3B-Instruct.log)：Qwen3-Coder-30B の並行処理反復ログ。
- [logs/repeat-Qwen3-Coder-480B-A35B-Instruct-FP8.log](logs/repeat-Qwen3-Coder-480B-A35B-Instruct-FP8.log)：Qwen3-Coder-480B の並行処理反復ログ。
- [logs/repeat-preview__Qwen3.6-35B-A3B.log](logs/repeat-preview__Qwen3.6-35B-A3B.log)：Qwen3.6-35B の並行処理反復ログ。
- [logs/repeat-preview__Kimi-K2.6.log](logs/repeat-preview__Kimi-K2.6.log)：Kimi-K2.6 の並行処理反復ログ。
- [logs/repeat-preview__Phi-4-mini-instruct-cpu.log](logs/repeat-preview__Phi-4-mini-instruct-cpu.log)：Phi-4-mini の並行処理反復ログ。
- [logs/file-manifest.txt](logs/file-manifest.txt)：調査成果物のファイル一覧。

### モデル別API応答・生成コード

各モデルには 3 課題それぞれの `response.json` と、そこから抽出した未修正の `solution.go` がある。`response.json` は API の raw response、`solution.go` は Go 検証へ投入したコードである。

#### preview/Qwen3-0.6B-cpu

- TopKFrequent：[API応答](responses/preview__Qwen3-0.6B-cpu/topk/response.json)・[生成コード](responses/preview__Qwen3-0.6B-cpu/topk/solution.go)
- ParallelMapOrdered：[API応答](responses/preview__Qwen3-0.6B-cpu/pmap/response.json)・[生成コード](responses/preview__Qwen3-0.6B-cpu/pmap/solution.go)
- ParseOptions：[API応答](responses/preview__Qwen3-0.6B-cpu/options/response.json)・[生成コード](responses/preview__Qwen3-0.6B-cpu/options/solution.go)

#### preview/Phi-4-mini-instruct-cpu

- TopKFrequent：[API応答](responses/preview__Phi-4-mini-instruct-cpu/topk/response.json)・[生成コード](responses/preview__Phi-4-mini-instruct-cpu/topk/solution.go)
- ParallelMapOrdered：[API応答](responses/preview__Phi-4-mini-instruct-cpu/pmap/response.json)・[生成コード](responses/preview__Phi-4-mini-instruct-cpu/pmap/solution.go)
- ParseOptions：[API応答](responses/preview__Phi-4-mini-instruct-cpu/options/response.json)・[生成コード](responses/preview__Phi-4-mini-instruct-cpu/options/solution.go)

#### preview/Phi-4-multimodal-instruct

- TopKFrequent：[API応答](responses/preview__Phi-4-multimodal-instruct/topk/response.json)・[生成コード](responses/preview__Phi-4-multimodal-instruct/topk/solution.go)
- ParallelMapOrdered：[API応答](responses/preview__Phi-4-multimodal-instruct/pmap/response.json)・[生成コード](responses/preview__Phi-4-multimodal-instruct/pmap/solution.go)
- ParseOptions：[API応答](responses/preview__Phi-4-multimodal-instruct/options/response.json)・[生成コード](responses/preview__Phi-4-multimodal-instruct/options/solution.go)

#### preview/Qwen3-VL-30B-A3B-Instruct

- TopKFrequent：[API応答](responses/preview__Qwen3-VL-30B-A3B-Instruct/topk/response.json)・[生成コード](responses/preview__Qwen3-VL-30B-A3B-Instruct/topk/solution.go)
- ParallelMapOrdered：[API応答](responses/preview__Qwen3-VL-30B-A3B-Instruct/pmap/response.json)・[生成コード](responses/preview__Qwen3-VL-30B-A3B-Instruct/pmap/solution.go)
- ParseOptions：[API応答](responses/preview__Qwen3-VL-30B-A3B-Instruct/options/response.json)・[生成コード](responses/preview__Qwen3-VL-30B-A3B-Instruct/options/solution.go)

#### 生成物：gpt-oss-120b

- TopKFrequent：[API応答](responses/gpt-oss-120b/topk/response.json)・[生成コード](responses/gpt-oss-120b/topk/solution.go)
- ParallelMapOrdered：[API応答](responses/gpt-oss-120b/pmap/response.json)・[生成コード](responses/gpt-oss-120b/pmap/solution.go)
- ParseOptions：[API応答](responses/gpt-oss-120b/options/response.json)・[生成コード](responses/gpt-oss-120b/options/solution.go)

#### Qwen3-Coder-30B-A3B-Instruct

- TopKFrequent：[API応答](responses/Qwen3-Coder-30B-A3B-Instruct/topk/response.json)・[生成コード](responses/Qwen3-Coder-30B-A3B-Instruct/topk/solution.go)
- ParallelMapOrdered：[API応答](responses/Qwen3-Coder-30B-A3B-Instruct/pmap/response.json)・[生成コード](responses/Qwen3-Coder-30B-A3B-Instruct/pmap/solution.go)
- ParseOptions：[API応答](responses/Qwen3-Coder-30B-A3B-Instruct/options/response.json)・[生成コード](responses/Qwen3-Coder-30B-A3B-Instruct/options/solution.go)

#### llm-jp-3.1-8x13b-instruct4

- TopKFrequent：[API応答](responses/llm-jp-3.1-8x13b-instruct4/topk/response.json)・[生成コード](responses/llm-jp-3.1-8x13b-instruct4/topk/solution.go)
- ParallelMapOrdered：[API応答](responses/llm-jp-3.1-8x13b-instruct4/pmap/response.json)・[生成コード](responses/llm-jp-3.1-8x13b-instruct4/pmap/solution.go)
- ParseOptions：[API応答](responses/llm-jp-3.1-8x13b-instruct4/options/response.json)・[生成コード](responses/llm-jp-3.1-8x13b-instruct4/options/solution.go)

#### 生成物：preview/Qwen3.6-35B-A3B

- TopKFrequent：[API応答](responses/preview__Qwen3.6-35B-A3B/topk/response.json)・[生成コード](responses/preview__Qwen3.6-35B-A3B/topk/solution.go)
- ParallelMapOrdered：[API応答](responses/preview__Qwen3.6-35B-A3B/pmap/response.json)・[生成コード](responses/preview__Qwen3.6-35B-A3B/pmap/solution.go)
- ParseOptions：[API応答](responses/preview__Qwen3.6-35B-A3B/options/response.json)・[生成コード](responses/preview__Qwen3.6-35B-A3B/options/solution.go)

#### 生成物：Qwen3-Coder-480B-A35B-Instruct-FP8

- TopKFrequent：[API応答](responses/Qwen3-Coder-480B-A35B-Instruct-FP8/topk/response.json)・[生成コード](responses/Qwen3-Coder-480B-A35B-Instruct-FP8/topk/solution.go)
- ParallelMapOrdered：[API応答](responses/Qwen3-Coder-480B-A35B-Instruct-FP8/pmap/response.json)・[生成コード](responses/Qwen3-Coder-480B-A35B-Instruct-FP8/pmap/solution.go)
- ParseOptions：[API応答](responses/Qwen3-Coder-480B-A35B-Instruct-FP8/options/response.json)・[生成コード](responses/Qwen3-Coder-480B-A35B-Instruct-FP8/options/solution.go)

#### 生成物：preview/Kimi-K2.6

- TopKFrequent：[API応答](responses/preview__Kimi-K2.6/topk/response.json)・[生成コード](responses/preview__Kimi-K2.6/topk/solution.go)
- ParallelMapOrdered：[API応答](responses/preview__Kimi-K2.6/pmap/response.json)・[生成コード](responses/preview__Kimi-K2.6/pmap/solution.go)
- ParseOptions：[API応答](responses/preview__Kimi-K2.6/options/response.json)・[生成コード](responses/preview__Kimi-K2.6/options/solution.go)
