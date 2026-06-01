#!/usr/bin/env python3
"""
Deterministic tool: Write Pulumi entrypoint files (main.go, Pulumi.yaml, Makefile) under iac/pulumi/.

Usage (stdin JSON):
  cat files.json | python3 _rules/deployment-component/_scripts/pulumi_entrypoint_write.py --provider aws --kindfolder awscloudfront --stdin --build

JSON shape:
{
  "files": [
    {"name": "main.go", "content": "..."},
    {"name": "Pulumi.yaml", "content": "..."},
    {"name": "Makefile", "content": "..."}
  ]
}

Outputs JSON similar to pulumi_module_write.py.
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


def base_paths(repo_root: str, provider: str, kind_folder: str) -> Tuple[str, str]:
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
    )
    return os.path.join(repo_root, rel), rel


def ensure_dir(path: str, created: List[str]) -> None:
    if not os.path.isdir(path):
        os.makedirs(path, exist_ok=True)
        created.append(path)


def norm_name(name: str) -> str:
    s = name.strip()
    if s.startswith("/") or s.startswith("~") or ".." in s:
        raise ValueError("invalid name")
    return s


def write_files(base_abs: str, base_rel: str, files: List[Dict[str, str]]) -> Tuple[List[Dict[str, str]], List[str]]:
    created_dirs: List[str] = []
    results: List[Dict[str, str]] = []
    ensure_dir(base_abs, created_dirs)
    for f in files:
        name = norm_name(f["name"])  # do not allow subdirs here
        content = f.get("content", "")
        target_abs = os.path.join(base_abs, name)
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
    # Build the entrypoint directory NON-RECURSIVELY, exactly as the release pipeline does
    # (release.pulumi-modules.yaml runs `go build -o <bin> ./<...>/v1/iac/pulumi`). A recursive
    # `./<base_rel>/...` build would compile the module/ library and PASS even when the root
    # `package main` entrypoint is missing or misplaced -- masking the very failure this writer
    # must catch. `-o /dev/null` discards the binary while still requiring a buildable root main.
    try:
        p = subprocess.run(["go", "build", "-o", os.devnull, "./" + base_rel], cwd=repo_root, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, check=False)
        return p.returncode, p.stdout, p.stderr
    except Exception as exc:
        return 127, "", str(exc)


def main() -> int:
    parser = argparse.ArgumentParser(description="Write Pulumi entrypoint files deterministically")
    parser.add_argument("--provider", required=True)
    parser.add_argument("--kindfolder", required=True)
    g = parser.add_mutually_exclusive_group(required=True)
    g.add_argument("--stdin", action="store_true")
    g.add_argument("--manifest-file")
    parser.add_argument("--build", action="store_true")
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
                raise ValueError("each file must include 'name' and 'content'")
    except Exception as exc:
        print(json.dumps({"wrote": False, "error": f"failed to parse manifest: {exc}"}))
        return 3

    repo_root = os.environ.get("REPO_ROOT", find_repo_root(os.getcwd()))
    base_abs, base_rel = base_paths(repo_root, provider, kind)

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


