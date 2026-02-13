---
title: "Unified Commands"
description: "Provisioner-agnostic commands that auto-detect Pulumi, OpenTofu, or Terraform from your manifest"
icon: "rocket"
order: 30
---

# Unified Commands

Unified commands let you deploy, preview, initialize, destroy, and refresh infrastructure without specifying which IaC engine to use. The CLI reads the provisioner from your manifest and routes to the correct engine automatically.

## How Provisioner Detection Works

Every OpenMCF manifest can include a label that declares which provisioner to use:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: my-database
  labels:
    openmcf.org/provisioner: tofu    # pulumi, tofu, or terraform
spec:
  # ...
```

When you run a unified command like `openmcf apply -f manifest.yaml`, the CLI:

1. Loads and validates the manifest
2. Reads the `openmcf.org/provisioner` label
3. Routes to the matching engine: Pulumi, OpenTofu, or Terraform
4. Passes all relevant flags through to the engine-specific handler

If the label is missing, the CLI prompts you to select a provisioner interactively.

The label value is case-insensitive. Valid values: `pulumi`, `tofu`, `terraform`.

## Commands

### apply

Deploy infrastructure. Routes to `pulumi update`, `tofu apply`, or `terraform apply` based on the manifest label.

```bash
openmcf apply -f database.yaml
openmcf apply --clipboard
openmcf apply --kustomize-dir services/api --overlay prod
openmcf apply -f api.yaml --set spec.container.replicas=3
```

When routed to Pulumi, this runs a `pulumi update` operation. When routed to OpenTofu or Terraform, this runs the equivalent `apply` operation.

### plan

Preview infrastructure changes without applying them. Routes to `pulumi preview`, `tofu plan`, or `terraform plan`.

```bash
openmcf plan -f database.yaml
openmcf preview -f database.yaml        # alias
openmcf plan -f app.yaml --destroy       # preview a destroy (Tofu/Terraform)
openmcf plan -f app.yaml --diff          # detailed diffs (Pulumi)
```

**Alias**: `preview` (for Pulumi-style workflows).

The `--destroy` flag creates a destroy plan instead of an apply plan. This flag only takes effect when routed to OpenTofu or Terraform.

### init

Initialize the backend or stack. Routes to `pulumi stack init`, `tofu init`, or `terraform init`.

```bash
openmcf init -f database.yaml
openmcf init -f app.yaml --backend-type s3 --backend-bucket my-state
openmcf init -f app.yaml --reconfigure
```

What initialization does depends on the provisioner:

| Provisioner | What `init` does |
|-------------|-----------------|
| Pulumi | Creates the stack if it does not exist |
| OpenTofu | Initializes backend configuration and downloads providers |
| Terraform | Initializes backend configuration and downloads providers |

For OpenTofu and Terraform, the backend type defaults to `local`. To use remote state, pass `--backend-type` and related flags. See the [init flags in the CLI Reference](./cli-reference#init-backend-configuration) for all backend configuration options.

### destroy

Tear down infrastructure. Routes to `pulumi destroy`, `tofu destroy`, or `terraform destroy`.

```bash
openmcf destroy -f database.yaml
openmcf delete -f database.yaml          # alias
openmcf destroy -f app.yaml --auto-approve
openmcf destroy -f app.yaml --yes         # auto-approve for Pulumi
```

**Alias**: `delete` (for kubectl-style workflows).

The `--auto-approve` flag skips interactive confirmation for OpenTofu and Terraform. For Pulumi, use `--yes` instead.

### refresh

Sync state with cloud reality without modifying any resources. Routes to `pulumi refresh`, `tofu refresh`, or `terraform refresh`.

```bash
openmcf refresh -f database.yaml
openmcf refresh -f app.yaml --diff       # detailed diffs (Pulumi)
```

This queries your cloud provider for the current state of managed resources and updates the state file to match. No cloud resources are created, modified, or deleted.

## Flag Groups

Unified commands accept flags from multiple groups. Which flags take effect depends on the provisioner the command routes to.

| Flag Group | Applies When | Reference |
|------------|-------------|-----------|
| Manifest source (`-f`, `-c`, `-i`, `--kustomize-dir`) | Always | [Manifest Source Flags](./cli-reference#manifest-source-flags) |
| Execution (`--set`, `--module-dir`, `--module-version`) | Always | [Execution Flags](./cli-reference#execution-flags) |
| Provider config (`-p`) | Always | [Provider Config Flags](./cli-reference#provider-config-flags) |
| Pulumi (`--stack`, `--yes`, `--diff`) | Pulumi only | [Pulumi Flags](./cli-reference#pulumi-specific-flags) |
| Tofu/Terraform apply (`--auto-approve`) | Tofu/Terraform only | [Apply/Destroy Flags](./cli-reference#apply-and-destroy) |
| Tofu/Terraform plan (`--destroy`) | Tofu/Terraform only | [Plan Flags](./cli-reference#plan) |
| Tofu/Terraform init (`--backend-type`, `--backend-bucket`, etc.) | Tofu/Terraform only | [Init Flags](./cli-reference#init-backend-configuration) |

Flags for an engine that is not selected are silently ignored. For example, `--yes` has no effect when the manifest routes to OpenTofu.

## Unified vs. Direct Commands

Unified commands and direct engine commands (`openmcf pulumi`, `openmcf tofu`, `openmcf terraform`) achieve the same result. The difference is in how the provisioner is selected and which flags are available.

| Aspect | Unified Commands | Direct Engine Commands |
|--------|-----------------|----------------------|
| Provisioner selection | Auto-detected from manifest label | Explicitly chosen by command |
| Manifest sources | All: `-f`, `-c`, `-i`, `--input-dir`, `--kustomize-dir` | `-f` (file path), `--input-dir`, `--kustomize-dir` only |
| Backend init flags | Full set (`--backend-bucket`, `--backend-key`, etc.) | Subset (`--backend-type`, `--backend-config` only) |
| Interactive prompt | Prompts if label missing | No prompt needed |
| Engine-specific subcommands | Not available (`cancel`, `delete`, `generate-variables`, etc.) | Available |

**Use unified commands when**:
- Your manifests include the `openmcf.org/provisioner` label
- You want a single workflow regardless of engine
- You need clipboard or stack-input manifest sources
- You need the full set of backend init flags

**Use direct engine commands when**:
- You need engine-specific subcommands like `pulumi cancel`, `pulumi delete`, `tofu generate-variables`, or `tofu load-tfvars`
- You want explicit control over which engine runs
- You are working with a specific engine's tooling conventions

## Execution Flow

Every unified command follows the same sequence:

1. **Load manifest** from the specified source (file, clipboard, stack-input, input-dir, or kustomize)
2. **Apply overrides** if `--set` flags are provided
3. **Validate** the manifest against its protobuf schema
4. **Detect provisioner** from the `openmcf.org/provisioner` label (or prompt)
5. **Resolve module** using the module resolution chain (see [Module Management](./module-management))
6. **Resolve credentials** from environment variables or `--provider-config` file
7. **Hand off** to the engine-specific handler (Pulumi, OpenTofu, or Terraform)

## What's Next

- **[CLI Reference](./cli-reference)** — Complete flag reference for all commands
- **[Pulumi Commands](./pulumi-commands)** — Pulumi-specific subcommands and workflows
- **[OpenTofu Commands](./tofu-commands)** — OpenTofu-specific subcommands and workflows
- **[Terraform Commands](./terraform-commands)** — Terraform-specific subcommands and workflows
