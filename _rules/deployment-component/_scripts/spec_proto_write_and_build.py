#!/usr/bin/env python3
"""
Deterministic tool: Write spec.proto content and run the API build.

Usage examples:
  # Read content from stdin
  cat my_spec.proto | python3 _rules/deployment-component/_scripts/spec_proto_write_and_build.py --provider aws --kindfolder awslambda --stdin

  # Or provide a file path
  python3 _rules/deployment-component/_scripts/spec_proto_write_and_build.py --provider aws --kindfolder awslambda --content-file /tmp/spec.proto

Outputs a JSON object to stdout with keys:
  - wrote: bool
  - path: absolute file path
  - relative_path: repo-relative path
  - bytes_written: int
  - created_dirs: list
  - sha256: string of the content
  - build_succeeded: bool
  - build_exit_code: int
  - build_stdout: string
  - build_stderr: string
  - error: optional error message
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


def run_apis_build(repo_root: str) -> Tuple[int, str, str]:
    """Run the API build using make under apis/. Returns (exit_code, stdout, stderr)."""
    apis_dir = os.path.join(repo_root, "apis")
    env = os.environ.copy()
    env["REPO_ROOT"] = repo_root
    try:
        completed = subprocess.run(
            ["make", "-C", apis_dir, "build"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            env=env,
            text=True,
            check=False,
        )
        return completed.returncode, completed.stdout, completed.stderr
    except Exception as exc:
        return 127, "", f"failed to execute make build: {exc}"


def main() -> int:
    parser = argparse.ArgumentParser(description="Write spec.proto and run apis build")
    parser.add_argument("--provider", required=True, help="Provider key (e.g., aws, gcp, azure)")
    parser.add_argument(
        "--kindfolder",
        required=True,
        help="Kind folder name (lowercase, no underscores), e.g., awslambda, gkeenvironment",
    )
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("--stdin", action="store_true", help="Read content from STDIN")
    group.add_argument("--content-file", help="Path to a file containing spec.proto content")
    args = parser.parse_args()

    try:
        provider = normalize_segment(args.provider)
        kind_folder = normalize_segment(args.kindfolder)
    except Exception as exc:
        print(json.dumps({"wrote": False, "build_succeeded": False, "error": f"invalid inputs: {exc}"}))
        return 2

    try:
        if args.stdin:
            content = sys.stdin.read()
        else:
            content_file = os.path.abspath(args.content_file)
            with open(content_file, "r", encoding="utf-8") as f:
                content = f.read()
    except Exception as exc:
        print(json.dumps({"wrote": False, "build_succeeded": False, "error": f"failed to read content: {exc}"}))
        return 3

    repo_root = os.environ.get("REPO_ROOT", find_repo_root(os.getcwd()))
    abs_path, rel_path = build_spec_proto_path(repo_root, provider, kind_folder)

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
        created_dirs = ensure_parent_dirs(abs_path)
        result["created_dirs"] = created_dirs
        with open(abs_path, "w", encoding="utf-8", newline="\n") as f:
            f.write(content)
        result["bytes_written"] = len(content.encode("utf-8"))
        result["wrote"] = True
        result["sha256"] = hashlib.sha256(content.encode("utf-8")).hexdigest()

        # Run the build non-interactively
        exit_code, stdout, stderr = run_apis_build(repo_root)
        result["build_exit_code"] = exit_code
        result["build_stdout"] = stdout
        result["build_stderr"] = stderr
        result["build_succeeded"] = exit_code == 0

        print(json.dumps(result))
        return 0 if exit_code == 0 else 4
    except Exception as exc:
        result["error"] = str(exc)
        print(json.dumps(result))
        return 1


if __name__ == "__main__":
    sys.exit(main())


