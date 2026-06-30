#!/usr/bin/env python3
"""
Deterministic tool: Write multiple Terraform module files under iac/tf for a provider/kind.

Also supports generating variables.tf via the Planton CLI and validating with Terraform.

Manifest JSON shape (stdin or file):
{
  "files": [
    {"name": "provider.tf", "content": "..."},
    {"name": "locals.tf", "content": "..."},
    {"name": "outputs.tf", "content": "..."},
    {"name": "resources/instance.tf", "content": "..."}
  ]
}

Usage examples:
  # Write files, run make local, generate variables.tf, and validate
  cat manifest.json | python3 _rules/deployment-component/_scripts/terraform_module_write.py \
    --provider aws --kindfolder awscloudfront --kind AwsCloudFront \
    --stdin --make-local --generate-variables --validate

Outputs JSON:
  - base_path, base_relative_path
  - files: [{name, path, relative_path, bytes_written, sha256}]
  - created_dirs: list
  - make_local: {exit_code, stdout, stderr} (if run)
  - generate_variables: {exit_code, stdout, stderr, output_file} (if run)
  - generate_module: {exit_code, stdout, stderr, output_dir} (if run; full thin module for projection kinds)
  - tf_init: {exit_code, stdout, stderr} (if run)
  - tf_validate: {exit_code, stdout, stderr} (if run)
  - validate_succeeded: bool
  - error: optional
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
        "planton",
        "provider",
        provider,
        kind_folder,
        "v1",
        "iac",
        "tf",
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
        name = norm_segment(f["name"])  # may include subdirs
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


def run(cmd, cwd=None, env=None):
    try:
        p = subprocess.run(cmd, cwd=cwd, env=env, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, check=False)
        return {"exit_code": p.returncode, "stdout": p.stdout, "stderr": p.stderr}
    except Exception as exc:
        return {"exit_code": 127, "stdout": "", "stderr": str(exc)}


def main() -> int:
    parser = argparse.ArgumentParser(description="Write Terraform module files deterministically")
    parser.add_argument("--provider", required=True)
    parser.add_argument("--kindfolder", required=True)
    parser.add_argument("--kind", required=True, help="PascalCase kind name, e.g., AwsCloudFront")
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("--stdin", action="store_true", help="Read manifest JSON from STDIN")
    group.add_argument("--manifest-file", help="Path to manifest JSON file")
    parser.add_argument("--make-local", action="store_true", help="Run 'make local' before generating variables.tf")
    parser.add_argument("--generate-variables", action="store_true", help="Generate variables.tf via Planton CLI")
    parser.add_argument("--generate-module", action="store_true",
                        help="Generate the ENTIRE thin module via Planton CLI (kubernetes_manifest-projection kinds only). "
                             "Emits variables.tf/locals.tf/main.tf/provider.tf/outputs.tf; no hand-authored files needed.")
    parser.add_argument("--validate", action="store_true", help="Run terraform init/validate after writing")
    args = parser.parse_args()

    try:
        provider = args.provider.strip().lower().replace("_", "")
        kind_folder = args.kindfolder.strip().lower().replace("_", "")
        kind_name = args.kind.strip()
        if any(x in provider for x in ["..", "/", "~"]) or any(x in kind_folder for x in ["..", "/", "~"]):
            raise ValueError("invalid provider/kindfolder")
    except Exception as exc:
        print(json.dumps({"error": f"invalid inputs: {exc}"}))
        return 2

    try:
        if args.stdin:
            manifest_raw = sys.stdin.read()
        else:
            with open(os.path.abspath(args.manifest_file), "r", encoding="utf-8") as mf:
                manifest_raw = mf.read()
        manifest = json.loads(manifest_raw)
        files = manifest.get("files", [])
        if not isinstance(files, list):
            raise ValueError("manifest.files must be an array (can be empty if only generating variables.tf)")
        for f in files:
            if not isinstance(f, dict) or "name" not in f or "content" not in f:
                raise ValueError("each file must be an object with 'name' and 'content'")
    except Exception as exc:
        print(json.dumps({"error": f"failed to parse manifest: {exc}"}))
        return 3

    repo_root = os.environ.get("REPO_ROOT", find_repo_root(os.getcwd()))
    base_abs, base_rel = module_base_paths(repo_root, provider, kind_folder)

    result = {
        "base_path": base_abs,
        "base_relative_path": base_rel,
        "files": [],
        "created_dirs": [],
    }

    env = os.environ.copy()

    # Optionally make local
    if args.make_local:
        result["make_local"] = run(["make", "-C", repo_root, "local"], cwd=repo_root, env=env)

    try:
        files_written, created_dirs = write_files(base_abs, base_rel, files)
        result["files"] = files_written
        result["created_dirs"] = created_dirs
    except Exception as exc:
        result["error"] = str(exc)
        print(json.dumps(result))
        return 1

    # Optionally generate variables.tf
    if args.generate_variables:
        out_file = os.path.join(base_abs, "variables.tf")
        gen_cmd = [
            "planton",
            "tofu",
            "generate-variables",
            kind_name,
            "--output-file",
            out_file,
        ]
        result["generate_variables"] = run(gen_cmd, cwd=repo_root, env=env)
        result["generate_variables"]["output_file"] = out_file

    # Optionally generate the entire thin module (kubernetes_manifest projection kinds).
    # This emits all five .tf files deterministically; the kind must carry a
    # kubernetes_manifest_projection in CloudResourceKindMeta or the CLI errors out.
    if args.generate_module:
        gen_cmd = [
            "planton",
            "tofu",
            "generate-module",
            kind_name,
            "--output-dir",
            base_abs,
        ]
        result["generate_module"] = run(gen_cmd, cwd=repo_root, env=env)
        result["generate_module"]["output_dir"] = base_abs

    # Optionally validate the module
    if args.validate:
        # terraform init -backend=false
        result["tf_init"] = run(["terraform", "-chdir=" + base_abs, "init", "-backend=false"], cwd=repo_root, env=env)
        # terraform validate
        result["tf_validate"] = run(["terraform", "-chdir=" + base_abs, "validate"], cwd=repo_root, env=env)
        result["validate_succeeded"] = (
            result.get("tf_init", {}).get("exit_code", 1) == 0 and result.get("tf_validate", {}).get("exit_code", 1) == 0
        )

    print(json.dumps(result))
    return 0


if __name__ == "__main__":
    sys.exit(main())


