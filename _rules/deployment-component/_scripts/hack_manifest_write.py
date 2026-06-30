#!/usr/bin/env python3
"""
Deterministic tool: Write top-level iac/hack/manifest.yaml for a provider/kind.

Usage (stdin):
  cat manifest.yaml | python3 _rules/deployment-component/_scripts/hack_manifest_write.py --provider aws --kindfolder awscloudfront --stdin

Usage (file):
  python3 _rules/deployment-component/_scripts/hack_manifest_write.py --provider aws --kindfolder awscloudfront --content-file /tmp/manifest.yaml

Outputs JSON:
  - wrote: bool
  - path: absolute file path
  - relative_path: repo-relative path
  - bytes_written: int
  - created_dirs: list
  - sha256: string
  - error: optional string
"""

import argparse
import hashlib
import json
import os
import sys
from typing import List, Tuple


def find_repo_root(start_dir: str) -> str:
    current = os.path.abspath(start_dir)
    while True:
        if os.path.isdir(os.path.join(current, ".git")) or os.path.isfile(os.path.join(current, "go.mod")):
            return current
        parent = os.path.dirname(current)
        if parent == current:
            return start_dir
        current = parent


def build_manifest_path(repo_root: str, provider: str, kind_folder: str) -> Tuple[str, str]:
    relative_path = os.path.join(
        "apis",
        "dev",
        "planton",
        "provider",
        provider,
        kind_folder,
        "v1",
        "iac",
        "hack",
        "manifest.yaml",
    )
    absolute_path = os.path.join(repo_root, relative_path)
    return absolute_path, relative_path


def ensure_parent_dirs(file_path: str) -> List[str]:
    created: List[str] = []
    parent = os.path.dirname(file_path)
    if not os.path.isdir(parent):
        os.makedirs(parent, exist_ok=True)
        created.append(parent)
    return created


def normalize_segment(segment: str) -> str:
    normalized = segment.strip().lower().replace("_", "")
    if ".." in normalized or normalized.startswith("/") or normalized.startswith("~"):
        raise ValueError("Invalid segment: path traversal not allowed")
    return normalized


def main() -> int:
    parser = argparse.ArgumentParser(description="Write hack manifest deterministically")
    parser.add_argument("--provider", required=True, help="Provider key (e.g., aws, gcp, azure)")
    parser.add_argument("--kindfolder", required=True, help="Kind folder (lowercase, no underscores)")
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("--stdin", action="store_true", help="Read manifest content from STDIN")
    group.add_argument("--content-file", help="Path to a file containing manifest content")
    args = parser.parse_args()

    try:
        provider = normalize_segment(args.provider)
        kind_folder = normalize_segment(args.kindfolder)
    except Exception as exc:
        print(json.dumps({"wrote": False, "error": f"invalid inputs: {exc}"}))
        return 2

    try:
        if args.stdin:
            content = sys.stdin.read()
        else:
            with open(os.path.abspath(args.content_file), "r", encoding="utf-8") as f:
                content = f.read()
    except Exception as exc:
        print(json.dumps({"wrote": False, "error": f"failed to read content: {exc}"}))
        return 3

    repo_root = os.environ.get("REPO_ROOT", find_repo_root(os.getcwd()))
    abs_path, rel_path = build_manifest_path(repo_root, provider, kind_folder)

    result = {
        "wrote": False,
        "path": abs_path,
        "relative_path": rel_path,
        "bytes_written": 0,
        "created_dirs": [],
        "sha256": "",
    }

    try:
        result["created_dirs"] = ensure_parent_dirs(abs_path)
        with open(abs_path, "w", encoding="utf-8", newline="\n") as f:
            f.write(content)
        result["bytes_written"] = len(content.encode("utf-8"))
        result["wrote"] = True
        result["sha256"] = hashlib.sha256(content.encode("utf-8")).hexdigest()
        print(json.dumps(result))
        return 0
    except Exception as exc:
        result["error"] = str(exc)
        print(json.dumps(result))
        return 1


if __name__ == "__main__":
    sys.exit(main())


