#!/usr/bin/env python3
"""
Deterministic tool: Write README.md for a provider/kind.

Usage:
  python3 _rules/deployment-component/_scripts/docs_write.py --provider aws --kindfolder awscloudfront --readme-file /tmp/README.md

Outputs JSON:
  - wrote_readme: bool
  - readme_path: absolute path
  - readme_relative_path: repo-relative path
  - readme_bytes: int
  - readme_sha256: string
  - created_dirs: list
  - error: optional string
"""

import argparse
import hashlib
import json
import os
import sys
from typing import List


def find_repo_root(start_dir: str) -> str:
    current = os.path.abspath(start_dir)
    while True:
        if os.path.isdir(os.path.join(current, ".git")) or os.path.isfile(os.path.join(current, "go.mod")):
            return current
        parent = os.path.dirname(current)
        if parent == current:
            return start_dir
        current = parent


def readme_path(repo_root: str, provider: str, kind_folder: str) -> tuple[str, str]:
    base_rel = os.path.join("apis", "dev", "planton", "provider", provider, kind_folder, "v1")
    readme_rel = os.path.join(base_rel, "README.md")
    return os.path.join(repo_root, readme_rel), readme_rel


def ensure_parent(dir_path: str) -> List[str]:
    created: List[str] = []
    if not os.path.isdir(dir_path):
        os.makedirs(dir_path, exist_ok=True)
        created.append(dir_path)
    return created


def norm(seg: str) -> str:
    s = seg.strip().lower().replace("_", "")
    if ".." in s or s.startswith("/") or s.startswith("~"):
        raise ValueError("invalid segment")
    return s


def read_file_content(path: str) -> str:
    abs_path = os.path.abspath(path)
    with open(abs_path, "r", encoding="utf-8") as f:
        return f.read()


def main() -> int:
    parser = argparse.ArgumentParser(description="Write README.md deterministically")
    parser.add_argument("--provider", required=True)
    parser.add_argument("--kindfolder", required=True)
    parser.add_argument("--readme-file", required=True, help="Path to README.md content source")
    args = parser.parse_args()

    try:
        provider = norm(args.provider)
        kind = norm(args.kindfolder)
    except Exception as exc:
        print(json.dumps({"error": f"invalid inputs: {exc}"}))
        return 2

    try:
        readme_content = read_file_content(args.readme_file)
    except Exception as exc:
        print(json.dumps({"error": f"failed to read content file: {exc}"}))
        return 3

    repo_root = os.environ.get("REPO_ROOT", find_repo_root(os.getcwd()))
    readme_abs, readme_rel = readme_path(repo_root, provider, kind)
    out_dir = os.path.dirname(readme_abs)

    result = {
        "wrote_readme": False,
        "readme_path": readme_abs,
        "readme_relative_path": readme_rel,
        "readme_bytes": 0,
        "readme_sha256": "",
        "created_dirs": [],
    }

    try:
        result["created_dirs"] = ensure_parent(out_dir)
        with open(readme_abs, "w", encoding="utf-8", newline="\n") as f:
            f.write(readme_content)
        result["wrote_readme"] = True
        result["readme_bytes"] = len(readme_content.encode("utf-8"))
        result["readme_sha256"] = hashlib.sha256(readme_content.encode("utf-8")).hexdigest()
        print(json.dumps(result))
        return 0
    except Exception as exc:
        result["error"] = str(exc)
        print(json.dumps(result))
        return 1


if __name__ == "__main__":
    sys.exit(main())
