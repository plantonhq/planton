---
title: "CLI Reference"
description: "Complete command-line reference for openmcf CLI - all commands, flags, and options"
icon: "terminal"
order: 1
---

# CLI Reference

Complete command-line reference for the `openmcf` CLI.

---

## Command Tree

```
openmcf
├── apply               Deploy infrastructure (unified, auto-detects provisioner)
├── destroy             Teardown infrastructure (or 'delete')
├── init                Initialize backend/stack (unified, auto-detects provisioner)
├── plan                Preview changes (or 'preview', unified, auto-detects provisioner)
├── refresh             Sync state with reality (unified, auto-detects provisioner)
├── pulumi              Manage infrastructure with Pulumi
│   ├── init           Initialize Pulumi stack
│   ├── preview        Preview infrastructure changes
│   ├── up             Deploy infrastructure (or 'update')
│   ├── refresh        Sync state with cloud reality
│   ├── destroy        Teardown infrastructure
│   ├── delete         Remove stack metadata (or 'rm')
│   └── cancel         Cancel ongoing Pulumi operation
├── tofu                Manage infrastructure with OpenTofu
│   ├── init           Initialize backend and providers
│   ├── plan           Preview infrastructure changes
│   ├── apply          Deploy infrastructure
│   ├── refresh        Sync state with cloud reality
│   ├── destroy        Teardown infrastructure
│   ├── generate-variables  Generate variables file from manifest
│   └── load-tfvars    Load tfvars from manifest
├── terraform           Manage infrastructure with Terraform
│   ├── init           Initialize backend and providers
│   ├── plan           Preview infrastructure changes
│   ├── apply          Deploy infrastructure
│   ├── refresh        Sync state with cloud reality
│   └── destroy        Teardown infrastructure
├── validate            Validate manifest against schema (or 'validate-manifest')
├── load-manifest       Load and display manifest with defaults
└── version             Show CLI version
```

---

## Top-Level Commands

### apply

**NEW!** Unified kubectl-style command to deploy infrastructure by automatically detecting the provisioner from the manifest label `openmcf.org/provisioner`.

**Usage**:

```bash
openmcf apply -f <file> [flags]
# or
openmcf apply -f <file> [flags]
```

**Example**:

```bash
# Auto-detect provisioner from manifest
openmcf apply -f database.yaml

# With kustomize
openmcf apply --kustomize-dir services/api --overlay prod

# With overrides
openmcf apply -f api.yaml --set spec.replicas=5
```

**How it works**:
1. Reads the `openmcf.org/provisioner` label from your manifest
2. Automatically routes to the appropriate provisioner (pulumi/tofu/terraform)
3. If label is missing, prompts you to select a provisioner interactively (defaults to Pulumi)

**Supported provisioners**: `pulumi`, `tofu`, `terraform` (case-insensitive)

### destroy

**NEW!** Unified kubectl-style command to destroy infrastructure. Works exactly like `apply` but tears down resources instead.

**Aliases**: `delete` (for kubectl compatibility)

**Usage**:

```bash
openmcf destroy -f <file> [flags]
openmcf delete -f <file> [flags]
```

**Example**:

```bash
# Auto-detect provisioner from manifest
openmcf destroy -f database.yaml

# Using kubectl-style delete alias
openmcf delete -f database.yaml

# With auto-approve (skips confirmation)
openmcf destroy -f api.yaml --auto-approve
```

### init

**NEW!** Unified command to initialize infrastructure backend or stack by automatically detecting the provisioner.

**Usage**:

```bash
openmcf init -f <file> [flags]
```

**Example**:

```bash
# Auto-detect provisioner from manifest
openmcf init -f database.yaml

# With kustomize
openmcf init --kustomize-dir services/api --overlay prod

# Reconfigure after backend changes
openmcf init -f app.yaml --reconfigure

# With tofu-specific backend config
openmcf init -f app.yaml --backend-type s3 --backend-config bucket=my-bucket
```

**How it works**:
1. Reads the `openmcf.org/provisioner` label from your manifest
2. Routes to appropriate initialization:
   - **Pulumi**: Creates stack if it doesn't exist
   - **Tofu**: Initializes backend and downloads providers
   - **Terraform**: Initializes backend and downloads providers

### plan

**NEW!** Unified command to preview infrastructure changes without applying them.

**Aliases**: `preview` (for Pulumi-style experience)

**Usage**:

```bash
openmcf plan -f <file> [flags]
openmcf preview -f <file> [flags]
```

**Example**:

```bash
# Auto-detect provisioner and preview changes
openmcf plan -f database.yaml

# Using preview alias (Pulumi-style)
openmcf preview -f database.yaml

# With kustomize
openmcf plan --kustomize-dir services/api --overlay staging

# Preview destroy plan (Tofu)
openmcf plan -f app.yaml --destroy
```

**How it works**:
1. Reads the `openmcf.org/provisioner` label from your manifest
2. Routes to appropriate preview operation:
   - **Pulumi**: Runs `pulumi preview`
   - **Tofu**: Runs `tofu plan`
   - **Terraform**: Runs `terraform plan`

### refresh

**NEW!** Unified command to sync state with cloud reality without modifying resources.

**Usage**:

```bash
openmcf refresh -f <file> [flags]
```

**Example**:

```bash
# Auto-detect provisioner and refresh state
openmcf refresh -f database.yaml

# With kustomize
openmcf refresh --kustomize-dir services/api --overlay prod

# Show detailed diffs (Pulumi)
openmcf refresh -f app.yaml --diff
```

**How it works**:
1. Queries your cloud provider for current resource state
2. Updates state file to reflect reality
3. Does NOT modify any cloud resources (read-only operation)
4. Routes based on provisioner:
   - **Pulumi**: Runs `pulumi refresh`
   - **Tofu**: Runs `tofu refresh`
   - **Terraform**: Runs `terraform refresh`

### pulumi

Manage infrastructure using Pulumi as the IaC engine.

**Subcommands**: `init`, `preview`, `up`/`update`, `refresh`, `destroy`, `delete`/`rm`, `cancel`

**Documentation**: See [Pulumi Commands Reference](/docs/cli/pulumi-commands)

**Example**:

```bash
openmcf pulumi up -f database.yaml
```

### tofu

Manage infrastructure using OpenTofu/Terraform as the IaC engine.

**Subcommands**: `init`, `plan`, `apply`, `refresh`, `destroy`

**Documentation**: See [OpenTofu Commands Reference](/docs/cli/tofu-commands)

**Example**:

```bash
openmcf tofu apply -f database.yaml
```

### validate

Validate a manifest against its Protocol Buffer schema without deploying.

**Usage**:

```bash
openmcf validate -f <file> [flags]
```

**Example**:

```bash
# Validate single manifest
openmcf validate -f ops/resources/database.yaml

# With kustomize
openmcf validate \
  --kustomize-dir services/api/kustomize \
  --overlay prod

# If valid: exits with code 0, no output
# If invalid: shows detailed errors, exits with code 1
```

**Flags**:
- `-f <file>`: Path to manifest file
- `--kustomize-dir <dir>`: Kustomize base directory
- `--overlay <name>`: Kustomize overlay name

### load-manifest

Load a manifest and display it with defaults applied and overrides resolved.

**Usage**:

```bash
openmcf load-manifest -f <file> [flags]
```

**Example**:

```bash
# Load manifest and see defaults
openmcf load-manifest -f database.yaml

# Load with overrides
openmcf load-manifest \
  -f api.yaml \
  --set spec.replicas=5

# Load kustomize-built manifest
openmcf load-manifest \
  --kustomize-dir services/api/kustomize \
  --overlay prod
```

**Flags**: Same as `validate`

**Output**: YAML manifest with defaults filled in and overrides applied

### version

Show OpenMCF CLI version information.

**Usage**:

```bash
openmcf version
```

**Example Output**:

```
openmcf version: v0.1.0
git commit: a1b2c3d
built: 2025-11-11T10:30:00Z
```

---

## Common Flags

These flags are available across multiple commands:

### Manifest Input

**`-f, -f <path>`**  
Path to manifest YAML file (local or URL). The `-f` shorthand is available for kubectl-style experience.

```bash
# Local file (kubectl-style)
-f ops/resources/database.yaml

# Local file (traditional)
-f ops/resources/database.yaml

# URL
-f https://raw.githubusercontent.com/myorg/manifests/main/db.yaml
```

**`--kustomize-dir <directory>`**  
Base directory containing kustomize structure.

```bash
--kustomize-dir services/api/kustomize
```

**`--overlay <name>`**  
Kustomize overlay environment to build (must be used with `--kustomize-dir`).

```bash
--overlay prod
```

**Priority**: `-f` > `--kustomize-dir` + `--overlay`

### Execution Control

**`--module-dir <path>`**  
Override IaC module directory (defaults to current directory).

```bash
--module-dir ~/projects/custom-modules/my-module
```

**`--set <key>=<value>`**  
Override manifest field values at runtime (repeatable).

```bash
--set spec.replicas=5 \
--set spec.container.image.tag=v2.0.0
```

### Pulumi-Specific Flags

**`--stack <org>/<project>/<stack>`**  
Override stack FQDN (instead of using manifest label).

```bash
--stack my-org/my-project/dev-stack
```

**`--yes`**  
Auto-approve operations without confirmation (Pulumi commands).

```bash
--yes
```

**`--force`**  
Force stack removal even if resources exist (`delete`/`rm` only).

```bash
--force
```

### OpenTofu/Terraform-Specific Flags

**`--auto-approve`**  
Skip interactive approval (`apply` and `destroy` commands).

```bash
--auto-approve
```

**`--reconfigure`**  
Reconfigure backend, ignoring any saved configuration. Use when backend configuration changes.

```bash
--reconfigure
```

**`--destroy`**  
Create destroy plan (`plan` command).

```bash
--destroy
```

### Provider Credentials

**Default Behavior (Environment Variables)**:

By default, the CLI reads credentials from environment variables - the same ones used by cloud provider CLIs. If you have `aws`, `gcloud`, or `az` configured, credentials are automatically available.

```bash
# These work without any credential flags if env vars are set
openmcf apply -f aws-vpc.yaml         # Uses AWS_ACCESS_KEY_ID, etc.
openmcf apply -f gcp-cluster.yaml     # Uses GOOGLE_APPLICATION_CREDENTIALS
openmcf apply -f azure-aks.yaml       # Uses ARM_CLIENT_ID, etc.
```

**Explicit Override (`-p, --provider-config <file>`)**:

Use the `-p` flag to override environment variables with an explicit credentials file:

```bash
-p ~/.config/aws-creds.yaml      # Override AWS credentials
-p ~/.config/gcp-creds.yaml      # Override GCP credentials
-p ~/.config/azure-creds.yaml    # Override Azure credentials
-p ~/.kube/config                # Override Kubernetes config
```

**How it works**: 
1. The CLI parses your manifest's `apiVersion` (e.g., `aws.openmcf.org/v1`)
2. Determines the required provider (e.g., AWS)
3. Loads credentials from environment variables OR from the file specified with `-p`

See the [Credentials Guide](/docs/guides/credentials) for the complete list of environment variables per provider.

---

## Environment Variables

### Respected by CLI

**`TF_LOG`** / **`PULUMI_LOG_LEVEL`**  
Enable verbose logging for debugging.

```bash
export TF_LOG=DEBUG
export PULUMI_LOG_LEVEL=3
```

### Provider Credentials

See [Credentials Guide](/docs/guides/credentials) for complete list of provider-specific environment variables.

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (validation failed, deployment failed, etc.) |

---

## Configuration Files

### Manifest Download Directory

Downloaded URL manifests are cached in:

```
~/.openmcf/manifests/downloaded/
```

### Module Cache

Cloned IaC modules are cached in:

```
~/.openmcf/modules/
```

---

## Examples by Use Case

### First Deployment

```bash
# Unified kubectl-style (recommended)
openmcf validate -f database.yaml
openmcf apply -f database.yaml

# Using Pulumi directly
openmcf validate -f database.yaml
openmcf pulumi up -f database.yaml

# Using OpenTofu directly
openmcf validate -f database.yaml
openmcf tofu init -f database.yaml
openmcf tofu plan -f database.yaml
openmcf tofu apply -f database.yaml
```

### Multi-Environment Deployment

```bash
# Deploy across environments with unified command
for env in dev staging prod; do
    openmcf apply \
        --kustomize-dir services/api/kustomize \
        --overlay $env \
        --yes
done
```

### CI/CD Deployment

```bash
# Non-interactive with dynamic values (unified command)
openmcf apply \
  -f deployment.yaml \
  --set spec.container.image.tag=$CI_COMMIT_SHA \
  --yes
```

### Testing Local Module Changes

```bash
# Point to local module during development
openmcf pulumi preview \
  -f test.yaml \
  --module-dir ~/dev/my-module
```

---

## Related Documentation

- [Pulumi Commands](/docs/cli/pulumi-commands) - Detailed Pulumi command guide
- [OpenTofu Commands](/docs/cli/tofu-commands) - Detailed OpenTofu command guide
- [Manifest Structure](/docs/guides/manifests) - Understanding manifests
- [Credentials Guide](/docs/guides/credentials) - Setting up cloud credentials
- [Advanced Usage](/docs/guides/advanced-usage) - Power user techniques

---

## Getting Help

**Command help**:

```bash
# General help
openmcf --help

# Command-specific help
openmcf pulumi --help
openmcf tofu apply --help
```

**Found an issue?** [Open an issue](https://github.com/plantonhq/openmcf/issues)

**Need support?** Check existing issues or discussions

