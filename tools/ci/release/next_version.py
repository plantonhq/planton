#!/usr/bin/env python3
"""Calculate the next semantic version from git tags.

Every release in this repository is a tag named by what it releases and its
version: the bare `vX.Y.Z` releases the repository (the catalog),
`operator/vX.Y.Z` the operator, `helm/<chart>/vX.Y.Z` a Helm chart. This
script finds the newest strict-semver tag under one such prefix and bumps it,
so each `make release-*` front door for Planton's own lines computes its next
version the same way. A third-party line (`agent-substrate/vX.Y.Z`) carries
upstream's version and is never bumped here.

Usage:
    next_version.py [patch|minor|major] [--prefix operator/]

With no prefix it reads the bare `vX.Y.Z` tags. With a prefix it reads
`<prefix>vX.Y.Z` tags and prints the bare next version (`vX.Y.Z`); the caller
prepends the prefix when it creates the tag. A namespace with no tag yet
yields v0.0.1 for a patch bump, so a line that must start elsewhere (a chart
that already has published versions from before it was tagged) cuts its first
tag with an explicit version.
"""
import argparse
import re
import subprocess
import sys
from typing import Iterable

# Strict semver pattern: vX.Y.Z where X, Y, Z are digits only. Pre-release and
# build-metadata tags (the auto-release tags carry `+`) never match.
SEMVER_PATTERN = re.compile(r"^v(\d+)\.(\d+)\.(\d+)$")


def list_tags(prefix: str) -> list[str]:
    """List the repository's tags under the prefix, newest version first."""
    result = subprocess.run(
        ["git", "tag", "--list", f"{prefix}v*", "--sort=-v:refname"],
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        return []
    return [tag.strip() for tag in result.stdout.splitlines() if tag.strip()]


def latest_version(tags: Iterable[str], prefix: str) -> str:
    """The newest strict-semver version among the tags, without the prefix.

    Tags are expected newest-first (git's version sort); the first one whose
    remainder is strict semver wins. A namespace with no such tag is v0.0.0.
    """
    for tag in tags:
        if not tag.startswith(prefix):
            continue
        version = tag[len(prefix):]
        if SEMVER_PATTERN.match(version):
            return version
    return "v0.0.0"


def bump_version(current: str, bump_type: str) -> str:
    """Bump the version according to semver rules."""
    match = SEMVER_PATTERN.match(current)
    if not match:
        raise ValueError(f"Invalid version format: {current}")
    major, minor, patch = int(match.group(1)), int(match.group(2)), int(match.group(3))
    if bump_type == "major":
        return f"v{major + 1}.0.0"
    if bump_type == "minor":
        return f"v{major}.{minor + 1}.0"
    if bump_type == "patch":
        return f"v{major}.{minor}.{patch + 1}"
    raise ValueError(f"Invalid bump type: {bump_type}. Use major, minor, or patch")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("bump", nargs="?", default="patch", choices=("patch", "minor", "major"))
    parser.add_argument(
        "--prefix",
        default="",
        help="tag namespace to read and bump, e.g. operator/ or helm/planton/ (default: the bare vX.Y.Z tags)",
    )
    args = parser.parse_args()
    if args.prefix and not args.prefix.endswith("/"):
        print(f"Error: --prefix must end with '/' (got {args.prefix!r})", file=sys.stderr)
        sys.exit(1)

    current = latest_version(list_tags(args.prefix), args.prefix)
    print(bump_version(current, args.bump))


if __name__ == "__main__":
    main()
