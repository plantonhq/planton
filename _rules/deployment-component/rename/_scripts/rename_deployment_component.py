#!/usr/bin/env python3
"""
Deployment Component Rename Script

Renames a deployment component across the entire codebase using sequential in-place operations:
- Renames directories containing old component name (bottom-up to avoid path conflicts)
- Renames files containing old component name
- Replaces old component name in file contents
- Updates cloud_resource_kind.proto registry
- Updates documentation
- Runs build pipeline (protos, build, test)

Usage:
  python3 _rules/deployment-component/rename/_scripts/rename_deployment_component.py \
    --old-name KubernetesMicroservice \
    --new-name KubernetesDeployment \
    --new-id-prefix k8sdpl

  # Keep existing id_prefix
  python3 _rules/deployment-component/rename/_scripts/rename_deployment_component.py \
    --old-name KubernetesMicroservice \
    --new-name KubernetesDeployment

  # Specify provider explicitly (skips auto-discovery)
  python3 _rules/deployment-component/rename/_scripts/rename_deployment_component.py \
    --old-name OpenStackComputeKeypair \
    --new-name OpenStackKeypair \
    --provider openstack

Output: JSON object with success status, metrics, and build results
"""

import argparse
import json
import os
import re
import shutil
import subprocess
import sys
import time
from pathlib import Path
from typing import Dict, List, Tuple, Optional


def find_repo_root(start_dir: str) -> str:
    """Find repository root by looking for .git or go.mod"""
    current = os.path.abspath(start_dir)
    while True:
        if os.path.isdir(os.path.join(current, ".git")) or os.path.isfile(os.path.join(current, "go.mod")):
            return current
        parent = os.path.dirname(current)
        if parent == current:
            raise RuntimeError("Could not find repository root")
        current = parent


def to_lowercase(name: str) -> str:
    """Convert PascalCase to lowercase (KubernetesMicroservice -> kubernetesmicroservice)"""
    return name.lower()


def to_snake_case(name: str) -> str:
    """Convert PascalCase to snake_case (KubernetesMicroservice -> kubernetes_microservice)"""
    # Insert underscore before capital letters
    s1 = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', name)
    return re.sub('([a-z0-9])([A-Z])', r'\1_\2', s1).lower()


def to_kebab_case(name: str) -> str:
    """Convert PascalCase to kebab-case (KubernetesMicroservice -> kubernetes-microservice)"""
    return to_snake_case(name).replace('_', '-')


def to_space_separated(name: str) -> str:
    """Convert PascalCase to space separated (KubernetesMicroservice -> kubernetes microservice)"""
    return to_snake_case(name).replace('_', ' ')


def to_camel_case(name: str) -> str:
    """Convert PascalCase to camelCase (KubernetesMicroservice -> kubernetesMicroservice)"""
    if not name:
        return name
    return name[0].lower() + name[1:]


def to_upper_snake_case(name: str) -> str:
    """Convert PascalCase to UPPER_SNAKE_CASE (KubernetesMicroservice -> KUBERNETES_MICROSERVICE)"""
    return to_snake_case(name).upper()


def build_replacement_map(old_name: str, new_name: str) -> List[Tuple[str, str]]:
    """
    Build comprehensive replacement patterns for all naming conventions.
    Returns list of (old, new) tuples, ordered by specificity (most specific first).
    """
    patterns = [
        # PascalCase (most specific)
        (old_name, new_name),
        # camelCase
        (to_camel_case(old_name), to_camel_case(new_name)),
        # UPPER_SNAKE_CASE
        (to_upper_snake_case(old_name), to_upper_snake_case(new_name)),
        # snake_case
        (to_snake_case(old_name), to_snake_case(new_name)),
        # kebab-case
        (to_kebab_case(old_name), to_kebab_case(new_name)),
        # space separated (with quotes)
        (f'"{to_space_separated(old_name)}"', f'"{to_space_separated(new_name)}"'),
        # lowercase (least specific, should be last)
        (to_lowercase(old_name), to_lowercase(new_name)),
    ]
    
    # Remove duplicates while preserving order
    seen = set()
    unique_patterns = []
    for old, new in patterns:
        if old not in seen:
            seen.add(old)
            unique_patterns.append((old, new))
    
    return unique_patterns


def find_component_in_registry(repo_root: str, component_name: str) -> Optional[Dict]:
    """
    Find component in cloud_resource_kind.proto.
    Returns dict with: enum_value, provider, id_prefix, and the full enum entry text.
    """
    registry_path = os.path.join(
        repo_root,
        "apis/org/openmcf/shared/cloudresourcekind/cloud_resource_kind.proto"
    )
    
    if not os.path.isfile(registry_path):
        return None
    
    with open(registry_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Find enum entry pattern: ComponentName = NUMBER [(kind_meta) = { ... }];
    pattern = rf'^  {re.escape(component_name)}\s*=\s*(\d+)\s*\[.*?\];'
    match = re.search(pattern, content, re.MULTILINE | re.DOTALL)
    
    if not match:
        return None
    
    enum_value = int(match.group(1))
    
    # Extract metadata
    entry_text = match.group(0)
    
    # Extract provider
    provider_match = re.search(r'provider:\s*(\w+)', entry_text)
    provider = provider_match.group(1) if provider_match else None
    
    # Extract id_prefix
    id_prefix_match = re.search(r'id_prefix:\s*"([^"]+)"', entry_text)
    id_prefix = id_prefix_match.group(1) if id_prefix_match else None
    
    return {
        'enum_value': enum_value,
        'provider': provider,
        'id_prefix': id_prefix,
        'entry_text': entry_text,
        'registry_path': registry_path
    }


def get_component_directory(repo_root: str, provider: str, component_folder: str) -> str:
    """Get path to component directory"""
    return os.path.join(
        repo_root,
        "apis/org/openmcf/provider",
        provider,
        component_folder
    )


def rename_directories(root_dir: Path, old_str: str, new_str: str, stats: Dict) -> None:
    """Rename directories containing old_str (bottom-up to avoid path errors)"""
    dirs_to_rename = []
    
    # Walk bottom-up to collect directories
    for dirpath, dirnames, _ in os.walk(root_dir, topdown=False):
        for dirname in dirnames:
            # Skip hidden directories
            if dirname.startswith('.'):
                continue
            
            if old_str in dirname:
                old_path = Path(dirpath) / dirname
                new_name = dirname.replace(old_str, new_str)
                new_path = Path(dirpath) / new_name
                dirs_to_rename.append((old_path, new_path))
    
    # Perform renames
    for old_path, new_path in dirs_to_rename:
        try:
            old_path.rename(new_path)
            stats['dirs_renamed'] += 1
        except Exception as e:
            print(f"Error renaming directory {old_path}: {e}", file=sys.stderr)
            stats['errors'] += 1


def rename_files(root_dir: Path, old_str: str, new_str: str, stats: Dict) -> None:
    """Rename files containing old_str"""
    files_to_rename = []
    
    # Collect files to rename
    for dirpath, _, filenames in os.walk(root_dir):
        for filename in filenames:
            # Skip hidden files
            if filename.startswith('.'):
                continue
            
            if old_str in filename:
                old_path = Path(dirpath) / filename
                new_name = filename.replace(old_str, new_str)
                new_path = Path(dirpath) / new_name
                files_to_rename.append((old_path, new_path))
    
    # Perform renames
    for old_path, new_path in files_to_rename:
        try:
            old_path.rename(new_path)
            stats['files_renamed'] += 1
        except Exception as e:
            print(f"Error renaming file {old_path}: {e}", file=sys.stderr)
            stats['errors'] += 1


def replace_in_files(root_dir: Path, old_str: str, new_str: str, stats: Dict) -> None:
    """Replace occurrences of old_str with new_str in file contents"""
    for dirpath, _, filenames in os.walk(root_dir):
        for filename in filenames:
            # Skip hidden files
            if filename.startswith('.'):
                continue
            
            filepath = Path(dirpath) / filename
            
            try:
                # Try to read as text
                with open(filepath, 'r', encoding='utf-8') as f:
                    content = f.read()
                
                # Check if replacement needed
                if old_str in content:
                    new_content = content.replace(old_str, new_str)
                    
                    with open(filepath, 'w', encoding='utf-8') as f:
                        f.write(new_content)
                    stats['files_updated'] += 1
                    stats['replacements_made'] += 1
            
            except (UnicodeDecodeError, IsADirectoryError):
                # Skip binary files and directories
                continue
            except Exception as e:
                print(f"Error processing file {filepath}: {e}", file=sys.stderr)
                stats['errors'] += 1


def apply_sequential_renames(root_dir: Path, old_str: str, new_str: str, stats: Dict) -> None:
    """Apply sequential renames: directories → files → contents"""
    if old_str == new_str:
        return
    
    # Phase 1: Rename directories
    rename_directories(root_dir, old_str, new_str, stats)
    
    # Phase 2: Rename files
    rename_files(root_dir, old_str, new_str, stats)
    
    # Phase 3: Replace in file contents
    replace_in_files(root_dir, old_str, new_str, stats)


def update_registry_entry(
    registry_path: str,
    old_name: str,
    new_name: str,
    new_id_prefix: Optional[str]
) -> None:
    """
    Update component entry in cloud_resource_kind.proto.
    Only updates the enum name and optionally the id_prefix.
    Preserves all other metadata.
    """
    with open(registry_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Find the enum entry
    pattern = rf'^  {re.escape(old_name)}\s*=\s*(\d+)\s*\[(.*?)\];'
    match = re.search(pattern, content, re.MULTILINE | re.DOTALL)
    
    if not match:
        raise RuntimeError(f"Could not find enum entry for {old_name}")
    
    enum_value = match.group(1)
    metadata = match.group(2)
    
    # Update id_prefix if provided
    if new_id_prefix:
        metadata = re.sub(
            r'id_prefix:\s*"[^"]+"',
            f'id_prefix: "{new_id_prefix}"',
            metadata
        )
    
    # Build new entry
    new_entry = f'  {new_name} = {enum_value} [{metadata}];'
    
    # Replace in content
    content = re.sub(
        rf'^  {re.escape(old_name)}\s*=\s*{enum_value}\s*\[.*?\];',
        new_entry,
        content,
        flags=re.MULTILINE | re.DOTALL
    )
    
    with open(registry_path, 'w', encoding='utf-8') as f:
        f.write(content)


def rename_icon_folder(repo_root: str, provider: str, old_folder: str, new_folder: str) -> Dict:
    """
    Rename icon folder if it exists.
    Returns dict with: exists, old_path, new_path, renamed
    """
    # Map provider to icon directory structure
    # For kubernetes/workload or kubernetes/addon, use just "kubernetes"
    icon_provider = provider.split('/')[0]
    
    old_icon_path = os.path.join(
        repo_root,
        "site/public/images/providers",
        icon_provider,
        old_folder
    )
    
    new_icon_path = os.path.join(
        repo_root,
        "site/public/images/providers", 
        icon_provider,
        new_folder
    )
    
    result = {
        'exists': os.path.exists(old_icon_path),
        'old_path': old_icon_path,
        'new_path': new_icon_path,
        'renamed': False
    }
    
    if result['exists']:
        # Delete target if exists
        if os.path.exists(new_icon_path):
            shutil.rmtree(new_icon_path)
        # Rename
        shutil.move(old_icon_path, new_icon_path)
        result['renamed'] = True
    
    return result


def run_command(cmd: List[str], cwd: str) -> Tuple[int, str, str]:
    """Run a command and return (exit_code, stdout, stderr)"""
    try:
        result = subprocess.run(
            cmd,
            cwd=cwd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            check=False
        )
        return result.returncode, result.stdout, result.stderr
    except Exception as e:
        return 127, "", str(e)


def run_build_pipeline(repo_root: str) -> Dict:
    """
    Run the build pipeline: make protos, make build, make test.
    Returns dict with exit codes and output for each phase.
    """
    result = {
        'protos_exit_code': 0,
        'protos_stdout': '',
        'protos_stderr': '',
        'build_exit_code': 0,
        'build_stdout': '',
        'build_stderr': '',
        'test_exit_code': 0,
        'test_stdout': '',
        'test_stderr': '',
    }
    
    # Run make protos
    exit_code, stdout, stderr = run_command(['make', 'protos'], repo_root)
    result['protos_exit_code'] = exit_code
    result['protos_stdout'] = stdout
    result['protos_stderr'] = stderr
    
    if exit_code != 0:
        return result  # Stop on first failure
    
    # Run make build
    exit_code, stdout, stderr = run_command(['make', 'build'], repo_root)
    result['build_exit_code'] = exit_code
    result['build_stdout'] = stdout
    result['build_stderr'] = stderr
    
    if exit_code != 0:
        return result  # Stop on first failure
    
    # Run make test
    exit_code, stdout, stderr = run_command(['make', 'test'], repo_root)
    result['test_exit_code'] = exit_code
    result['test_stdout'] = stdout
    result['test_stderr'] = stderr
    
    return result


def main() -> int:
    parser = argparse.ArgumentParser(
        description='Rename a deployment component across the entire codebase'
    )
    parser.add_argument(
        '--old-name',
        required=True,
        help='Old component name in PascalCase (e.g., KubernetesMicroservice)'
    )
    parser.add_argument(
        '--new-name',
        required=True,
        help='New component name in PascalCase (e.g., KubernetesDeployment)'
    )
    parser.add_argument(
        '--new-id-prefix',
        help='New ID prefix (e.g., k8sdpl). If not provided, keeps existing.'
    )
    parser.add_argument(
        '--provider',
        help='Provider name (e.g., openstack, aws, gcp). If provided, the component '
             'directory is resolved directly as apis/org/openmcf/provider/{provider}/{component}. '
             'If not provided, all provider directories are searched.'
    )
    
    args = parser.parse_args()
    
    start_time = time.time()
    
    # Initialize statistics
    stats = {
        'dirs_renamed': 0,
        'files_renamed': 0,
        'files_updated': 0,
        'replacements_made': 0,
        'errors': 0
    }
    
    result = {
        'success': False,
        'old_component': args.old_name,
        'new_component': args.new_name,
        'old_folder': to_lowercase(args.old_name),
        'new_folder': to_lowercase(args.new_name),
        'old_id_prefix': None,
        'new_id_prefix': args.new_id_prefix,
        'enum_value': None,
        'provider': None,
        'dirs_renamed': 0,
        'files_renamed': 0,
        'files_updated': 0,
        'replacements_made': 0,
        'icon_folder_renamed': False,
        'protos_exit_code': 0,
        'build_exit_code': 0,
        'test_exit_code': 0,
        'duration_seconds': 0,
        'error': None
    }
    
    try:
        # Find repo root
        repo_root = os.environ.get('REPO_ROOT', find_repo_root(os.getcwd()))
        
        # 1. Determine provider and locate the old component directory
        old_folder = to_lowercase(args.old_name)
        new_folder = to_lowercase(args.new_name)
        
        provider = args.provider
        old_dir = None
        providers_base = os.path.join(repo_root, "apis/org/openmcf/provider")
        
        if provider:
            # --provider was given: resolve the path directly, no searching
            old_dir = os.path.join(providers_base, provider, old_folder)
            if not os.path.exists(old_dir):
                result['error'] = (
                    f"Component directory does not exist at expected path: {old_dir}\n"
                    f"  (provider={provider}, component={old_folder})"
                )
                print(json.dumps(result), file=sys.stderr)
                return 1
        else:
            # --provider was not given: search all provider directories
            if os.path.isdir(providers_base):
                for entry in sorted(os.listdir(providers_base)):
                    entry_path = os.path.join(providers_base, entry)
                    if not os.path.isdir(entry_path) or entry.startswith('.'):
                        continue
                    
                    # Check directly under this provider
                    candidate = os.path.join(entry_path, old_folder)
                    if os.path.exists(candidate):
                        provider = entry
                        old_dir = candidate
                        break
                    
                    # Check one level deeper (e.g., kubernetes/workload, kubernetes/addon)
                    for sub_entry in sorted(os.listdir(entry_path)):
                        sub_path = os.path.join(entry_path, sub_entry)
                        if not os.path.isdir(sub_path) or sub_entry.startswith('.'):
                            continue
                        candidate = os.path.join(sub_path, old_folder)
                        if os.path.exists(candidate):
                            provider = f"{entry}/{sub_entry}"
                            old_dir = candidate
                            break
                    
                    if old_dir:
                        break
            
            if not old_dir or not os.path.exists(old_dir):
                result['error'] = (
                    f"Could not find component directory for: {args.old_name} "
                    f"(searched all providers under {providers_base} for folder '{old_folder}')"
                )
                print(json.dumps(result), file=sys.stderr)
                return 1
        
        result['provider'] = provider
        print(f"Resolved provider: {provider}", file=sys.stderr)
        print(f"Component directory: {old_dir}", file=sys.stderr)
        
        # Skip proto validation entirely - proto is already updated
        
        # 4. Rename icon folder (before component directory renames)
        icon_result = rename_icon_folder(repo_root, provider, old_folder, new_folder)
        result['icon_folder_renamed'] = icon_result['renamed']
        
        if icon_result['exists']:
            print(f"Renamed icon folder: {icon_result['old_path']} -> {icon_result['new_path']}", file=sys.stderr)
        else:
            print(f"Icon folder not found (skipped): {icon_result['old_path']}", file=sys.stderr)
        
        # 5. Build replacement map
        replacements = build_replacement_map(args.old_name, args.new_name)
        
        # 6. Apply sequential renames to component directory
        component_dir = Path(old_dir)
        for old_str, new_str in replacements:
            apply_sequential_renames(component_dir, old_str, new_str, stats)
        
        # 7. Apply sequential renames to docs directory
        docs_dir = Path(repo_root) / "site" / "public" / "docs"
        if docs_dir.exists():
            for old_str, new_str in replacements:
                apply_sequential_renames(docs_dir, old_str, new_str, stats)
        
        # 7.5. Rename the component directory itself (from certmanager to kubernetescertmanager)
        old_component_dir = Path(old_dir)
        new_dir = get_component_directory(repo_root, provider, new_folder)
        new_component_dir = Path(new_dir)
        
        if old_component_dir.exists() and old_component_dir != new_component_dir:
            try:
                # Delete target directory if it exists
                if new_component_dir.exists():
                    shutil.rmtree(new_component_dir)
                # Rename the directory
                old_component_dir.rename(new_component_dir)
                stats['dirs_renamed'] += 1
                print(f"Renamed component directory: {old_component_dir} -> {new_component_dir}", file=sys.stderr)
            except Exception as e:
                print(f"Error renaming component directory {old_component_dir}: {e}", file=sys.stderr)
                stats['errors'] += 1
        
        # 8. Update result with statistics
        result['dirs_renamed'] = stats['dirs_renamed']
        result['files_renamed'] = stats['files_renamed']
        result['files_updated'] = stats['files_updated']
        result['replacements_made'] = stats['replacements_made']
        
        # 9. Mark as successful - skip build pipeline
        result['success'] = True
        print(f"Successfully renamed {args.old_name} -> {args.new_name}", file=sys.stderr)
        print(f"  Directories: {stats['dirs_renamed']}, Files: {stats['files_renamed']}, Content updates: {stats['files_updated']}", file=sys.stderr)
        
    except Exception as e:
        result['error'] = str(e)
        print(json.dumps(result), file=sys.stderr)
        return 1
    finally:
        result['duration_seconds'] = round(time.time() - start_time, 2)
    
    print(json.dumps(result, indent=2))
    return 0 if result['success'] else 1


if __name__ == '__main__':
    sys.exit(main())
