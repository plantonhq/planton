---
title: "OpenTofu Commands Reference"
description: "Complete guide to managing infrastructure with openmcf tofu commands - init, plan, apply, refresh, and destroy"
icon: "terminal"
order: 3
---

# OpenTofu Commands Reference

Your complete guide to managing infrastructure with `openmcf tofu` commands.

---

## Overview

Think of OpenTofu as your infrastructure's blueprint compiler. Just as a compiler takes source code and produces an executable, OpenTofu takes your manifest and produces a plan for infrastructure changes. The `openmcf` CLI wraps OpenTofu operations with manifest-driven workflows, giving you the same consistent experience as Pulumi but using the battle-tested Terraform/OpenTofu engine.

### The Infrastructure Lifecycle

```
┌──────────┐    ┌─────────┐    ┌────────┐    ┌─────────┐    ┌─────────┐
│   init   │ -> │  plan   │ -> │ apply  │ -> │ refresh │ -> │ destroy │
└──────────┘    └─────────┘    └────────┘    └─────────┘    └─────────┘
     │               │              │              │              │
 Initialize       Preview        Deploy        Sync State     Teardown
  Backend        Changes       Resources      with Cloud     Resources
```

### Key Concepts

**Manifest**: A YAML file describing your infrastructure resource (e.g., `r2-bucket.yaml`, `eks-cluster.yaml`). Think of it as a recipe.

**State File**: Where OpenTofu stores information about managed resources. Unlike Pulumi's "stack" concept, OpenTofu uses state files stored in backends (S3, GCS, local, etc.).

**Module Directory**: Where the Terraform/OpenTofu IaC code lives. Usually auto-detected from your manifest's resource kind, but can be overridden with `--module-dir`.

**Backend**: Where OpenTofu stores state files. Configure this in your module's `backend.tf` file (supports S3, GCS, Azure Blob, local file system, and more).

**Plan File**: An optional output from the `plan` command that can be applied exactly as planned (not commonly used with openmcf).

### OpenTofu vs Pulumi: Key Differences

If you're familiar with the Pulumi commands, here's how OpenTofu differs:

| Concept | Pulumi | OpenTofu |
|---------|--------|----------|
| **State Container** | Stack (org/project/stack) | Workspace + State File |
| **Preview Changes** | `preview` command | `plan` command |
| **Deploy** | `up` or `update` | `apply` |
| **Auto-approve Flag** | `--yes` | `--auto-approve` |
| **Stack Deletion** | `delete` or `rm` | No equivalent (just delete state) |
| **Initialization** | Creates stack metadata | Initializes backend connection |

---

## Commands

### `init` - Initialize Backend

**What it does**: Initializes the OpenTofu backend, downloads required provider plugins, and prepares the module directory for operations. Think of this as "setting up your workspace."

**When to use**:
- First time working with a module
- After cleaning `.terraform` directory
- After changing backend configuration
- When provider requirements change

**Behavior**:
- Downloads and installs provider plugins based on module requirements
- Initializes the backend (S3, GCS, local, etc.) for state storage
- Creates `.terraform` directory with cached plugins and configuration
- Does NOT create or modify any cloud resources
- Idempotent - safe to run multiple times

**Usage**:

```bash
openmcf tofu init -f <manifest-file> [flags]
```

**Examples**:

```bash
# Initialize for a basic deployment
openmcf tofu init \
  -f ops/cloud-resources/prod/r2-bucket.yaml

# Initialize using kustomize overlay
openmcf tofu init \
  --kustomize-dir backend/services/api \
  --overlay prod

# Initialize with explicit module directory (for development/testing)
openmcf tofu init \
  -f ops/resources/vpc.yaml \
  --module-dir ~/projects/custom-modules/aws-vpc
```

**What you'll see**:

```
● Loading manifest...
✔ Manifest loaded
● Validating manifest...
✔ Manifest validated
● Preparing OpenTofu execution...
✔ Execution prepared

🤝 Handing off to OpenTofu...
   Output below is from OpenTofu

Initializing the backend...
Initializing provider plugins...
- Finding latest version of hashicorp/aws...
- Installing hashicorp/aws v5.30.0...
- Installed hashicorp/aws v5.30.0

OpenTofu has been successfully initialized!

You may now begin working with OpenTofu. Try running "tofu plan" to see
any changes that are required for your infrastructure.
```

**Important Notes**:

- **Run this first**: Unlike `pulumi up` which auto-creates stacks, OpenTofu requires explicit initialization
- **Cached plugins**: Downloaded plugins are cached in `.terraform/` - delete this to force re-download
- **Backend setup**: The first init for a new state file creates it in your backend
- **Module updates**: Run init again after updating module dependencies

---

### `plan` - Preview Infrastructure Changes

**What it does**: Creates an execution plan showing what OpenTofu will do when you apply your manifest. This is your "diff" before committing changes—you see exactly what will be created, modified, or destroyed.

**When to use**:
- Before running `apply` to understand what will change
- To validate your manifest produces the expected infrastructure changes
- During code review to demonstrate infrastructure changes
- When debugging unexpected behavior
- To create a "destroy plan" (using `--destroy` flag)

**Behavior**:
- Compares your manifest against the current infrastructure state
- Shows a detailed execution plan with additions, modifications, and deletions
- Does NOT modify any cloud resources
- Does NOT modify state file
- Optionally saves the plan to a file (not commonly done with openmcf)

**Usage**:

```bash
openmcf tofu plan -f <manifest-file> [flags]
```

**Examples**:

```bash
# Plan changes for a Kubernetes deployment
openmcf tofu plan \
  -f services/api/deployment.yaml

# Plan with field overrides (useful for testing different configurations)
openmcf tofu plan \
  -f services/api/deployment.yaml \
  --set spec.replicas=5 \
  --set spec.container.image.tag=v2.0.0

# Plan using kustomize (common for multi-environment setups)
openmcf tofu plan \
  --kustomize-dir backend/services/api \
  --overlay staging

# Create a destroy plan (preview what destroy will do)
openmcf tofu plan \
  -f ops/resources/test-cluster.yaml \
  --destroy
```

**Reading the output**:

```
● Loading manifest...
✔ Manifest loaded
● Validating manifest...
✔ Manifest validated
● Preparing OpenTofu execution...
✔ Execution prepared

🤝 Handing off to OpenTofu...
   Output below is from OpenTofu

OpenTofu will perform the following actions:

  # cloudflare_r2_bucket.bucket will be created
  + resource "cloudflare_r2_bucket" "bucket" {
      + account_id = "abc123"
      + id         = (known after apply)
      + location   = "WNAM"
      + name       = "my-bucket"
    }

Plan: 1 to add, 0 to change, 0 to destroy.
```

**Legend**:
- `+` = Resource will be created
- `~` = Resource will be modified (shows specific attribute changes)
- `-` = Resource will be deleted
- `-/+` = Resource will be replaced (destroy then create, often due to immutable property changes)
- `(known after apply)` = Value will be computed during apply

**Plan Modes**:

```bash
# Standard plan (create/update resources)
openmcf tofu plan -f resource.yaml

# Destroy plan (preview destruction)
openmcf tofu plan -f resource.yaml --destroy
```

---

### `apply` - Deploy Infrastructure

**What it does**: Applies your manifest to create, update, or configure cloud resources. This is the "execute" command—it actually makes the infrastructure changes.

**When to use**:
- After reviewing changes with `plan` and confirming they look correct
- Initial deployment of new infrastructure
- Updating existing infrastructure with new configurations
- Applying configuration changes (scaling, updates, feature flags)

**Behavior**:
- Shows a plan of changes (unless you use `--auto-approve`)
- Waits for your confirmation before proceeding (unless `--auto-approve` is provided)
- Creates/updates/deletes cloud resources to match your manifest
- Updates state file to reflect the new infrastructure state
- Attempts to roll back on failures (where provider supports it)

**Usage**:

```bash
openmcf tofu apply -f <manifest-file> [flags]
```

**Examples**:

```bash
# Interactive deployment (will show plan and ask for confirmation)
openmcf tofu apply \
  -f ops/resources/database.yaml

# Non-interactive deployment (CI/CD pipelines)
openmcf tofu apply \
  -f ops/resources/database.yaml \
  --auto-approve

# Deploy with field overrides
openmcf tofu apply \
  -f ops/resources/cache.yaml \
  --set spec.instanceSize=large \
  --set spec.replicas=3

# Deploy using kustomize overlay
openmcf tofu apply \
  --kustomize-dir backend/services/worker \
  --overlay prod \
  --auto-approve
```

**What you'll see**:

```
● Loading manifest...
✔ Manifest loaded
● Validating manifest...
✔ Manifest validated
● Preparing OpenTofu execution...
✔ Execution prepared

🤝 Handing off to OpenTofu...
   Output below is from OpenTofu

OpenTofu will perform the following actions:

  # gcp_sql_database_instance.main_db will be created
  + resource "gcp_sql_database_instance" "main_db" {
      + database_version = "POSTGRES_14"
      + name             = "main-db"
      + project          = "my-project"
      + region           = "us-central1"
      + tier             = "db-n1-standard-1"
    }

Plan: 1 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  OpenTofu will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

gcp_sql_database_instance.main_db: Creating...
gcp_sql_database_instance.main_db: Still creating... [10s elapsed]
gcp_sql_database_instance.main_db: Still creating... [20s elapsed]
gcp_sql_database_instance.main_db: Creation complete after 2m15s [id=main-db]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

**Important Notes**:

- **Plan-then-apply workflow**: By default, `apply` shows you a plan and waits for confirmation. This is your safety net.
- **No automatic initialization**: Unlike Pulumi's `up`, you must run `init` before first `apply`
- **State locking**: OpenTofu automatically locks state during apply to prevent concurrent modifications (if backend supports it)
- **Failed applies**: If apply fails midway, state will reflect the partial changes. You can re-run `apply` to continue or use `refresh` to sync state
- **Idempotent**: Running apply with no changes is safe—OpenTofu detects no changes needed

---

### `refresh` - Sync State with Reality

**What it does**: Updates OpenTofu's state file to match the actual resources in your cloud provider without modifying any resources. Think of this as "git fetch" for infrastructure—it synchronizes your state with reality.

**When to use**:
- After manual changes made outside OpenTofu (e.g., via cloud console, CLI, or other tools)
- Before running `plan` or `apply` to ensure state accuracy
- After failed deployments to resynchronize state
- When troubleshooting drift between desired and actual state
- After importing existing resources into OpenTofu management

**Behavior**:
- Queries your cloud provider for the current state of managed resources
- Updates state file to reflect actual resource properties
- Does NOT modify any cloud resources
- Does NOT change your manifest file
- Shows what changed in the state (if anything)
- Detects resources that were deleted outside OpenTofu

**Usage**:

```bash
openmcf tofu refresh -f <manifest-file> [flags]
```

**Examples**:

```bash
# Refresh to sync state after manual changes
openmcf tofu refresh \
  -f ops/resources/s3-bucket.yaml

# Refresh before important operations
openmcf tofu refresh \
  -f ops/resources/production-db.yaml && \
openmcf tofu apply \
  -f ops/resources/production-db.yaml
```

**What you'll see**:

```
● Loading manifest...
✔ Manifest loaded
● Validating manifest...
✔ Manifest validated
● Preparing OpenTofu execution...
✔ Execution prepared

🤝 Handing off to OpenTofu...
   Output below is from OpenTofu

aws_s3_bucket.assets: Refreshing state... [id=assets-prod-xyz123]

Note: Objects have changed outside of OpenTofu

OpenTofu detected the following changes made outside of OpenTofu since the
last "tofu apply":

  # aws_s3_bucket.assets has changed
  ~ resource "aws_s3_bucket" "assets" {
        id                          = "assets-prod-xyz123"
      ~ tags                        = {
          + "Environment" = "production"
        }
        # (10 unchanged attributes hidden)
    }

This is a data-only operation. No changes will be made to your
infrastructure.
```

**Understanding Drift**:

Drift occurs when someone (or something) modifies infrastructure outside of OpenTofu:

```
Before refresh:
  State File: bucket versioning = false

Actual Cloud:
  AWS Console: someone enabled versioning = true

After refresh:
  State File: bucket versioning = true (synced!)
```

**Next steps after refresh**:
1. If changes match your manifest → You're good, carry on
2. If unexpected changes → Investigate who/what made them
3. If you want to revert manual changes → Run `apply` to restore manifest configuration

**Important Note**: Starting with Terraform 0.15+ and OpenTofu, refresh is automatically run as part of `plan` and `apply`. Explicit refresh is mainly useful for viewing drift without planning changes.

---

### `destroy` - Teardown Infrastructure

**What it does**: Destroys all cloud resources managed by your manifest. This is the "rm -rf" of infrastructure—use with extreme caution.

**When to use**:
- Tearing down temporary environments (dev, testing, ephemeral previews)
- Decommissioning infrastructure that's no longer needed
- Cleaning up after testing or experimentation
- Cost optimization (shutting down unused resources)
- Before major refactoring (destroy old, deploy new)

**Behavior**:
- Shows a destroy plan (resources to be deleted)
- Waits for explicit confirmation (unless `--auto-approve` is provided)
- Deletes resources in reverse dependency order (children before parents)
- Updates state file to reflect deletion
- **Cannot be undone** once confirmed

**Usage**:

```bash
openmcf tofu destroy -f <manifest-file> [flags]
```

**Examples**:

```bash
# Interactive destroy (will ask for confirmation)
openmcf tofu destroy \
  -f ops/resources/dev-cluster.yaml

# Non-interactive destroy (automation/CI)
openmcf tofu destroy \
  -f ops/resources/test-environment.yaml \
  --auto-approve

# Destroy temporary environment
openmcf tofu destroy \
  --kustomize-dir backend/services/api \
  --overlay pr-123 \
  --auto-approve
```

**What you'll see**:

```
● Loading manifest...
✔ Manifest loaded
● Validating manifest...
✔ Manifest validated
● Preparing OpenTofu execution...
✔ Execution prepared

🤝 Handing off to OpenTofu...
   Output below is from OpenTofu

OpenTofu will perform the following actions:

  # gcp_container_cluster.test_cluster will be destroyed
  - resource "gcp_container_cluster" "test_cluster" {
      - name     = "test-cluster" -> null
      - location = "us-central1"  -> null
      - ...
    }

  # gcp_container_node_pool.default_pool will be destroyed
  - resource "gcp_container_node_pool" "default_pool" {
      - name = "default-pool" -> null
      - ...
    }

Plan: 0 to add, 0 to change, 2 to destroy.

Do you really want to destroy all resources?
  OpenTofu will destroy all your managed infrastructure, as shown above.
  There is no undo. Only 'yes' will be accepted to confirm.

  Enter a value: yes

gcp_container_node_pool.default_pool: Destroying... [id=default-pool]
gcp_container_node_pool.default_pool: Still destroying... [10s elapsed]
gcp_container_node_pool.default_pool: Destruction complete after 45s
gcp_container_cluster.test_cluster: Destroying... [id=test-cluster]
gcp_container_cluster.test_cluster: Still destroying... [10s elapsed]
gcp_container_cluster.test_cluster: Destruction complete after 2m30s

Destroy complete! Resources: 2 destroyed.
```

**⚠️ Safety Warnings**:

1. **Permanent deletion**: Most cloud providers permanently delete resources. Some have soft-delete/trash, but don't count on it.
2. **Data loss**: Databases, storage buckets, and other stateful resources will lose their data unless you have backups.
3. **Dependency risk**: If other resources depend on what you're destroying, they may break.
4. **No undo**: Once you confirm, there's no rollback. The resources are gone.
5. **State remains**: Unlike cloud resources, the state file remains. To clean it up, manually delete the state file from your backend.

**Best Practices**:

```bash
# ✅ Good: Review before destroying
openmcf tofu plan -f prod.yaml --destroy  # Preview destruction
openmcf tofu destroy -f prod.yaml          # Interactive confirmation

# ⚠️ Risky: Blind destruction
openmcf tofu destroy -f prod.yaml --auto-approve

# ✅ Good: Backup data first
aws s3 sync s3://my-bucket ./backup-$(date +%Y%m%d)/
openmcf tofu destroy -f s3-bucket.yaml

# ✅ Good: Verify manifest before destroying
cat prod.yaml  # Make absolutely sure this is the right file
openmcf tofu destroy -f prod.yaml
```

---

## Common Flags

All commands support these flags. They're like the universal remote for infrastructure management.

### Manifest Input

**`-f <file>`**: Path to your resource manifest YAML file.

```bash
openmcf tofu apply -f ops/resources/my-resource.yaml
```

**`--kustomize-dir <dir>`** + **`--overlay <name>`**: Use kustomize for environment-specific configurations.

```bash
# Loads kustomize base + overlays/prod
openmcf tofu apply \
  --kustomize-dir backend/services/api \
  --overlay prod
```

**Priority**: `-f` > `--kustomize-dir` + `--overlay`

### Execution Control

**`--module-dir <path>`**: Override the OpenTofu module directory (defaults to current directory).

```bash
# Use local development module instead of released version
openmcf tofu apply \
  -f my-resource.yaml \
  --module-dir ~/projects/custom-modules/my-module
```

**`--auto-approve`**: Auto-approve without confirmation prompts (for CI/CD). Available for `apply` and `destroy` commands.

```bash
openmcf tofu apply -f resource.yaml --auto-approve
```

**`--reconfigure`**: Reconfigure backend, ignoring any saved configuration. Use when backend configuration changes.

```bash
# Reconfigure after backend label changes
openmcf tofu init -f resource.yaml --reconfigure

# Also works with other commands
openmcf tofu apply -f resource.yaml --reconfigure
```

**`--destroy`**: Create a destruction plan (only for `plan` command).

```bash
# Preview what destroy will do
openmcf tofu plan -f resource.yaml --destroy
```

**`--set <key>=<value>`**: Override manifest values at runtime (repeatable flag).

```bash
openmcf tofu apply \
  -f deployment.yaml \
  --set spec.replicas=10 \
  --set spec.container.image.tag=v2.1.0 \
  --set metadata.env=staging
```

### Provider Credentials

**Default Behavior**: Credentials are automatically loaded from environment variables. If you have your cloud provider CLI configured (e.g., `aws configure`, `gcloud auth`, `az login`), no additional credential setup is needed.

```bash
# Works without -p flag if environment variables are set
openmcf tofu apply -f ops/aws-resources/vpc.yaml  # Uses AWS_ACCESS_KEY_ID, etc.
openmcf tofu apply -f ops/gcp-resources/gke.yaml  # Uses GOOGLE_APPLICATION_CREDENTIALS
```

**Explicit Override**: Use `-p` / `--provider-config` to override environment variables with a credentials file.

**`-p, --provider-config <file>`**: Path to provider credential YAML file

```bash
# Override with explicit credentials file
openmcf tofu apply \
  -f ops/aws-resources/vpc.yaml \
  -p ~/.config/openmcf/credentials/aws-prod.yaml

openmcf tofu apply \
  -f ops/gcp-resources/cluster.yaml \
  -p ~/.config/openmcf/credentials/gcp-prod.yaml
```

The CLI auto-detects which provider is needed from your manifest's `apiVersion`. See the [Credentials Guide](/docs/guides/credentials) for environment variable details.

---

## Common Workflows

### First-Time Deployment

```bash
# 1. Initialize the backend and providers
openmcf tofu init -f my-resource.yaml

# 2. Preview what will be created
openmcf tofu plan -f my-resource.yaml

# 3. Deploy the infrastructure
openmcf tofu apply -f my-resource.yaml
```

### Updating Existing Infrastructure

```bash
# 1. Edit your manifest
vim ops/resources/my-app.yaml

# 2. Preview the changes
openmcf tofu plan -f ops/resources/my-app.yaml

# 3. Apply if changes look good
openmcf tofu apply -f ops/resources/my-app.yaml
```

### Testing Configuration Changes

```bash
# Preview with overrides (no changes to manifest file)
openmcf tofu plan \
  -f api-deployment.yaml \
  --set spec.replicas=20 \
  --set spec.resources.limits.cpu=4000m

# If it looks good, apply with same overrides
openmcf tofu apply \
  -f api-deployment.yaml \
  --set spec.replicas=20 \
  --set spec.resources.limits.cpu=4000m

# Later, commit the changes to manifest
vim api-deployment.yaml  # Make changes permanent
```

### Emergency Rollback

```bash
# Scenario: v2.0.0 deployment has issues, need to roll back to v1.9.5

# Option 1: Override the current manifest
openmcf tofu apply \
  -f deployment.yaml \
  --set spec.container.image.tag=v1.9.5

# Option 2: Revert manifest to previous version
git checkout HEAD~1 deployment.yaml
openmcf tofu apply -f deployment.yaml

# Option 3: Use a previous Git revision
git show HEAD~5:deployment.yaml > /tmp/previous-deployment.yaml
openmcf tofu apply -f /tmp/previous-deployment.yaml
```

### Syncing After Manual Changes

```bash
# Someone made changes via AWS console, need to sync state

# 1. Refresh to see what changed
openmcf tofu refresh -f s3-bucket.yaml

# 2. Review the diff
openmcf tofu plan -f s3-bucket.yaml

# 3. Decide:
#    - Changes match manifest? → Do nothing, state is synced
#    - Changes don't match? → Update manifest or revert via `apply`

# 4. If reverting manual changes:
openmcf tofu apply -f s3-bucket.yaml  # Restores manifest config
```

### Multi-Environment Deployment

```bash
# Using kustomize overlays for different environments

# Deploy to dev
openmcf tofu apply \
  --kustomize-dir services/api \
  --overlay dev

# Preview staging changes
openmcf tofu plan \
  --kustomize-dir services/api \
  --overlay staging

# Deploy to production (with extra caution)
openmcf tofu plan \
  --kustomize-dir services/api \
  --overlay prod
# Review carefully...
openmcf tofu apply \
  --kustomize-dir services/api \
  --overlay prod
```

### Local Module Development

```bash
# Testing changes to OpenTofu module code without publishing

cd ~/projects/openmcf/apis/.../.../iac/tofu

# Initialize with local module
openmcf tofu init \
  -f ~/manifests/test-resource.yaml \
  --module-dir .

# Preview with local module
openmcf tofu plan \
  -f ~/manifests/test-resource.yaml \
  --module-dir .

# Iterate: edit module code, run plan again
vim main.tf
openmcf tofu plan \
  -f ~/manifests/test-resource.yaml \
  --module-dir .

# Deploy with local module
openmcf tofu apply \
  -f ~/manifests/test-resource.yaml \
  --module-dir .
```

### CI/CD Pipeline

```bash
#!/bin/bash
# deploy.sh - Automated deployment script

set -e  # Exit on error

MANIFEST="ops/resources/app-${ENV}.yaml"

echo "🔍 Planning changes..."
openmcf tofu plan -f "$MANIFEST"

echo "🚀 Deploying infrastructure..."
openmcf tofu apply -f "$MANIFEST" --auto-approve

echo "✅ Deployment complete"
```

**GitHub Actions Example**:

```yaml
name: Deploy Infrastructure

on:
  push:
    branches: [main]
    paths: ['ops/resources/**']

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup OpenTofu
        uses: opentofu/setup-opentofu@v1
        with:
          tofu_version: 1.6.0
      
      - name: Deploy Resources
        run: |
          openmcf tofu init \
            -f ops/resources/prod-infra.yaml
          
          openmcf tofu apply \
            -f ops/resources/prod-infra.yaml \
            --auto-approve
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
```

---

## Troubleshooting

### Error: "Backend initialization required"

**Symptom**: Commands fail saying backend needs initialization.

**Cause**: Haven't run `init` yet or `.terraform` directory was deleted.

**Solution**:

```bash
# Run init to set up backend and providers
openmcf tofu init -f my-resource.yaml

# Then try your command again
openmcf tofu plan -f my-resource.yaml
```

### Error: "state lock"

**Symptom**: Command fails saying state is locked by another process.

**Cause**: A previous operation crashed or is still running, leaving the state locked.

**Solution**:

```bash
# Option 1: Wait for the other operation to complete

# Option 2: Force unlock (use with caution - only if you're certain no operation is running)
cd <module-directory>
tofu force-unlock <lock-id>

# Then retry your operation
openmcf tofu apply -f my-resource.yaml
```

### Provider Authentication Failures

**Symptom**: "failed to configure provider" or authentication errors.

**Causes**: Missing or invalid cloud provider credentials.

**Solutions**:

**For AWS**:
```bash
# Check credentials
aws sts get-caller-identity

# Or provide provider config file
openmcf tofu apply \
  -f resource.yaml \
  -p ~/.aws/credentials-prod.yaml
```

**For GCP**:
```bash
# Check credentials
gcloud auth list
gcloud config get-value project

# Or set environment variable
export GOOGLE_APPLICATION_CREDENTIALS=~/gcp-key.json
openmcf tofu apply -f resource.yaml
```

**For Cloudflare**:
```bash
# Set API token
export CLOUDFLARE_API_TOKEN="your-token-here"
openmcf tofu apply -f resource.yaml
```

### Plan Shows Unexpected Changes

**Symptom**: `plan` shows modifications you didn't make.

**Causes**:
1. Someone made manual changes outside OpenTofu
2. Provider API defaults changed
3. Computed values changed upstream
4. State file is out of sync

**Solution**:

```bash
# First, sync state with reality
openmcf tofu refresh -f resource.yaml

# Then plan again
openmcf tofu plan -f resource.yaml

# If changes persist, check for:
# - Manual modifications in cloud console
# - Provider version changes
# - Upstream resource changes
```

### Resource Already Exists

**Symptom**: "resource already exists" error during apply.

**Cause**: Resources exist in cloud but not in state file (created outside OpenTofu or state lost).

**Solution**:

```bash
# Option 1: Import existing resources (advanced)
cd <module-directory>
tofu import <resource-type>.<resource-name> <cloud-resource-id>

# Option 2: Manually delete cloud resources
# (Use cloud provider console/CLI to delete conflicting resources)

# Option 3: Use different resource names in manifest
vim my-resource.yaml  # Change metadata.name or resource IDs
```

### Module Not Found

**Symptom**: "module not found" or "no configuration files" error.

**Cause**: OpenTofu can't find the module directory or it contains no `.tf` files.

**Solution**:

```bash
# Check module directory exists
ls -la <module-directory>

# Verify it contains .tf files
ls <module-directory>/*.tf

# If using custom module, ensure --module-dir points to correct location
openmcf tofu init \
  -f resource.yaml \
  --module-dir /correct/path/to/module
```

---

## Best Practices

### 1. **Always Plan Before Applying**

```bash
# ✅ Good: Review changes first
openmcf tofu plan -f resource.yaml
# Read output, verify changes look correct
openmcf tofu apply -f resource.yaml

# ⚠️ Risky: Blind deployment
openmcf tofu apply -f resource.yaml --auto-approve
```

**Why**: Plan is your safety net. It catches mistakes before they become expensive incidents.

### 2. **Use Version Control for Manifests**

```bash
# ✅ Good: Track changes in Git
git add ops/resources/my-resource.yaml
git commit -m "feat: increase database instance size"
git push
# Deploy via CI/CD or manually

# ❌ Bad: Direct edits without version control
vim /tmp/my-resource.yaml
openmcf tofu apply -f /tmp/my-resource.yaml
```

**Why**: Version control gives you change history, rollback capability, and code review.

### 3. **Initialize Before Each Session**

```bash
# ✅ Good: Run init when starting work
openmcf tofu init -f resource.yaml
openmcf tofu plan -f resource.yaml

# ⚠️ Risky: Assuming init was already run
openmcf tofu plan -f resource.yaml  # May fail
```

**Why**: Init is fast and idempotent. Running it ensures providers and backend are ready.

### 4. **Use Descriptive Resource Names**

```yaml
# ✅ Good: Clear, hierarchical naming
metadata:
  name: prod-api-database

# ❌ Bad: Generic, unclear names
metadata:
  name: db1
```

**Why**: Good names make it obvious what infrastructure the manifest manages.

### 5. **Test Changes in Lower Environments First**

```bash
# ✅ Good: Progressive deployment
openmcf tofu apply --kustomize-dir services/api --overlay dev
# Test in dev...
openmcf tofu apply --kustomize-dir services/api --overlay staging
# Test in staging...
openmcf tofu apply --kustomize-dir services/api --overlay prod

# ❌ Bad: YOLO to production
openmcf tofu apply --kustomize-dir services/api --overlay prod --auto-approve
```

**Why**: Lower environments catch issues before they impact production.

### 6. **Use `--set` for Temporary Overrides Only**

```bash
# ✅ Good: Quick testing
openmcf tofu plan \
  -f deployment.yaml \
  --set spec.replicas=1  # Test with minimal resources

# ❌ Bad: Permanent changes via flag
# (6 months later: "Why is prod running 1 replica?!")
openmcf tofu apply \
  -f deployment.yaml \
  --set spec.replicas=1 \
  --auto-approve
```

**Why**: Flags don't persist. Commit important changes to your manifest.

### 7. **Document Provider Credentials**

```bash
# ✅ Good: Document in README
# ops/README.md
# Deploy with:
#   export CLOUDFLARE_API_TOKEN=$(pass cloudflare/api-token)
#   openmcf tofu apply -f r2-bucket.yaml

# ⚠️ Bad: Tribal knowledge
# (New team member: "How do I deploy this?")
```

**Why**: Documentation prevents "works on my machine" situations.

### 8. **Clean Up State Files After Destroy**

```bash
# After destroying resources, consider cleaning up state file
openmcf tofu destroy -f temp-resource.yaml --auto-approve

# Optionally, remove state file from backend
# (This step depends on your backend - S3, GCS, etc.)
aws s3 rm s3://my-terraform-state/path/to/state/terraform.tfstate
```

**Why**: Prevents confusion from orphaned state files.

---

## Tips & Tricks

### Viewing State

```bash
# View current state (requires being in module directory)
cd <module-directory>
tofu state list
tofu state show <resource-address>
```

### Targeting Specific Resources

```bash
# Apply changes to only specific resources (requires being in module directory)
cd <module-directory>
tofu apply -target=aws_s3_bucket.my_bucket
```

### Viewing Outputs

```bash
# View outputs from last apply (requires being in module directory)
cd <module-directory>
tofu output
tofu output -json  # Get outputs as JSON
```

### Debugging

```bash
# Enable verbose logging
export TF_LOG=DEBUG
openmcf tofu plan -f resource.yaml

# Or trace level for maximum verbosity
export TF_LOG=TRACE
openmcf tofu plan -f resource.yaml
```

### Format Validation

```bash
# Validate Terraform syntax (requires being in module directory)
cd <module-directory>
tofu validate

# Format Terraform files
tofu fmt
```

---

## State Management

### Understanding State

OpenTofu tracks your infrastructure in a **state file** (`terraform.tfstate`). This file maps your manifest to real-world resources.

**State storage options**:
- **Local**: State file on disk (default, not recommended for teams)
- **S3**: State in AWS S3 bucket (common for AWS deployments)
- **GCS**: State in Google Cloud Storage (common for GCP deployments)
- **Azure Blob**: State in Azure Storage (common for Azure deployments)
- **Pulumi Cloud, Terraform Cloud, etc.**: Managed state backends

**Why state matters**:
- Tracks which cloud resources belong to which manifest
- Stores resource IDs and properties
- Enables drift detection (comparing state to reality)
- Supports locking to prevent concurrent modifications

**State file location** is configured in your module's `backend.tf` (or similar configuration).

### State vs Stack

If you're coming from Pulumi:

| Pulumi | OpenTofu |
|--------|----------|
| Stack = org/project/stack | State file + (optional) workspace |
| Stack stored in Pulumi backend | State stored in configured backend |
| `pulumi stack` commands | `tofu state` commands |
| Stack FQDN in manifest labels | State location in module backend config |

---

## Related Documentation

- [OpenTofu Documentation](https://opentofu.org/docs/) - Official OpenTofu documentation
- [Terraform Compatibility](https://opentofu.org/docs/intro/compatibility/) - OpenTofu's compatibility with Terraform
- [Manifest Structure Guide](/docs/guides/manifests) - Understanding OpenMCF manifests
- [Credentials Guide](/docs/guides/credentials) - Setting up cloud provider credentials
- [CLI Reference](/docs/cli/cli-reference) - Complete CLI command reference

---

## Getting Help

**Found a bug?** [Open an issue](https://github.com/plantonhq/openmcf/issues)

**Need support?** Check existing issues or discussions

**Contributing?** Pull requests welcome!

---

**Remember**: Infrastructure as code is code. Apply the same discipline you'd apply to application code—version control, testing, code review, and automation. Your infrastructure deserves it. 🚀

