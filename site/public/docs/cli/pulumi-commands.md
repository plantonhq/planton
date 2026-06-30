---
title: "Pulumi Commands"
description: "Deploy, preview, and manage infrastructure using Pulumi as the IaC engine"
icon: "code"
order: 40
---

# Pulumi Commands

The `planton pulumi` command group runs infrastructure operations using Pulumi as the IaC engine. Each deployment component in Planton ships with a Pulumi module written in Go that translates the manifest spec into cloud resources.

## Prerequisites

Pulumi commands require one of these execution modes:

- **Pre-built binary** (default): The CLI downloads pre-built Pulumi module binaries from GitHub releases. No local toolchain required.
- **Staging area**: If no pre-built binary is available, the CLI uses the cloned Planton repository at `~/.planton/staging/planton/` and builds the module. Requires Go installed.
- **Local module** (`--local-module`): Points to a local checkout of the Planton repository for development.

Pulumi state must be stored in a Pulumi backend. Configure your backend with `PULUMI_BACKEND_URL` or `pulumi login` before running commands.

## Subcommands

### init

Create a new Pulumi stack for the resource defined in the manifest. Extracts the stack FQDN from the manifest labels and creates the stack in your configured Pulumi backend.

```bash
planton pulumi init --manifest database.yaml
planton pulumi init --manifest database.yaml --stack my-org/my-project/dev
```

If `--stack` is not provided, the stack FQDN is derived from the manifest's metadata labels.

### preview

Preview infrastructure changes without applying them. Shows what resources would be created, updated, or deleted.

```bash
planton pulumi preview --manifest database.yaml
planton pulumi preview --manifest database.yaml --diff
planton pulumi preview --manifest api.yaml --set spec.container.replicas=5
```

The `--diff` flag shows detailed property-level diffs for each resource change.

### update

Deploy infrastructure by creating, updating, or replacing resources to match the manifest spec.

```bash
planton pulumi update --manifest database.yaml
planton pulumi up --manifest database.yaml                    # alias
planton pulumi up --manifest database.yaml --yes              # skip confirmation
planton pulumi up --manifest api.yaml --set spec.version=v2
```

**Alias**: `up`

By default, `update` shows a preview and prompts for confirmation before proceeding. Pass `--yes` to skip the confirmation and apply immediately.

### destroy

Tear down all resources managed by the stack. The resources defined in the manifest are deleted from the cloud provider.

```bash
planton pulumi destroy --manifest database.yaml
planton pulumi destroy --manifest database.yaml --yes
```

This removes cloud resources but preserves the stack metadata. To also remove the stack, run `delete` after `destroy`.

### delete

Remove the Pulumi stack metadata from the backend. This permanently deletes the stack's configuration and state.

```bash
planton pulumi delete --manifest database.yaml
planton pulumi rm --manifest database.yaml                    # alias
planton pulumi delete --manifest database.yaml --force
```

**Alias**: `rm`

This command does NOT destroy cloud resources. If the stack still has deployed resources, run `destroy` first. Use `--force` to remove the stack even if it still references resources. This is intended for cases where resources were already cleaned up manually or through other means.

### cancel

Cancel an in-progress Pulumi stack operation. Use this when a stack is locked due to a crashed or interrupted operation.

```bash
planton pulumi cancel --manifest database.yaml
planton pulumi cancel --manifest database.yaml --stack my-org/my-project/dev
```

This unlocks the stack so you can run other operations. It does not roll back any changes that were partially applied.

### refresh

Sync the Pulumi state with the actual cloud state without modifying any resources.

```bash
planton pulumi refresh --manifest database.yaml
planton pulumi refresh --manifest database.yaml --diff
```

This queries the cloud provider for the current state of all managed resources and updates the state file to match reality. No resources are created, modified, or deleted.

## Flags

All `planton pulumi` subcommands inherit persistent flags from the parent command. Unlike [unified commands](./unified-commands), direct Pulumi commands register `--manifest` without the `-f` shorthand. Use `--manifest` for the full flag name.

### Manifest and Input

| Flag | Short | Description |
|------|-------|-------------|
| `--manifest` | | Path to the deployment-component manifest file |
| `--stack-input` | `-i` | Path to YAML file containing stack input (extracts manifest from `target` field) |
| `--input-dir` | | Directory containing `target.yaml` and credential YAML files |
| `--kustomize-dir` | | Directory containing kustomize configuration |
| `--overlay` | | Kustomize overlay (e.g., `prod`, `dev`, `staging`) |

### Execution

| Flag | Description |
|------|-------------|
| `--module-dir` | Directory containing the Pulumi module (default: current directory) |
| `--module-version` | Check out a specific version of IaC modules in the workspace copy |
| `--no-cleanup` | Keep the workspace copy after execution |
| `--kube-context` | kubectl context for Kubernetes deployments (overrides manifest label) |
| `--set` | Override manifest values using `key=value` pairs |
| `--provider-config`, `-p` | Path to provider credentials file |

### Pulumi-Specific

| Flag | Description |
|------|-------------|
| `--stack` | Pulumi stack FQDN: `<org>/<project>/<stack>` |
| `--yes` | Auto-approve operations without confirmation |
| `--diff` | Show detailed property-level resource diffs |
| `--force` | Force stack removal even if resources exist (`delete`/`rm` only) |

## Typical Workflow

A standard Pulumi deployment lifecycle with Planton:

```bash
# 1. Initialize the stack
planton pulumi init --manifest database.yaml

# 2. Preview what will be created
planton pulumi preview --manifest database.yaml

# 3. Deploy
planton pulumi up --manifest database.yaml --yes

# 4. Make changes to the manifest, then preview and update
planton pulumi preview --manifest database.yaml --diff
planton pulumi up --manifest database.yaml --yes

# 5. When done, tear down resources
planton pulumi destroy --manifest database.yaml --yes

# 6. Remove the stack metadata
planton pulumi delete --manifest database.yaml
```

## What's Next

- **[CLI Reference](./cli-reference)** — Complete flag reference
- **[Unified Commands](./unified-commands)** — Provisioner-agnostic alternative to direct Pulumi commands
- **[OpenTofu Commands](./tofu-commands)** — OpenTofu engine commands
- **[Module Management](./module-management)** — Module versioning and the staging area
