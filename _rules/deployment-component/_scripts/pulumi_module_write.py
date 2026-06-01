#!/usr/bin/env python3
"""
Deterministic tool: Write multiple files into iac/pulumi/module for a provider/kind.

Inputs are provided as a JSON manifest containing an array of files with name and content.

Manifest JSON shape:
{
  "files": [
    {"name": "main.go", "content": "..."},
    {"name": "locals.go", "content": "..."},
    {"name": "resources/security_group.go", "content": "..."}
  ]
}

Usage examples:
  # Read manifest from stdin and build afterwards
  cat manifest.json | python3 _rules/deployment-component/_scripts/pulumi_module_write.py --provider aws --kindfolder awscloudfront --stdin --build

  # Or provide a manifest file without building
  python3 _rules/deployment-component/_scripts/pulumi_module_write.py --provider aws --kindfolder awscloudfront --manifest-file /tmp/manifest.json

Outputs JSON:
  - wrote: bool
  - base_path: absolute module directory
  - base_relative_path: repo-relative module directory
  - files: array of {name, path, relative_path, bytes_written, sha256}
  - created_dirs: list
  - build_ran: bool
  - build_succeeded: bool
  - build_exit_code: int
  - build_stdout: string
  - build_stderr: string
  - error: optional string
"""

import argparse
import hashlib
import json
import os
import subprocess
import sys
from typing import Dict, List, Tuple


def find_repo_root(start_dir: str) -> str:
    current = os.path.abspath(start_dir)
    while True:
        if os.path.isdir(os.path.join(current, ".git")) or os.path.isfile(os.path.join(current, "go.mod")):
            return current
        parent = os.path.dirname(current)
        if parent == current:
            return start_dir
        current = parent


def module_base_paths(repo_root: str, provider: str, kind_folder: str) -> Tuple[str, str]:
    rel = os.path.join(
        "apis",
        "org",
        "openmcf",
        "provider",
        provider,
        kind_folder,
        "v1",
        "iac",
        "pulumi",
        "module",
    )
    return os.path.join(repo_root, rel), rel


def ensure_dir(path: str, created: List[str]) -> None:
    if not os.path.isdir(path):
        os.makedirs(path, exist_ok=True)
        created.append(path)


def norm_segment(seg: str) -> str:
    s = seg.strip()
    if s.startswith("/") or s.startswith("~") or ".." in s:
        raise ValueError("invalid filename: path traversal not allowed")
    return s


def write_files(base_abs: str, base_rel: str, files: List[Dict[str, str]]) -> Tuple[List[Dict[str, str]], List[str]]:
    created_dirs: List[str] = []
    results: List[Dict[str, str]] = []
    ensure_dir(base_abs, created_dirs)
    for f in files:
        name = norm_segment(f["name"])  # may include subdirs like "resources/sg.go"
        content = f.get("content", "")
        target_abs = os.path.join(base_abs, name)
        target_dir = os.path.dirname(target_abs)
        ensure_dir(target_dir, created_dirs)
        with open(target_abs, "w", encoding="utf-8", newline="\n") as out:
            out.write(content)
        results.append(
            {
                "name": name,
                "path": target_abs,
                "relative_path": os.path.join(base_rel, name),
                "bytes_written": len(content.encode("utf-8")),
                "sha256": hashlib.sha256(content.encode("utf-8")).hexdigest(),
            }
        )
    return results, created_dirs


def run_go_build(repo_root: str, base_rel: str) -> Tuple[int, str, str]:
    # Build the module directory; user requested not to create BUILD.bazel here.
    try:
        p = subprocess.run(
            ["go", "build", "./" + base_rel],
            cwd=repo_root,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            check=False,
        )
        return p.returncode, p.stdout, p.stderr
    except Exception as exc:
        return 127, "", str(exc)


def main() -> int:
    parser = argparse.ArgumentParser(description="Write multiple Pulumi module files deterministically")
    parser.add_argument("--provider", required=True)
    parser.add_argument("--kindfolder", required=True)
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("--stdin", action="store_true", help="Read manifest JSON from STDIN")
    group.add_argument("--manifest-file", help="Path to manifest JSON file")
    parser.add_argument("--build", action="store_true", help="Run 'go build' on the module directory after writing")
    args = parser.parse_args()

    try:
        provider = args.provider.strip().lower().replace("_", "")
        kind = args.kindfolder.strip().lower().replace("_", "")
        if any(x in provider for x in ["..", "/", "~"]) or any(x in kind for x in ["..", "/", "~"]):
            raise ValueError("invalid inputs")
    except Exception as exc:
        print(json.dumps({"wrote": False, "error": f"invalid inputs: {exc}"}))
        return 2

    try:
        if args.stdin:
            manifest_raw = sys.stdin.read()
        else:
            with open(os.path.abspath(args.manifest_file), "r", encoding="utf-8") as mf:
                manifest_raw = mf.read()
        manifest = json.loads(manifest_raw)
        files = manifest.get("files", [])
        if not isinstance(files, list) or not files:
            raise ValueError("manifest.files must be a non-empty array")
        for f in files:
            if not isinstance(f, dict) or "name" not in f or "content" not in f:
                raise ValueError("each file must be an object with 'name' and 'content'")
    except Exception as exc:
        print(json.dumps({"wrote": False, "error": f"failed to parse manifest: {exc}"}))
        return 3

    repo_root = os.environ.get("REPO_ROOT", find_repo_root(os.getcwd()))
    base_abs, base_rel = module_base_paths(repo_root, provider, kind)

    result = {
        "wrote": False,
        "base_path": base_abs,
        "base_relative_path": base_rel,
        "files": [],
        "created_dirs": [],
        "build_ran": False,
        "build_succeeded": False,
        "build_exit_code": 0,
        "build_stdout": "",
        "build_stderr": "",
    }

    try:
        files_written, created_dirs = write_files(base_abs, base_rel, files)
        result["files"] = files_written
        result["created_dirs"] = created_dirs
        result["wrote"] = True

        if args.build:
            code, out, err = run_go_build(repo_root, base_rel)
            result["build_ran"] = True
            result["build_exit_code"] = code
            result["build_stdout"] = out
            result["build_stderr"] = err
            result["build_succeeded"] = code == 0

        print(json.dumps(result))
        return 0 if (not args.build or result["build_succeeded"]) else 4
    except Exception as exc:
        result["error"] = str(exc)
        print(json.dumps(result))
        return 1


if __name__ == "__main__":
    sys.exit(main())


