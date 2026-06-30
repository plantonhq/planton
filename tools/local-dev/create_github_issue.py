#!/usr/bin/env python3
"""
Deterministic GitHub issue creation utility for Planton.

This script creates a GitHub issue using the GitHub CLI (gh) with the provided
title, body, and optional metadata (labels, assignees, milestone, project).

Inputs are explicit to avoid flaky LLM-to-shell orchestration:
  - --title:     Issue title
  - --body:      Issue description text (string)
  - --body-file: Issue description from file (alternative to --body)

Optional inputs:
  - --labels:    Comma-separated labels to apply
  - --assignees: Comma-separated GitHub usernames to assign
  - --milestone: Milestone name or number
  - --project:   Project name or number
  - --web:       Open the created issue in browser

Example:
  python3 tools/local-dev/create_github_issue.py \
    --title "Postgres deployment component spec validation broken" \
    --body-file /path/to/issue-body.md \
    --labels "bug,area/deployment-component,priority/high" \
    --assignees "user1,user2" \
    --web
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
        description="Create a GitHub Issue deterministically using gh.",
        formatter_class=argparse.ArgumentDefaultsHelpFormatter,
    )
    parser.add_argument("--title", required=True, help="Issue title")
    
    body_group = parser.add_mutually_exclusive_group(required=True)
    body_group.add_argument("--body", help="Issue description text")
    body_group.add_argument("--body-file", help="Path to file with issue description")
    
    parser.add_argument(
        "--labels",
        default=None,
        help="Comma-separated labels to apply to the issue",
    )
    parser.add_argument(
        "--assignees",
        default=None,
        help="Comma-separated GitHub usernames to assign to the issue",
    )
    parser.add_argument(
        "--milestone",
        default=None,
        help="Milestone name or number to add the issue to",
    )
    parser.add_argument(
        "--project",
        default=None,
        help="Project name or number to add the issue to",
    )
    parser.add_argument(
        "--web",
        action="store_true",
        help="Open the created issue in browser",
    )
    return parser.parse_args()


def run(cmd: list[str], cwd: Path | None = None, capture: bool = False) -> subprocess.CompletedProcess:
    """Run a command and return the result."""
    return subprocess.run(
        cmd,
        cwd=str(cwd) if cwd else None,
        check=True,
        capture_output=capture,
        text=True,
    )


def check_prerequisites() -> None:
    """Check that required tools are installed and configured."""
    if shutil.which("gh") is None:
        print("Error: GitHub CLI (gh) is not installed. Install with: brew install gh", file=sys.stderr)
        sys.exit(1)
    try:
        subprocess.run(
            ["gh", "auth", "status"],
            check=True,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
    except subprocess.CalledProcessError:
        print("Error: gh is not authenticated. Run: gh auth login", file=sys.stderr)
        sys.exit(1)
    try:
        subprocess.run(
            ["git", "rev-parse", "--is-inside-work-tree"],
            check=True,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
    except subprocess.CalledProcessError:
        print("Error: not inside a git repository.", file=sys.stderr)
        sys.exit(1)


def get_repo_root() -> Path:
    """Get the root directory of the git repository."""
    try:
        proc = run(["git", "rev-parse", "--show-toplevel"], capture=True)
        return Path(proc.stdout.strip())
    except Exception:
        return Path.cwd()


def create_issue(
    repo_root: Path,
    title: str,
    body_text: str,
    labels: str | None,
    assignees: str | None,
    milestone: str | None,
    project: str | None,
    web: bool,
) -> str:
    """
    Create a GitHub issue using gh CLI.
    
    Returns the issue URL.
    """
    # Write body to a temporary file to avoid quoting issues
    with tempfile.NamedTemporaryFile("w", delete=False, suffix=".md") as tf:
        tf.write(body_text)
        body_file = tf.name

    cmd: list[str] = [
        "gh",
        "issue",
        "create",
        "--title",
        title,
        "--body-file",
        body_file,
    ]
    
    if labels:
        cmd.extend(["--label", labels])
    if assignees:
        cmd.extend(["--assignee", assignees])
    if milestone:
        cmd.extend(["--milestone", milestone])
    if project:
        cmd.extend(["--project", project])
    if web:
        cmd.append("--web")

    print(f"Creating GitHub issue: {title}")
    try:
        cp = subprocess.run(
            cmd,
            cwd=str(repo_root),
            check=True,
            capture_output=True,
            text=True,
        )
        
        # Extract issue URL from output
        issue_url = ""
        if cp.stdout:
            # gh issue create outputs the URL on the last line typically
            for line in cp.stdout.strip().split("\n"):
                if line.startswith("http"):
                    issue_url = line.strip()
            print(cp.stdout, end="")
        if cp.stderr:
            print(cp.stderr, end="", file=sys.stderr)
            
        return issue_url
    except subprocess.CalledProcessError as exc:
        print("gh issue create failed.", file=sys.stderr)
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


def print_issue_summary(title: str, labels: str | None, assignees: str | None) -> None:
    """Print a summary of the issue that was created."""
    print("\n" + "="*60)
    print("✅ GitHub Issue Created Successfully!")
    print("="*60)
    print(f"Title: {title}")
    if labels:
        print(f"Labels: {labels}")
    if assignees:
        print(f"Assignees: {assignees}")
    print("="*60 + "\n")


def main() -> None:
    args = parse_args()
    check_prerequisites()
    repo_root = get_repo_root()

    title: str = args.title
    labels: str | None = args.labels
    assignees: str | None = args.assignees
    milestone: str | None = args.milestone
    project: str | None = args.project
    web: bool = args.web

    # Read body text from file or argument
    if args.body_file:
        try:
            body_text = Path(args.body_file).read_text(encoding="utf-8")
        except Exception as e:
            print(f"Error reading body file: {e}", file=sys.stderr)
            sys.exit(1)
    else:
        body_text = args.body or ""

    try:
        issue_url = create_issue(
            repo_root=repo_root,
            title=title,
            body_text=body_text,
            labels=labels,
            assignees=assignees,
            milestone=milestone,
            project=project,
            web=web,
        )
        
        print_issue_summary(title, labels, assignees)
        
        if issue_url:
            print(f"Issue URL: {issue_url}")
            
    except subprocess.CalledProcessError as exc:
        print(
            f"Command failed with exit code {exc.returncode}: "
            f"{' '.join(exc.cmd) if hasattr(exc, 'cmd') else ''}",
            file=sys.stderr,
        )
        sys.exit(exc.returncode or 1)


if __name__ == "__main__":
    main()
