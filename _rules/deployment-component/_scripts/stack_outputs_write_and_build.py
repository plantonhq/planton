#!/usr/bin/env python3
"""
Deterministic tool: Write stack_outputs.proto and run apis build.

Usage:
  cat stack_outputs.proto | python3 _rules/deployment-component/_scripts/stack_outputs_write_and_build.py --provider aws --kindfolder awscloudfront --stdin

Outputs JSON:
  - wrote: bool
  - path, relative_path
  - bytes_written, created_dirs, sha256
  - build_succeeded, build_exit_code, build_stdout, build_stderr
  - error (optional)
"""

import argparse
import hashlib
import json
import os
import subprocess
import sys
from typing import List, Tuple


def find_repo_root(start_dir: str) -> str:
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


def outputs_path(repo_root: str, provider: str, kind_folder: str) -> Tuple[str, str]:
    rel = os.path.join("apis", "dev", "planton", "provider", provider, kind_folder, "v1", "stack_outputs.proto")
    return os.path.join(repo_root, rel), rel


def ensure_parent(file_path: str) -> List[str]:
    created = []
    parent = os.path.dirname(file_path)
    if not os.path.isdir(parent):
        os.makedirs(parent, exist_ok=True)
        created.append(parent)
    return created


def normalize(segment: str) -> str:
    s = segment.strip().lower().replace("_", "")
    if ".." in s or s.startswith("/") or s.startswith("~"):
        raise ValueError("invalid segment")
    return s


def run_build(repo_root: str) -> Tuple[int, str, str]:
    apis_dir = os.path.join(repo_root, "apis")
    try:
        p = subprocess.run(["make", "-C", apis_dir, "build"], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, check=False)
        return p.returncode, p.stdout, p.stderr
    except Exception as exc:
        return 127, "", str(exc)


def main() -> int:
    parser = argparse.ArgumentParser(description="Write stack_outputs.proto and build")
    parser.add_argument("--provider", required=True)
    parser.add_argument("--kindfolder", required=True)
    g = parser.add_mutually_exclusive_group(required=True)
    g.add_argument("--stdin", action="store_true")
    g.add_argument("--content-file")
    args = parser.parse_args()

    try:
        provider = normalize(args.provider)
        kind = normalize(args.kindfolder)
    except Exception as exc:
        print(json.dumps({"wrote": False, "build_succeeded": False, "error": f"invalid inputs: {exc}"}))
        return 2

    try:
        if args.stdin:
            content = sys.stdin.read()
        else:
            with open(os.path.abspath(args.content_file), "r", encoding="utf-8") as f:
                content = f.read()
    except Exception as exc:
        print(json.dumps({"wrote": False, "build_succeeded": False, "error": f"failed to read content: {exc}"}))
        return 3

    repo_root = os.environ.get("REPO_ROOT", find_repo_root(os.getcwd()))
    abs_path, rel_path = outputs_path(repo_root, provider, kind)
    result = {
        "wrote": False,
        "path": abs_path,
        "relative_path": rel_path,
        "bytes_written": 0,
        "created_dirs": [],
        "sha256": "",
        "build_succeeded": False,
        "build_exit_code": 0,
        "build_stdout": "",
        "build_stderr": "",
    }

    try:
        result["created_dirs"] = ensure_parent(abs_path)
        with open(abs_path, "w", encoding="utf-8", newline="\n") as f:
            f.write(content)
        result["bytes_written"] = len(content.encode("utf-8"))
        result["wrote"] = True
        result["sha256"] = hashlib.sha256(content.encode("utf-8")).hexdigest()

        code, out, err = run_build(repo_root)
        result["build_exit_code"] = code
        result["build_stdout"] = out
        result["build_stderr"] = err
        result["build_succeeded"] = code == 0
        print(json.dumps(result))
        return 0 if code == 0 else 4
    except Exception as exc:
        result["error"] = str(exc)
        print(json.dumps(result))
        return 1


if __name__ == "__main__":
    sys.exit(main())


