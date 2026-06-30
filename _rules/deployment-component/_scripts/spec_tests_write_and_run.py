#!/usr/bin/env python3
"""
Deterministic tool: Write spec_test.go for a provider/kind and run `go test`.

Usage examples:
  # From stdin
  cat spec_test.go | python3 _rules/deployment-component/_scripts/spec_tests_write_and_run.py --provider aws --kindfolder awscloudfront --stdin

  # From file
  python3 _rules/deployment-component/_scripts/spec_tests_write_and_run.py --provider aws --kindfolder awscloudfront --content-file /tmp/spec_test.go

Outputs JSON:
  - wrote: bool
  - path: absolute file path
  - relative_path: repo-relative path
  - bytes_written: int
  - created_dirs: list
  - sha256: string
  - test_succeeded: bool
  - test_exit_code: int
  - test_stdout: string
  - test_stderr: string
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


def build_test_path(repo_root: str, provider: str, kind_folder: str) -> Tuple[str, str, str]:
    base_rel = os.path.join(
        "apis",
        "dev",
        "planton",
        "provider",
        provider,
        kind_folder,
        "v1",
    )
    relative_path = os.path.join(base_rel, "spec_test.go")
    absolute_path = os.path.join(repo_root, relative_path)
    return absolute_path, relative_path, base_rel


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


def run_go_test(repo_root: str, rel_dir: str) -> Tuple[int, str, str]:
    env = os.environ.copy()
    env["REPO_ROOT"] = repo_root
    cmd = ["go", "test", "./" + rel_dir + "/..."]
    try:
        completed = subprocess.run(
            cmd,
            cwd=repo_root,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            env=env,
            check=False,
        )
        return completed.returncode, completed.stdout, completed.stderr
    except Exception as exc:
        return 127, "", f"failed to execute go test: {exc}"


def format_go_file(file_path: str) -> Tuple[int, str, str]:
    """Format a single Go file in-place using `gofmt -w`.

    Returns a tuple of (exit_code, stdout, stderr). If gofmt isn't available,
    returns exit code 127.
    """
    try:
        completed = subprocess.run(
            ["gofmt", "-w", file_path],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            check=False,
        )
        return completed.returncode, completed.stdout, completed.stderr
    except Exception as exc:
        return 127, "", f"failed to execute gofmt: {exc}"


def format_go_package(repo_root: str, rel_dir: str) -> Tuple[int, str, str]:
    """Format a Go package directory using `go fmt ./<rel_dir>` as a fallback."""
    env = os.environ.copy()
    env["REPO_ROOT"] = repo_root
    cmd = ["go", "fmt", "./" + rel_dir]
    try:
        completed = subprocess.run(
            cmd,
            cwd=repo_root,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            env=env,
            check=False,
        )
        return completed.returncode, completed.stdout, completed.stderr
    except Exception as exc:
        return 127, "", f"failed to execute go fmt: {exc}"


def main() -> int:
    parser = argparse.ArgumentParser(description="Write spec_test.go and run go test")
    parser.add_argument("--provider", required=True, help="Provider key (e.g., aws, gcp, azure)")
    parser.add_argument(
        "--kindfolder",
        required=True,
        help="Kind folder name (lowercase, no underscores), e.g., awscloudfront, gkeenvironment",
    )
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("--stdin", action="store_true", help="Read content from STDIN")
    group.add_argument("--content-file", help="Path to a file containing spec_test.go content")
    args = parser.parse_args()

    try:
        provider = normalize_segment(args.provider)
        kind_folder = normalize_segment(args.kindfolder)
    except Exception as exc:
        print(json.dumps({"wrote": False, "test_succeeded": False, "error": f"invalid inputs: {exc}"}))
        return 2

    try:
        if args.stdin:
            content = sys.stdin.read()
        else:
            content_file = os.path.abspath(args.content_file)
            with open(content_file, "r", encoding="utf-8") as f:
                content = f.read()
    except Exception as exc:
        print(json.dumps({"wrote": False, "test_succeeded": False, "error": f"failed to read content: {exc}"}))
        return 3

    repo_root = os.environ.get("REPO_ROOT", find_repo_root(os.getcwd()))
    abs_path, rel_path, rel_dir = build_test_path(repo_root, provider, kind_folder)

    result = {
        "wrote": False,
        "path": abs_path,
        "relative_path": rel_path,
        "bytes_written": 0,
        "created_dirs": [],
        "sha256": "",
        "test_succeeded": False,
        "test_exit_code": 0,
        "test_stdout": "",
        "test_stderr": "",
    }

    try:
        created_dirs = ensure_parent_dirs(abs_path)
        result["created_dirs"] = created_dirs
        with open(abs_path, "w", encoding="utf-8", newline="\n") as f:
            f.write(content)
        result["bytes_written"] = len(content.encode("utf-8"))
        result["wrote"] = True
        result["sha256"] = hashlib.sha256(content.encode("utf-8")).hexdigest()

        # Best-effort format to ensure proper indentation in editors before tests run
        fmt_code, _, _ = format_go_file(abs_path)
        if fmt_code != 0:
            # Fallback: try formatting the whole package directory
            format_go_package(repo_root, rel_dir)

        exit_code, stdout, stderr = run_go_test(repo_root, rel_dir)
        result["test_exit_code"] = exit_code
        result["test_stdout"] = stdout
        result["test_stderr"] = stderr
        result["test_succeeded"] = exit_code == 0

        print(json.dumps(result))
        return 0 if exit_code == 0 else 4
    except Exception as exc:
        result["error"] = str(exc)
        print(json.dumps(result))
        return 1


if __name__ == "__main__":
    sys.exit(main())


