#!/usr/bin/env python3
"""
Deterministic tool: Read existing spec.proto content for a given provider and kind folder.

Usage examples:
  python3 _rules/deployment-component/_scripts/spec_proto_reader.py --provider aws --kindfolder awslambda

Outputs a JSON object to stdout with keys:
  - exists: bool
  - path: absolute file path where spec.proto is expected
  - relative_path: repo-relative path
  - content: string (empty if not exists)
  - error: optional error message
"""

import argparse
import json
import os
import sys
from typing import Tuple


def find_repo_root(start_dir: str) -> str:
    """Returns the repository root (dir containing .git or go.mod) or start_dir if not found."""
    current = os.path.abspath(start_dir)
    while True:
        git_dir = os.path.join(current, ".git")
        go_mod = os.path.join(current, "go.mod")
        if os.path.isdir(git_dir) or os.path.isfile(go_mod):
            return current
        parent = os.path.dirname(current)
        if parent == current:
            return start_dir
        current = parent


def build_spec_proto_path(repo_root: str, provider: str, kind_folder: str) -> Tuple[str, str]:
    relative_path = os.path.join(
        "apis",
        "org",
        "planton",
        "provider",
        provider,
        kind_folder,
        "v1",
        "spec.proto",
    )
    absolute_path = os.path.join(repo_root, relative_path)
    return absolute_path, relative_path


def normalize_segment(segment: str) -> str:
    # Keep lowercase letters, digits, and hyphens; convert underscores to nothing to match folder style
    normalized = segment.strip().lower().replace("_", "")
    # Disallow path traversal
    if ".." in normalized or normalized.startswith("/") or normalized.startswith("~"):
        raise ValueError("Invalid segment: path traversal not allowed")
    return normalized


def main() -> int:
    parser = argparse.ArgumentParser(description="Read existing spec.proto for provider/kind folder")
    parser.add_argument("--provider", required=True, help="Provider key (e.g., aws, gcp, azure)")
    parser.add_argument(
        "--kindfolder",
        required=True,
        help="Kind folder name (lowercase, no underscores), e.g., awslambda, gkeenvironment",
    )
    args = parser.parse_args()

    try:
        provider = normalize_segment(args.provider)
        kind_folder = normalize_segment(args.kindfolder)
    except Exception as exc:
        print(json.dumps({"exists": False, "error": f"invalid inputs: {exc}"}))
        return 2

    repo_root = os.environ.get("REPO_ROOT", find_repo_root(os.getcwd()))
    abs_path, rel_path = build_spec_proto_path(repo_root, provider, kind_folder)

    result = {
        "exists": False,
        "path": abs_path,
        "relative_path": rel_path,
        "content": "",
    }

    try:
        if os.path.isfile(abs_path):
            with open(abs_path, "r", encoding="utf-8") as f:
                result["content"] = f.read()
            result["exists"] = True
        print(json.dumps(result))
        return 0
    except Exception as exc:
        result["error"] = str(exc)
        print(json.dumps(result))
        return 1


if __name__ == "__main__":
    sys.exit(main())


