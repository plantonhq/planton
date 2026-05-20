#!/usr/bin/env python3
"""
Deterministic tool: Write Terraform docs (README.md) under iac/tf/ for a provider/kind.

Usage:
  python3 _rules/deployment-component/_scripts/terraform_docs_write.py --provider aws --kindfolder awscloudfront \
    --readme-file /tmp/README.md

Outputs JSON similar to pulumi_docs_write.py (without debug.sh).
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


def base_paths(repo_root: str, provider: str, kind_folder: str) -> Tuple[str, str]:
    rel = os.path.join("apis", "org", "openmcf", "provider", provider, kind_folder, "v1", "iac", "tf")
    return os.path.join(repo_root, rel), rel


def ensure_dir(path: str, created: List[str]) -> None:
    if not os.path.isdir(path):
        os.makedirs(path, exist_ok=True)
        created.append(path)


def norm(seg: str) -> str:
    s = seg.strip().lower().replace("_", "")
    if ".." in s or s.startswith("/") or s.startswith("~"):
        raise ValueError("invalid segment")
    return s


def read_file(path: str) -> str:
    with open(os.path.abspath(path), "r", encoding="utf-8") as f:
        return f.read()


def main() -> int:
    parser = argparse.ArgumentParser(description="Write Terraform docs deterministically")
    parser.add_argument("--provider", required=True)
    parser.add_argument("--kindfolder", required=True)
    parser.add_argument("--readme-file", required=True)
    args = parser.parse_args()

    try:
        provider = norm(args.provider)
        kind = norm(args.kindfolder)
    except Exception as exc:
        print(json.dumps({"error": f"invalid inputs: {exc}"}))
        return 2

    try:
        readme_content = read_file(args.readme_file)
    except Exception as exc:
        print(json.dumps({"error": f"failed to read content file: {exc}"}))
        return 3

    repo_root = os.environ.get("REPO_ROOT", find_repo_root(os.getcwd()))
    base_abs, base_rel = base_paths(repo_root, provider, kind)

    result = {
        "base_path": base_abs,
        "base_relative_path": base_rel,
        "created_dirs": [],
        "wrote_readme": False,
    }

    try:
        ensure_dir(base_abs, result["created_dirs"])

        readme_abs = os.path.join(base_abs, "README.md")
        with open(readme_abs, "w", encoding="utf-8", newline="\n") as f:
            f.write(readme_content)
        result.update({
            "readme_path": readme_abs,
            "readme_relative_path": os.path.join(base_rel, "README.md"),
            "readme_bytes": len(readme_content.encode("utf-8")),
            "readme_sha256": hashlib.sha256(readme_content.encode("utf-8")).hexdigest(),
            "wrote_readme": True,
        })

        print(json.dumps(result))
        return 0
    except Exception as exc:
        result["error"] = str(exc)
        print(json.dumps(result))
        return 1


if __name__ == "__main__":
    sys.exit(main())
