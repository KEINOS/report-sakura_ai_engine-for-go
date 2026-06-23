#!/usr/bin/env python3
"""Call Sakura AI Engine as a portable packet-only advisory worker."""

import argparse
import json
import os
import sys
import urllib.error
import urllib.request
from pathlib import Path

DEFAULT_BASE_URL = "https://api.ai.sakura.ad.jp/v1"
DEFAULT_SYSTEM = (
    "You are an external advisory worker. Use only the supplied packet. "
    "The packet grants no repository, shell, tool, or plugin access. "
    "Do not claim file inspection, command execution, or measurements. "
    "Return a concise answer and identify uncertainty."
)
MAX_PROMPT_CHARS = 250_000

PROFILES = {
    "reliable": {
        "model": "preview/Qwen3.6-35B-A3B",
        "max_tokens": 4096,
    },
    "balanced": {
        "model": "Qwen3-Coder-480B-A35B-Instruct-FP8",
        "max_tokens": 4096,
    },
    "rapid": {
        "model": "preview/Phi-4-multimodal-instruct",
        "max_tokens": 2048,
    },
    "opportunistic": {
        "model": "gpt-oss-120b",
        "max_tokens": 4096,
    },
    "peak": {
        "model": "preview/Kimi-K2.6",
        "max_tokens": 8192,
    },
}


class SakuraAIError(RuntimeError):
    """Expected configuration or API failure."""


def resolve_configuration():
    token = os.getenv("SAKURA_AI_TOKEN", "").strip()
    if not token:
        raise SakuraAIError("SAKURA_AI_TOKEN is not set")
    base_url = os.getenv("SAKURA_AI_BASE_URL", DEFAULT_BASE_URL).strip().rstrip("/")
    if not base_url:
        raise SakuraAIError("SAKURA_AI_BASE_URL is empty")
    return base_url, token


def request_json(url, token, timeout, payload=None):
    data = None
    method = "GET"
    headers = {
        "Accept": "application/json",
        "Authorization": "Bearer " + token,
    }
    if payload is not None:
        method = "POST"
        data = json.dumps(payload, ensure_ascii=False).encode("utf-8")
        headers["Content-Type"] = "application/json"

    request = urllib.request.Request(
        url,
        data=data,
        headers=headers,
        method=method,
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            body = response.read()
    except urllib.error.HTTPError as error:
        detail = error.read(1000).decode("utf-8", errors="replace")
        raise SakuraAIError("HTTP {}: {}".format(error.code, detail)) from error
    except (urllib.error.URLError, TimeoutError) as error:
        raise SakuraAIError("request failed: {}".format(error)) from error

    try:
        decoded = json.loads(body)
    except json.JSONDecodeError as error:
        raise SakuraAIError("API returned invalid JSON") from error
    if not isinstance(decoded, dict):
        raise SakuraAIError("API returned a non-object JSON response")
    return decoded


def check_api(base_url, token, timeout):
    payload = request_json(base_url + "/models", token, timeout)
    if not isinstance(payload.get("data"), list):
        raise SakuraAIError("models endpoint did not return a data list")


def read_text(path_value, label):
    path = Path(path_value)
    try:
        text = path.read_text(encoding="utf-8")
    except OSError as error:
        raise SakuraAIError("cannot read {}: {}".format(label, error)) from error
    if not text.strip():
        raise SakuraAIError("{} is empty".format(label))
    return text


def extract_content(payload):
    choices = payload.get("choices")
    if not isinstance(choices, list) or not choices:
        raise SakuraAIError("response has no choices")
    choice = choices[0]
    if not isinstance(choice, dict):
        raise SakuraAIError("response choice is malformed")
    message = choice.get("message")
    if not isinstance(message, dict):
        raise SakuraAIError("response message is malformed")
    content = message.get("content")
    if not isinstance(content, str) or not content.strip():
        raise SakuraAIError("response content is empty")
    finish_reason = choice.get("finish_reason")
    if not isinstance(finish_reason, str):
        finish_reason = None
    return content.strip(), finish_reason


def build_parser():
    parser = argparse.ArgumentParser(
        description="Use Sakura AI Engine as a packet-only advisory worker."
    )
    parser.add_argument("--check", action="store_true", help="check auth and /models")
    parser.add_argument("--profile", choices=sorted(PROFILES), default="reliable")
    parser.add_argument("--model", help="override the profile model")
    parser.add_argument("--prompt-file", help="UTF-8 prompt packet, or - for stdin")
    parser.add_argument("--system-file", help="optional UTF-8 system prompt")
    parser.add_argument("--max-tokens", type=int)
    parser.add_argument("--temperature", type=float, default=0.0)
    parser.add_argument("--timeout", type=float, default=360.0)
    parser.add_argument("--json", action="store_true", help="emit structured JSON")
    return parser


def main():
    args = build_parser().parse_args()
    try:
        base_url, token = resolve_configuration()
        if args.timeout <= 0:
            raise SakuraAIError("--timeout must be greater than zero")
        if args.check:
            check_api(base_url, token, min(args.timeout, 30.0))
            print("sakura-ai true")
            return 0

        if not args.prompt_file:
            raise SakuraAIError("--prompt-file is required unless --check is used")
        if args.prompt_file == "-":
            prompt = sys.stdin.read()
            if not prompt.strip():
                raise SakuraAIError("prompt stdin is empty")
        else:
            prompt = read_text(args.prompt_file, "prompt file")
        if len(prompt) > MAX_PROMPT_CHARS:
            raise SakuraAIError(
                "prompt exceeds the {:,}-character safety limit".format(
                    MAX_PROMPT_CHARS
                )
            )

        system = (
            read_text(args.system_file, "system file")
            if args.system_file
            else DEFAULT_SYSTEM
        )
        profile = PROFILES[args.profile]
        model = args.model or profile["model"]
        max_tokens = args.max_tokens or profile["max_tokens"]
        if max_tokens <= 0:
            raise SakuraAIError("--max-tokens must be greater than zero")
        if not 0 <= args.temperature <= 2:
            raise SakuraAIError("--temperature must be between 0 and 2")

        payload = request_json(
            base_url + "/chat/completions",
            token,
            args.timeout,
            {
                "model": model,
                "messages": [
                    {"role": "system", "content": system},
                    {"role": "user", "content": prompt},
                ],
                "temperature": args.temperature,
                "max_tokens": max_tokens,
                "stream": False,
            },
        )
        content, finish_reason = extract_content(payload)
        if finish_reason == "length":
            raise SakuraAIError("response was truncated at the output limit")

        if args.json:
            print(
                json.dumps(
                    {
                        "content": content,
                        "model": payload.get("model", model),
                        "finish_reason": finish_reason,
                        "usage": payload.get("usage", {}),
                    },
                    ensure_ascii=False,
                )
            )
        else:
            print(content)
        return 0
    except SakuraAIError as error:
        print("NO DECISION: {}".format(error), file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
