#!/usr/bin/env python3
"""
Deterministic PR creation utility for Planton repository.

This script creates or switches to the given branch, stages and commits any
changes with the provided commit message (if there are changes), pushes the
branch, and opens a GitHub Pull Request using gh.

Inputs are explicit to avoid flaky LLM-to-shell orchestration:
  - --title:          PR title
  - --branch-name:    Branch to create/switch to (head)
  - --commit-message: Commit message to use if there are staged/unstaged changes
  - --body | --body-file: PR description (string or file)

Optional inputs:
  - --base-branch: Base branch for the PR (defaults to repo default branch)
  - --draft:       Create the PR as a draft
  - --reviewers:   Comma-separated GitHub usernames for review
  - --labels:      Comma-separated labels to apply

Example:
  python3 tools/local-dev/create_pull_request.py \
    --title "feat(api): add user search endpoint" \
    --branch-name "feat/api-add-user-search-endpoint" \
    --commit-message "feat(api): add user search endpoint" \
    --body-file /path/to/body.md \
    --reviewers user1,user2 --labels "area/api,triage"
"""

from __future__ import annotations

import argparse
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Create a GitHub Pull Request deterministically using gh.",
        formatter_class=argparse.ArgumentDefaultsHelpFormatter,
    )
    parser.add_argument("--title", required=True, help="PR title")
    parser.add_argument(
        "--branch-name",
        required=True,
        dest="branch_name",
        help="Branch name to create/switch to",
    )
    parser.add_argument(
        "--commit-message",
        required=True,
        dest="commit_message",
        help="Commit message to use when committing local changes",
    )
    body_group = parser.add_mutually_exclusive_group(required=True)
    body_group.add_argument("--body", help="PR description text")
    body_group.add_argument("--body-file", help="Path to file with PR description")
    parser.add_argument(
        "--base-branch",
        dest="base_branch",
        default=None,
        help="Base branch to target; auto-detected if omitted",
    )
    parser.add_argument("--draft", action="store_true", help="Create as draft PR")
    parser.add_argument(
        "--reviewers",
        default=None,
        help="Comma-separated GitHub usernames to request review from",
    )
    parser.add_argument(
        "--labels",
        default=None,
        help="Comma-separated labels to apply to the PR",
    )
    return parser.parse_args()


def run(cmd: list[str], cwd: Path | None = None, capture: bool = False) -> subprocess.CompletedProcess:
    return subprocess.run(
        cmd,
        cwd=str(cwd) if cwd else None,
        check=True,
        capture_output=capture,
        text=True,
    )


def check_prerequisites() -> None:
    if shutil.which("gh") is None:
        print("Error: GitHub CLI (gh) is not installed. Install with: brew install gh", file=sys.stderr)
        sys.exit(1)
    try:
        subprocess.run(["gh", "auth", "status"], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    except subprocess.CalledProcessError:
        print("Error: gh is not authenticated. Run: gh auth login", file=sys.stderr)
        sys.exit(1)
    try:
        subprocess.run(["git", "rev-parse", "--is-inside-work-tree"], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    except subprocess.CalledProcessError:
        print("Error: not inside a git repository.", file=sys.stderr)
        sys.exit(1)


def get_repo_root() -> Path:
    try:
        proc = run(["git", "rev-parse", "--show-toplevel"], capture=True)
        return Path(proc.stdout.strip())
    except Exception:
        return Path.cwd()


def detect_base_branch(repo_root: Path) -> str:
    # Try gh defaultBranchRef
    try:
        proc = run(["gh", "repo", "view", "--json", "defaultBranchRef", "-q", ".defaultBranchRef.name"], cwd=repo_root, capture=True)
        branch = proc.stdout.strip()
        if branch:
            return branch
    except Exception:
        pass
    # Try origin/HEAD
    try:
        proc = run(["git", "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"], cwd=repo_root, capture=True)
        ref = proc.stdout.strip()
        if ref.startswith("origin/"):
            return ref[len("origin/"):]
        if ref:
            return ref
    except Exception:
        pass
    return "main"


def create_or_switch_branch(repo_root: Path, branch_name: str) -> None:
    # Check if branch exists
    exists = subprocess.run(["git", "rev-parse", "--verify", "-q", branch_name], cwd=str(repo_root)).returncode == 0
    if exists:
        print(f"Switching to existing branch: {branch_name}")
        run(["git", "checkout", branch_name], cwd=repo_root)
    else:
        print(f"Creating and switching to new branch: {branch_name}")
        run(["git", "checkout", "-b", branch_name], cwd=repo_root)


def commit_changes_if_any(repo_root: Path, commit_message: str) -> None:
    """
    Stage all changes (including untracked files) and commit if anything is staged.

    This avoids missing untracked files (e.g., brand new docs) which `git diff` does not report.
    """
    # Always attempt to stage everything first so untracked files are included
    print("Staging all changes...")
    run(["git", "add", "-A"], cwd=repo_root)

    # Determine if there is anything staged to commit
    staged_after_add = subprocess.run(["git", "diff", "--cached", "--quiet"], cwd=str(repo_root)).returncode != 0
    if staged_after_add:
        print("Committing changes...")
        run(["git", "commit", "-m", commit_message], cwd=repo_root)
    else:
        print("No changes detected; skipping commit.")


def push_branch(repo_root: Path, branch_name: str) -> None:
    # Determine if upstream is set
    has_upstream = subprocess.run(["git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"], cwd=str(repo_root), stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL).returncode == 0
    if has_upstream:
        print("Pushing branch (existing upstream)...")
        run(["git", "push"], cwd=repo_root)
    else:
        print("Pushing branch and setting upstream...")
        run(["git", "push", "-u", "origin", branch_name], cwd=repo_root)


def create_pr(
    repo_root: Path,
    base_branch: str,
    head_branch: str,
    title: str,
    body_text: str,
    draft: bool,
    reviewers: str | None,
    labels: str | None,
) -> None:
    # Write body to a temporary file to avoid quoting issues
    with tempfile.NamedTemporaryFile("w", delete=False) as tf:
        tf.write(body_text)
        body_file = tf.name

    cmd: list[str] = [
        "gh",
        "pr",
        "create",
        "--base",
        base_branch,
        "--head",
        head_branch,
        "--title",
        title,
        "--body-file",
        body_file,
    ]
    if draft:
        cmd.append("--draft")
    if reviewers:
        cmd.extend(["--reviewer", reviewers])
    if labels:
        cmd.extend(["--label", labels])

    print(f"Creating PR: base={base_branch} head={head_branch}")
    try:
        cp = subprocess.run(cmd, cwd=str(repo_root), check=True, capture_output=True, text=True)
        if cp.stdout:
            print(cp.stdout, end="")
        if cp.stderr:
            print(cp.stderr, end="", file=sys.stderr)
    except subprocess.CalledProcessError as exc:
        print("gh pr create failed.", file=sys.stderr)
        if getattr(exc, "stdout", None):
            print(exc.stdout, end="")
        if getattr(exc, "stderr", None):
            print(exc.stderr, end="", file=sys.stderr)
        raise
    finally:
        try:
            Path(body_file).unlink(missing_ok=True)
        except Exception:
            pass


def print_pr_url(repo_root: Path) -> None:
    try:
        proc = run(["gh", "pr", "view", "--json", "url", "-q", ".url"], cwd=repo_root, capture=True)
        url = proc.stdout.strip()
        if url:
            print(url)
    except Exception:
        # Non-fatal if we can't print the URL
        pass


def main() -> None:
    args = parse_args()
    check_prerequisites()
    repo_root = get_repo_root()

    title: str = args.title
    branch_name: str = args.branch_name
    commit_message: str = args.commit_message
    draft: bool = args.draft
    reviewers: str | None = args.reviewers
    labels: str | None = args.labels

    if args.body_file:
        try:
            body_text = Path(args.body_file).read_text(encoding="utf-8")
        except Exception as e:
            print(f"Error reading body file: {e}", file=sys.stderr)
            sys.exit(1)
    else:
        body_text = args.body or ""

    base_branch = args.base_branch or detect_base_branch(repo_root)

    try:
        create_or_switch_branch(repo_root, branch_name)
        commit_changes_if_any(repo_root, commit_message)
        push_branch(repo_root, branch_name)
        create_pr(
            repo_root=repo_root,
            base_branch=base_branch,
            head_branch=branch_name,
            title=title,
            body_text=body_text,
            draft=draft,
            reviewers=reviewers,
            labels=labels,
        )
        print_pr_url(repo_root)
    except subprocess.CalledProcessError as exc:
        print(f"Command failed with exit code {exc.returncode}: {' '.join(exc.cmd) if hasattr(exc, 'cmd') else ''}", file=sys.stderr)
        sys.exit(exc.returncode or 1)


if __name__ == "__main__":
    main()




