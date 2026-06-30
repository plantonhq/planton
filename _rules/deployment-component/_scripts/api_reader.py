#!/usr/bin/env python3
import argparse
import json
import os
import sys
from typing import Tuple


def find_repo_root(start_dir: str) -> str:
    current = os.path.abspath(start_dir)
    while True:
        if os.path.isdir(os.path.join(current, ".git")) or os.path.isfile(os.path.join(current, "go.mod")):
            return current
        parent = os.path.dirname(current)
        if parent == current:
            return start_dir
        current = parent


def api_path(repo_root: str, provider: str, kind_folder: str) -> Tuple[str, str]:
    rel = os.path.join("apis", "org", "planton", "provider", provider, kind_folder, "v1", "api.proto")
    return os.path.join(repo_root, rel), rel


def norm(seg: str) -> str:
    s = seg.strip().lower().replace("_", "")
    if ".." in s or s.startswith("/") or s.startswith("~"):
        raise ValueError("invalid segment")
    return s


def main() -> int:
    p = argparse.ArgumentParser(description="Read existing api.proto for provider/kind")
    p.add_argument("--provider", required=True)
    p.add_argument("--kindfolder", required=True)
    args = p.parse_args()

    try:
        provider = norm(args.provider)
        kind = norm(args.kindfolder)
    except Exception as exc:
        print(json.dumps({"exists": False, "error": f"invalid inputs: {exc}"}))
        return 2

    repo_root = os.environ.get("REPO_ROOT", find_repo_root(os.getcwd()))
    abs_path, rel_path = api_path(repo_root, provider, kind)

    out = {"exists": False, "path": abs_path, "relative_path": rel_path, "content": ""}
    try:
        if os.path.isfile(abs_path):
            with open(abs_path, "r", encoding="utf-8") as f:
                out["content"] = f.read()
            out["exists"] = True
        print(json.dumps(out))
        return 0
    except Exception as exc:
        out["error"] = str(exc)
        print(json.dumps(out))
        return 1


if __name__ == "__main__":
    sys.exit(main())


