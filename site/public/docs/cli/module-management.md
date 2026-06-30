---
title: "Module Management"
description: "Manage IaC module versions, the staging area, and CLI upgrades"
icon: "package"
order: 70
---

# Module Management

Planton deployment components ship as IaC modules — Pulumi programs (Go) and HCL configurations (Terraform/OpenTofu) — that are resolved and executed at runtime. This page covers the commands that manage those modules and the CLI binary itself.

## Module Resolution

When you run a deployment command, the CLI needs to locate the correct IaC module for the component kind in your manifest. The resolution chain differs by engine.

### Pulumi Module Resolution

1. **Direct path**: If `--module-dir` points to a directory containing `Pulumi.yaml`, use it directly
2. **Pre-built binary**: Download a pre-compiled binary from GitHub releases to `~/.planton/pulumi/binaries/{version}/`
3. **Staging area**: Copy from `~/.planton/staging/planton/` to an isolated workspace at `~/.planton/pulumi/staging-workspaces/{stack}/`

### OpenTofu / Terraform Module Resolution

1. **Direct path**: If `--module-dir` points to a directory containing `.tf` files, use it directly
2. **Release archive**: Download `terraform-{component}.zip` from GitHub releases to `~/.planton/terraform/modules/{version}/{component}/`
3. **Staging area**: Copy from `~/.planton/staging/planton/` to an isolated workspace

For development CLI versions, the binary/archive download step is skipped and the CLI falls back to the staging area.

### Local Module Override

The `--local-module` global flag bypasses the entire resolution chain and derives the module path from a local Planton repository checkout:

- Pulumi: `{repo}/apis/dev/planton/provider/{provider}/{kind}/v1/iac/pulumi`
- Terraform/OpenTofu: `{repo}/apis/dev/planton/provider/{provider}/{kind}/v1/iac/tf`

The repository path defaults to `~/scm/github.com/plantonhq/planton` and can be overridden with `--planton-git-repo`.

## The Staging Area

The staging area at `~/.planton/staging/planton/` is a local clone of the Planton repository. It serves as the fallback module source when pre-built binaries or archives are not available.

The staging area is created automatically the first time it is needed. You can manage it explicitly with the commands below.

### checkout

Switch the staging area to a specific version. The version argument can be a release tag, branch name, commit SHA, or the special value `latest`.

```bash
planton checkout latest
planton checkout v0.2.273
planton checkout main
planton checkout abc1234
```

| Argument | Description |
|----------|-------------|
| `latest` | Resolves to the most recent release tag |
| `v0.2.273` | A specific release tag |
| `main` | A branch name |
| `abc1234` | A commit SHA |

If the staging area does not exist, it is cloned first.

Use `checkout` when you need IaC modules from a specific release, when testing newer modules before upgrading the CLI, or when rolling back to an older version for compatibility.

### pull

Fetch and pull the latest changes from the upstream repository into the staging area.

```bash
planton pull
```

This runs `git fetch --all` followed by `git pull` in the staging directory. If the staging area does not exist, it is cloned first. After pulling, the command displays the staging path and current version.

Run `pull` periodically to ensure you have access to the latest deployment components and bug fixes.

### modules-version

Display the current version of IaC modules in the staging area.

```bash
planton modules-version
```

Reads the version from `~/.planton/staging/.version` and displays the staging location and current version. If no staging area exists, the command suggests running `pull` to set it up.

## Version Pinning with --module-version

The `--module-version` flag, available on deployment commands, checks out a specific version in the workspace copy without changing the staging area. This isolates version changes to a single execution.

```bash
planton apply -f database.yaml --module-version v0.2.270
planton tofu apply --manifest database.yaml --module-version main
```

This is useful for testing a specific module version without permanently switching the staging area.

## CLI Version Management

These commands manage the `planton` CLI binary itself, not the IaC modules.

### upgrade

Upgrade the CLI to the latest available version or to a specific version.

```bash
planton upgrade
planton upgrade v0.3.10-cli.20260110.0
planton upgrade --check
planton upgrade --force
```

| Flag | Short | Description |
|------|-------|-------------|
| `--check` | `-c` | Check for updates without installing |
| `--force` | `-f` | Force upgrade even if already on the latest or specified version |

On macOS with Homebrew, `upgrade` without a version argument uses `brew upgrade --cask`. When a specific version is provided, or on other platforms, the binary is downloaded directly from GitHub releases.

### downgrade

Install a specific previous version of the CLI.

```bash
planton downgrade v0.3.5-cli.20260108.0
planton downgrade v0.3.5-cli.20260108.0 --force
```

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Force install even if already on the specified version |

The version argument is required. The binary is downloaded directly from GitHub releases.

## Workspace Isolation

When the CLI copies modules from the staging area or downloads them from releases, it creates isolated workspace copies. This ensures that concurrent deployments of different components or versions do not interfere with each other.

- **Pulumi workspaces**: `~/.planton/pulumi/staging-workspaces/{stack-fqdn}/`
- **Tofu/Terraform workspaces**: Temporary directories that are cleaned up after execution

The `--no-cleanup` flag preserves workspace copies after execution for debugging. Without this flag, workspaces are automatically cleaned up.

## What's Next

- **[CLI Reference](./cli-reference)** — Complete flag reference including `--module-dir`, `--module-version`, `--no-cleanup`, `--local-module`
- **[Configuration](./configuration)** — Config, validation, and manifest loading utilities
