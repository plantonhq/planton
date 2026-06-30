---
title: "OpenTofu Commands"
description: "Deploy, preview, and manage infrastructure using OpenTofu as the IaC engine"
icon: "code"
order: 50
---

# OpenTofu Commands

The `planton tofu` command group runs infrastructure operations using OpenTofu as the IaC engine. Each deployment component in Planton ships with an HCL module in `iac/tf/` that translates the manifest spec into cloud resources.

## Prerequisites

- **OpenTofu binary**: The `tofu` binary must be installed and available on your `PATH`. Install from [opentofu.org](https://opentofu.org/docs/intro/install/).
- **Module resolution**: The CLI resolves HCL modules through the module resolution chain: direct path, GitHub release archive, or staging area. See [Module Management](./module-management) for details.

## Subcommands

### init

Initialize the backend configuration and download required providers. This must be run before any other operation on a new component or after changing the backend configuration.

```bash
planton tofu init --manifest database.yaml
planton tofu init --manifest database.yaml --backend-type s3 --backend-config bucket=my-state
planton tofu init --manifest database.yaml --backend-type gcs --backend-config bucket=my-state
```

The `--backend-type` defaults to `local`. To use remote state storage, pass `--backend-type` with one of the supported backend types (`s3`, `gcs`, `azurerm`, etc.) and provide additional configuration via `--backend-config` key-value pairs.

| Flag | Default | Description |
|------|---------|-------------|
| `--backend-type` | `local` | Backend type: `local`, `s3`, `gcs`, `azurerm`, etc. |
| `--backend-config` | | Backend configuration `key=value` pairs (repeatable) |
| `--module-version` | | Check out a specific version of the IaC modules |

### plan

Preview infrastructure changes without applying them. Shows what resources would be created, updated, or deleted.

```bash
planton tofu plan --manifest database.yaml
planton tofu plan --manifest database.yaml --destroy
```

| Flag | Default | Description |
|------|---------|-------------|
| `--destroy` | `false` | Create a destroy plan instead of an apply plan |
| `--module-version` | | Check out a specific version of the IaC modules |
| `--no-cleanup` | `false` | Keep workspace copy after execution |

### apply

Deploy infrastructure by applying the planned changes. Creates, updates, or replaces resources to match the manifest spec.

```bash
planton tofu apply --manifest database.yaml
planton tofu apply --manifest database.yaml --auto-approve
planton tofu apply --manifest api.yaml --set spec.container.replicas=5
```

By default, `apply` shows a plan and prompts for confirmation. Pass `--auto-approve` to skip the prompt.

| Flag | Default | Description |
|------|---------|-------------|
| `--auto-approve` | `false` | Skip interactive approval before applying |
| `--module-version` | | Check out a specific version of the IaC modules |
| `--no-cleanup` | `false` | Keep workspace copy after execution |

### destroy

Tear down all resources managed by the current state. Removes the infrastructure defined in the manifest from the cloud provider.

```bash
planton tofu destroy --manifest database.yaml
planton tofu destroy --manifest database.yaml --auto-approve
```

| Flag | Default | Description |
|------|---------|-------------|
| `--auto-approve` | `false` | Skip interactive approval before destroying |
| `--module-version` | | Check out a specific version of the IaC modules |
| `--no-cleanup` | `false` | Keep workspace copy after execution |

### refresh

Sync the OpenTofu state with actual cloud state without modifying any resources.

```bash
planton tofu refresh --manifest database.yaml
```

Queries the cloud provider for the current state of all managed resources and updates the local state file to match. No resources are created, modified, or deleted.

| Flag | Default | Description |
|------|---------|-------------|
| `--module-version` | | Check out a specific version of the IaC modules |
| `--no-cleanup` | `false` | Keep workspace copy after execution |

### generate-variables

Generate Terraform `variables.tf` content for a specified deployment component kind. This is useful when building custom HCL modules that need to accept Planton manifests as input.

```bash
planton tofu generate-variables KubernetesPostgres
planton tofu generate-variables AwsS3Bucket --output-file variables.tf
```

Takes exactly one argument: the deployment component kind name (e.g., `KubernetesPostgres`, `AwsS3Bucket`, `GcpCloudSqlPostgres`). The kind name matches the `CloudResourceKind` enum values from the protobuf definitions.

| Flag | Default | Description |
|------|---------|-------------|
| `--output-file` | stdout | File path to write the generated variables |

### load-tfvars

Load an Planton manifest and convert it to tfvars format. This is useful for integrating Planton manifests with standard Terraform/OpenTofu workflows.

```bash
planton tofu load-tfvars manifest.yaml
```

Takes exactly one argument: the path to the manifest file. Outputs the manifest content in tfvars format to stdout.

## Flags

All `planton tofu` subcommands inherit persistent flags from the parent command. Unlike [unified commands](./unified-commands), direct OpenTofu commands register `--manifest` without the `-f` shorthand.

### Parent Persistent Flags (All Subcommands)

| Flag | Short | Description |
|------|-------|-------------|
| `--manifest` | | Path to the deployment-component manifest file |
| `--input-dir` | | Directory containing `target.yaml` and credential YAML files |
| `--kustomize-dir` | | Directory containing kustomize configuration |
| `--overlay` | | Kustomize overlay (e.g., `prod`, `dev`, `staging`) |
| `--module-dir` | | Directory containing the OpenTofu module (default: current directory) |
| `--set` | | Override manifest values using `key=value` pairs |
| `--provider-config` | `-p` | Path to provider credentials file |

Direct OpenTofu commands do not support `--clipboard`, `--stack-input`, or the `-f` shorthand for `--manifest`. Use [unified commands](./unified-commands) for those input methods.

## Typical Workflow

```bash
# 1. Initialize backend and download providers
planton tofu init --manifest database.yaml --backend-type s3 \
  --backend-config bucket=my-state \
  --backend-config key=database/terraform.tfstate \
  --backend-config region=us-east-1

# 2. Preview what will be created
planton tofu plan --manifest database.yaml

# 3. Deploy
planton tofu apply --manifest database.yaml --auto-approve

# 4. After manifest changes, preview and apply
planton tofu plan --manifest database.yaml
planton tofu apply --manifest database.yaml --auto-approve

# 5. When done, tear down resources
planton tofu destroy --manifest database.yaml --auto-approve
```

## What's Next

- **[CLI Reference](./cli-reference)** — Complete flag reference
- **[Unified Commands](./unified-commands)** — Provisioner-agnostic alternative to direct OpenTofu commands
- **[Terraform Commands](./terraform-commands)** — Terraform engine commands (shared execution engine)
- **[Module Management](./module-management)** — Module versioning and the staging area
