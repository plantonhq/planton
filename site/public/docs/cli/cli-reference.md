---
title: "CLI Reference"
description: "Complete command tree, flag reference, and exit codes for the openmcf CLI"
icon: "code"
order: 20
---

# CLI Reference

This page is the single authoritative reference for every command, flag, and option in the `openmcf` CLI. Other CLI documentation pages link here for flag details.

## Command Tree

The complete command tree, verified against source. Commands are grouped by purpose.

```
openmcf
│
├── Unified Commands (provisioner auto-detected from manifest)
│   ├── apply                 Deploy infrastructure
│   ├── plan                  Preview changes (alias: preview)
│   ├── init                  Initialize backend or stack
│   ├── destroy               Tear down infrastructure (alias: delete)
│   └── refresh               Sync state with cloud reality
│
├── Pulumi Commands
│   └── pulumi
│       ├── init              Initialize a new Pulumi stack
│       ├── preview           Preview infrastructure changes
│       ├── update            Deploy infrastructure (alias: up)
│       ├── destroy           Tear down infrastructure
│       ├── refresh           Sync state with cloud reality
│       ├── delete            Remove stack metadata (alias: rm)
│       └── cancel            Cancel in-progress stack operation
│
├── OpenTofu Commands
│   └── tofu
│       ├── init              Initialize backend and providers
│       ├── plan              Preview infrastructure changes
│       ├── apply             Deploy infrastructure
│       ├── destroy           Tear down infrastructure
│       ├── refresh           Sync state with cloud reality
│       ├── generate-variables Generate Terraform variables from manifest
│       └── load-tfvars       Load manifest into tfvars format
│
├── Terraform Commands
│   └── terraform
│       ├── init              Initialize backend and providers
│       ├── plan              Preview infrastructure changes
│       ├── apply             Deploy infrastructure
│       ├── destroy           Tear down infrastructure
│       └── refresh           Sync state with cloud reality
│
├── Module Management
│   ├── checkout              Check out a specific module version in staging
│   ├── pull                  Pull latest modules from upstream
│   ├── upgrade               Upgrade the CLI to latest or specified version
│   ├── downgrade             Install a previous CLI version
│   └── modules-version       Show current module version in staging
│
├── Manifest Utilities
│   ├── validate-manifest     Validate manifest against schema (alias: validate)
│   └── load-manifest         Load and display manifest with defaults (alias: load)
│
├── Configuration
│   └── config
│       ├── set               Set a configuration value
│       ├── get               Get a configuration value
│       └── list              List all configuration values
│
└── version                   Show CLI version and update status
```

## Global Flags

These flags are inherited by every subcommand.

| Flag | Default | Description |
|------|---------|-------------|
| `--local-module` | `false` | Use local openmcf git repository for IaC modules instead of downloading |
| `--openmcf-git-repo` | `~/scm/github.com/plantonhq/openmcf` | Path to local openmcf git repository (used with `--local-module`) |
| `-v`, `--version` | | Show version information (root command only) |

## Flag Reference

All flags below are organized by the group they belong to. The "Used by" column shows which commands accept each flag.

### Manifest Source Flags

These flags control where the manifest is loaded from. When multiple source flags are provided, the CLI uses the first one found in priority order: clipboard, stack-input, manifest file, input-dir, kustomize.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--manifest` | `-f` | | Path to the deployment-component manifest file |
| `--clipboard` | `-c` | `false` | Read manifest from system clipboard |
| `--stack-input` | `-i` | | Path to YAML file containing stack input (extracts manifest from `target` field) |
| `--input-dir` | | | Directory containing `target.yaml` and credential YAML files |
| `--kustomize-dir` | | | Directory containing kustomize configuration |
| `--overlay` | | | Kustomize overlay to use (e.g., `prod`, `dev`, `staging`). Requires `--kustomize-dir` |

Hidden aliases for `--clipboard`: `--clip`, `--cb`

**Used by**: All unified commands (`apply`, `plan`, `init`, `destroy`, `refresh`), `validate-manifest`, `load-manifest`.

**Resolution priority**: `--clipboard` > `--stack-input` > `--manifest` > `--input-dir` > `--kustomize-dir` + `--overlay`

### Execution Flags

These flags control how the IaC module is located and executed.

| Flag | Default | Description |
|------|---------|-------------|
| `--module-dir` | current directory | Directory containing the provisioner module |
| `--module-version` | | Check out a specific version (tag, branch, or SHA) of the IaC modules in the workspace copy |
| `--no-cleanup` | `false` | Do not clean up the workspace copy after execution |
| `--kube-context` | | kubectl context for Kubernetes deployments (overrides manifest label) |
| `--set` | | Override manifest values using `key=value` pairs (repeatable) |
| `--local-module` | `false` | Use the local openmcf repository to derive the module directory |

**Used by**: All unified commands (`apply`, `plan`, `init`, `destroy`, `refresh`).

### Provider Config Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--provider-config` | `-p` | | Path to provider credentials file. Provider type is auto-detected from the manifest's `apiVersion` and `kind` |

**Used by**: All unified commands, all `pulumi` subcommands, all `tofu` subcommands, all `terraform` subcommands.

### Pulumi-Specific Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--stack` | | Pulumi stack FQDN in the format `<org>/<project>/<stack>` |
| `--yes` | `false` | Automatically approve and perform the update after previewing (Pulumi) |
| `--diff` | `false` | Show detailed resource diffs (Pulumi) |
| `--force` | `false` | Force stack removal even if resources exist (`pulumi delete` only) |

**Used by**: All unified commands (Pulumi flags are passed through when provisioner is Pulumi), all `pulumi` subcommands.

### OpenTofu / Terraform Flags

#### Apply and Destroy

| Flag | Default | Description |
|------|---------|-------------|
| `--auto-approve` | `false` | Skip interactive approval of plan before applying |

**Used by**: Unified `apply` and `destroy`, `tofu apply`, `tofu destroy`, `terraform apply`, `terraform destroy`.

#### Plan

| Flag | Default | Description |
|------|---------|-------------|
| `--destroy` | `false` | Create a destroy plan instead of an apply plan |

**Used by**: Unified `plan`, `tofu plan`, `terraform plan`.

#### Init (Backend Configuration)

| Flag | Default | Description |
|------|---------|-------------|
| `--backend-type` | `local` | Backend type: `local`, `s3`, `gcs`, `azurerm`, and others |
| `--backend-bucket` | | State bucket name (S3/GCS) or container name (Azure) |
| `--backend-key` | | State file path within bucket (e.g., `env/prod/terraform.tfstate`) |
| `--backend-region` | | Region for S3 backend (use `auto` for S3-compatible backends like R2) |
| `--backend-endpoint` | | Custom S3-compatible endpoint URL (required for R2, MinIO, etc.) |
| `--backend-config` | | Additional backend configuration `key=value` pairs (repeatable) |
| `--reconfigure` | `false` | Reconfigure backend, ignoring any saved configuration |

**Used by**: Unified `init`. For direct `tofu init` and `terraform init`, only `--backend-type`, `--backend-config`, and `--module-version` are available.

### Engine-Specific Persistent Flags (Direct Commands)

When using `openmcf tofu` or `openmcf terraform` directly (not unified commands), each engine has its own set of persistent flags:

| Flag | Short | Description |
|------|-------|-------------|
| `--manifest` | | Path to the deployment-component manifest file |
| `--input-dir` | | Directory containing `target.yaml` and credential YAML files |
| `--kustomize-dir` | | Directory containing kustomize configuration |
| `--overlay` | | Kustomize overlay |
| `--module-dir` | | Directory containing the module (default: current directory) |
| `--set` | | Override manifest values using `key=value` pairs |
| `--provider-config` | `-p` | Path to provider credentials file |

These are registered on the `tofu` and `terraform` parent commands and inherited by all their subcommands.

**Note**: Direct engine commands do not support `--clipboard` or `--stack-input` manifest sources. Use the unified commands for those input methods.

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Error: validation failure, deployment failure, missing prerequisites, or invalid flags |

## File System Paths

The CLI uses the following directories under `~/.openmcf/`:

| Path | Purpose |
|------|---------|
| `~/.openmcf/config.yaml` | CLI configuration file |
| `~/.openmcf/staging/openmcf/` | Cloned openmcf repository for module resolution |
| `~/.openmcf/staging/.version` | Current module version in staging |
| `~/.openmcf/pulumi/binaries/{version}/` | Cached pre-built Pulumi module binaries |
| `~/.openmcf/pulumi/staging-workspaces/{stack}/` | Pulumi workspace copies per stack |
| `~/.openmcf/terraform/modules/{version}/{component}/` | Cached Terraform/OpenTofu module archives |
| `~/.openmcf/downloads/` | Downloaded URL manifests |

## Getting Help

```bash
# General help
openmcf --help

# Command-specific help
openmcf apply --help
openmcf pulumi --help
openmcf tofu init --help
openmcf terraform apply --help
```

Every command and subcommand supports `--help` to display its usage, flags, and description.

## What's Next

- **[Unified Commands](./unified-commands)** — How provisioner auto-detection works and when to use unified vs. direct commands
- **[Pulumi Commands](./pulumi-commands)** — Pulumi-specific subcommands and workflows
- **[OpenTofu Commands](./tofu-commands)** — OpenTofu-specific subcommands and workflows
- **[Terraform Commands](./terraform-commands)** — Terraform-specific subcommands and workflows
- **[Module Management](./module-management)** — Module versioning, caching, and the staging area
- **[Configuration](./configuration)** — Config, validation, manifest loading, and version utilities
