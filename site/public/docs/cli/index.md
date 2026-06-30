---
title: "CLI"
description: "Command-line interface documentation for the planton CLI — installation, command reference, and engine-specific guides"
icon: "code"
order: 20
---

# CLI

The `planton` CLI is a single binary that handles the full deployment lifecycle: manifest loading, validation, module resolution, provisioner execution, and state management across Pulumi, OpenTofu, and Terraform.

## Installation

```bash
# macOS (Homebrew)
brew install plantonhq/tap/planton

# Verify
planton version
```

For other platforms, download the binary from [GitHub Releases](https://github.com/plantonhq/planton/releases).

You also need at least one IaC engine installed:

```bash
# Pulumi
brew install pulumi

# OpenTofu
brew install opentofu

# Terraform
brew install terraform
```

## Shell Completion

Enable tab-completion for all commands, subcommands, and flags. This is a one-time setup per machine.

**Zsh (default on macOS):**

Add this line to your `~/.zshrc`:

```bash
source <(planton completion zsh)
```

**Bash:**

Add this line to your `~/.bashrc`:

```bash
source <(planton completion bash)
```

> Bash completion requires the `bash-completion` package. On macOS: `brew install bash-completion@2`.

**Fish:**

```bash
planton completion fish > ~/.config/fish/completions/planton.fish
```

**PowerShell:**

Add this line to your PowerShell profile:

```powershell
planton completion powershell | Out-String | Invoke-Expression
```

After setup, open a new terminal and type `planton <Tab>` to see available commands.

## How the CLI Works

Every deployment follows the same sequence regardless of which engine you choose:

1. **Load** a manifest from a file, clipboard, kustomize build, or stack input
2. **Validate** the manifest against its Protocol Buffer schema
3. **Resolve** the IaC module for the component kind
4. **Execute** the operation through the selected provisioner

You can let the CLI detect the provisioner automatically from the manifest's `planton.dev/provisioner` label using [unified commands](./unified-commands), or choose explicitly with `planton pulumi`, `planton tofu`, or `planton terraform`.

## In This Section

- **[CLI Reference](./cli-reference)** — Complete command tree, all flags organized by group, exit codes, and file system paths. The single source of truth for flag names and behavior.

- **[Unified Commands](./unified-commands)** — Provisioner-agnostic commands (`apply`, `plan`, `init`, `destroy`, `refresh`) that auto-detect the IaC engine from your manifest.

- **[Pulumi Commands](./pulumi-commands)** — Pulumi-specific subcommands: `init`, `preview`, `update`/`up`, `destroy`, `delete`/`rm`, `cancel`, `refresh`.

- **[OpenTofu Commands](./tofu-commands)** — OpenTofu-specific subcommands: `init`, `plan`, `apply`, `destroy`, `refresh`, `generate-variables`, `load-tfvars`.

- **[Terraform Commands](./terraform-commands)** — Terraform-specific subcommands: `init`, `plan`, `apply`, `destroy`, `refresh`. Shares the same HCL modules and execution engine as OpenTofu.

- **[Module Management](./module-management)** — Module resolution chain, the staging area, version pinning with `checkout` and `pull`, and CLI version management with `upgrade` and `downgrade`.

- **[Configuration & Utilities](./configuration)** — CLI configuration (`config set/get/list`), manifest validation (`validate`), manifest loading (`load`), and version checking.

## Quick Start

```bash
# Validate a manifest
planton validate -f database.yaml

# Deploy with automatic provisioner detection
planton apply -f database.yaml

# Preview changes before applying
planton plan -f database.yaml

# Tear down
planton destroy -f database.yaml
```

Or use a specific engine directly:

```bash
# Pulumi
planton pulumi up --manifest database.yaml --yes

# OpenTofu
planton tofu init --manifest database.yaml
planton tofu apply --manifest database.yaml --auto-approve

# Terraform
planton terraform init --manifest database.yaml
planton terraform apply --manifest database.yaml --auto-approve
```

## Getting Help

```bash
planton --help
planton apply --help
planton pulumi --help
planton tofu init --help
```

Every command and subcommand supports `--help`.
