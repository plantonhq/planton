---
title: "Terraform Commands"
description: "Deploy, preview, and manage infrastructure using Terraform as the IaC engine"
icon: "code"
order: 60
---

# Terraform Commands

The `planton terraform` command group runs infrastructure operations using Terraform as the IaC engine. It shares the same HCL modules and execution engine as the [OpenTofu commands](./tofu-commands) but invokes the `terraform` binary instead of `tofu`.

## Prerequisites

- **Terraform binary**: The `terraform` binary must be installed and available on your `PATH`. Install from [developer.hashicorp.com/terraform](https://developer.hashicorp.com/terraform/install).
- **Module resolution**: The CLI resolves HCL modules through the same resolution chain as OpenTofu: direct path, GitHub release archive, or staging area. See [Module Management](./module-management) for details.

## Relationship to OpenTofu

Planton's Terraform and OpenTofu command groups share the same:

- **HCL modules**: Every deployment component has a single `iac/tf/` directory used by both engines
- **Module resolution**: Both use the same download, caching, and staging mechanisms
- **Execution engine**: Both route through the same Go package (`tofumodule.RunCommand`) with the binary name as the only difference

The practical difference is the binary invoked: `terraform` vs. `tofu`. Choose based on your organization's licensing and toolchain preferences.

## Subcommands

### init

Initialize the backend configuration and download required providers.

```bash
planton terraform init --manifest database.yaml
planton terraform init --manifest database.yaml --backend-type s3 --backend-config bucket=my-state
```

| Flag | Default | Description |
|------|---------|-------------|
| `--backend-type` | `local` | Backend type: `local`, `s3`, `gcs`, `azurerm`, etc. |
| `--backend-config` | | Backend configuration `key=value` pairs (repeatable) |
| `--module-version` | | Check out a specific version of the IaC modules |

### plan

Preview infrastructure changes without applying them.

```bash
planton terraform plan --manifest database.yaml
planton terraform plan --manifest database.yaml --destroy
```

| Flag | Default | Description |
|------|---------|-------------|
| `--destroy` | `false` | Create a destroy plan instead of an apply plan |
| `--module-version` | | Check out a specific version of the IaC modules |
| `--no-cleanup` | `false` | Keep workspace copy after execution |

### apply

Deploy infrastructure by applying the planned changes.

```bash
planton terraform apply --manifest database.yaml
planton terraform apply --manifest database.yaml --auto-approve
planton terraform apply --manifest api.yaml --set spec.container.replicas=5
```

| Flag | Default | Description |
|------|---------|-------------|
| `--auto-approve` | `false` | Skip interactive approval before applying |
| `--module-version` | | Check out a specific version of the IaC modules |
| `--no-cleanup` | `false` | Keep workspace copy after execution |

### destroy

Tear down all resources managed by the current state.

```bash
planton terraform destroy --manifest database.yaml
planton terraform destroy --manifest database.yaml --auto-approve
```

| Flag | Default | Description |
|------|---------|-------------|
| `--auto-approve` | `false` | Skip interactive approval before destroying |
| `--module-version` | | Check out a specific version of the IaC modules |
| `--no-cleanup` | `false` | Keep workspace copy after execution |

### refresh

Sync the Terraform state with actual cloud state without modifying any resources.

```bash
planton terraform refresh --manifest database.yaml
```

| Flag | Default | Description |
|------|---------|-------------|
| `--module-version` | | Check out a specific version of the IaC modules |
| `--no-cleanup` | `false` | Keep workspace copy after execution |

## Flags

All `planton terraform` subcommands inherit persistent flags from the parent command. Like OpenTofu direct commands, `--manifest` does not have the `-f` shorthand.

### Parent Persistent Flags (All Subcommands)

| Flag | Short | Description |
|------|-------|-------------|
| `--manifest` | | Path to the deployment-component manifest file |
| `--input-dir` | | Directory containing `target.yaml` and credential YAML files |
| `--kustomize-dir` | | Directory containing kustomize configuration |
| `--overlay` | | Kustomize overlay (e.g., `prod`, `dev`, `staging`) |
| `--module-dir` | | Directory containing the Terraform module (default: current directory) |
| `--set` | | Override manifest values using `key=value` pairs |
| `--provider-config` | `-p` | Path to provider credentials file |

Direct Terraform commands do not support `--clipboard`, `--stack-input`, or the `-f` shorthand for `--manifest`. Use [unified commands](./unified-commands) for those input methods.

## Differences from OpenTofu Commands

| Feature | Terraform | OpenTofu |
|---------|-----------|----------|
| Subcommands | 5 (init, plan, apply, destroy, refresh) | 7 (adds `generate-variables`, `load-tfvars`) |
| Binary | `terraform` | `tofu` |
| License | BSL 1.1 (HashiCorp) | MPL 2.0 (OpenTofu) |
| HCL modules | Same | Same |
| Module resolution | Same | Same |

The `generate-variables` and `load-tfvars` utility commands are only available under `planton tofu`. Their output is compatible with both Terraform and OpenTofu, so you can use `planton tofu generate-variables` even if you plan to run the generated files with Terraform.

## Typical Workflow

```bash
# 1. Initialize backend and download providers
planton terraform init --manifest database.yaml \
  --backend-type s3 \
  --backend-config bucket=my-state \
  --backend-config key=database/terraform.tfstate

# 2. Preview
planton terraform plan --manifest database.yaml

# 3. Deploy
planton terraform apply --manifest database.yaml --auto-approve

# 4. Tear down when done
planton terraform destroy --manifest database.yaml --auto-approve
```

## What's Next

- **[CLI Reference](./cli-reference)** — Complete flag reference
- **[Unified Commands](./unified-commands)** — Provisioner-agnostic alternative to direct Terraform commands
- **[OpenTofu Commands](./tofu-commands)** — OpenTofu engine commands (shared execution engine)
- **[Module Management](./module-management)** — Module versioning and the staging area
