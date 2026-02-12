---
title: "Pulumi Commands Reference"
description: "Complete guide to managing infrastructure with openmcf pulumi commands - init, preview, up, refresh, destroy, and delete"
icon: "code"
order: 2
---

# Pulumi Commands Reference

Your complete guide to managing infrastructure with `openmcf pulumi` commands.

---

## Overview

Think of Pulumi as your infrastructure's version control system. Just as Git lets you commit, preview diffs, and push code changes, Pulumi's lifecycle commands let you initialize, preview, deploy, refresh, and destroy infrastructure. The `openmcf` CLI wraps these Pulumi operations with manifest-driven workflows, giving you a consistent experience across all cloud providers.

### The Infrastructure Lifecycle

```
┌──────────┐    ┌─────────┐    ┌────────┐    ┌─────────┐    ┌─────────┐    ┌────────┐
│   init   │ -> │ preview │ -> │   up   │ -> │ refresh │ -> │ destroy │ -> │ delete │
└──────────┘    └─────────┘    └────────┘    └─────────┘    └─────────┘    └────────┘
     │               │              │              │              │              │
  Create          Review         Deploy        Sync State     Teardown       Remove
   Stack          Changes       Resources      with Cloud     Resources      Stack
```

### Key Concepts

**Manifest**: A YAML file describing your infrastructure resource (e.g., `r2-bucket.yaml`, `eks-cluster.yaml`). Think of it as a blueprint.

**Stack**: A deployment instance with its own state. The stack name follows the format `<org>/<project>/<stack>` (e.g., `my-org/my-project/prod.CloudflareR2Bucket.my-bucket`). Your manifest contains this in its labels.

**Module Directory**: Where the Pulumi IaC code lives. Usually auto-detected from your manifest's resource kind, but can be overridden with `--module-dir`.

**State Backend**: Where Pulumi stores your infrastructure's state. Configure this once with `pulumi login` (supports Pulumi Cloud, S3, GCS, or local file system).

---

## Commands

### `init` - Initialize a New Stack

**What it does**: Creates a new Pulumi stack in your backend for the resource defined in your manifest. Think of this as "initializing a Git repository" for your infrastructure.

**When to use**:
- First time deploying a new resource
- Creating a new environment (dev/staging/prod) for an existing resource type
- After getting "stack not found" errors from other commands

**Behavior**:
- Reads the stack FQDN from your manifest's `pulumi.openmcf.org/stack.name` label
- Creates the stack in your configured Pulumi backend
- If the stack already exists, gracefully skips initialization (idempotent operation)
- Does NOT create any cloud resources—it only prepares the state tracking

**Usage**:

```bash
openmcf pulumi init -f <manifest-file> [flags]
```

**Examples**:

```bash
# Initialize a Cloudflare R2 bucket stack
openmcf pulumi init \
  -f ops/cloud-resources/prod/r2-bucket.yaml

# Initialize using kustomize overlay (for projects using kustomize)
openmcf pulumi init \
  --kustomize-dir backend/services/api \
  --overlay prod

# Initialize with explicit module directory (for development/testing)
openmcf pulumi init \
  -f ops/resources/vpc.yaml \
  --module-dir ~/projects/custom-modules/aws-vpc
```

**What you'll see**:

```
● Loading manifest...
✔ Manifest loaded
● Validating manifest...
✔ Manifest validated
● Initializing Pulumi stack...

🤝 Handing off to Pulumi...
   Output below is from Pulumi

Using Pulumi stack from manifest labels: my-org/my-project/prod.CloudflareR2Bucket.pipeline-logs

pulumi module directory: /path/to/module
Initializing stack: my-org/my-project/prod.CloudflareR2Bucket.pipeline-logs

Created stack 'my-org/my-project/prod.CloudflareR2Bucket.pipeline-logs'

✓ Successfully initialized stack: my-org/my-project/prod.CloudflareR2Bucket.pipeline-logs

✔ Pulumi execution succeeded
```

---

### `preview` - Preview Infrastructure Changes

**What it does**: Shows you what changes Pulumi will make to your infrastructure without actually applying them. This is like `git diff` before committing—you see what's going to change before it happens.

**When to use**:
- Before running `up` to understand what will be created/modified/deleted
- To validate your manifest changes produce the expected infrastructure changes
- During code review to demonstrate infrastructure changes
- When debugging unexpected behavior

**Behavior**:
- Compares your manifest against the current infrastructure state
- Shows a detailed diff: additions (+), modifications (~), deletions (-)
- Does NOT modify any cloud resources
- Does NOT modify Pulumi state
- Requires the stack to exist (run `init` first if needed)

**Usage**:

```bash
openmcf pulumi preview -f <manifest-file> [flags]
```

**Examples**:

```bash
# Preview changes for a Kubernetes deployment
openmcf pulumi preview \
  -f services/api/deployment.yaml

# Preview with field overrides (useful for testing different configurations)
openmcf pulumi preview \
  -f services/api/deployment.yaml \
  --set spec.replicas=5 \
  --set spec.container.image.tag=v2.0.0

# Preview using kustomize (common for multi-environment setups)
openmcf pulumi preview \
  --kustomize-dir backend/services/api \
  --overlay staging
```

**Reading the output**:

```
Previewing update (my-org/prod.CloudflareR2Bucket.logs):

     Type                        Name                Plan       Info
 +   pulumi:pulumi:Stack         my-org       create
 +   └─ cloudflare:R2Bucket      bucket              create

Resources:
    + 2 to create

```

**Legend**:
- `+` = Resource will be created
- `~` = Resource will be modified (shows detailed diff of what changes)
- `-` = Resource will be deleted
- `+-` = Resource will be replaced (delete + create, often due to immutable property changes)

---

### `up` (or `update`) - Deploy Infrastructure

**What it does**: Applies your manifest to create, update, or configure cloud resources. This is the "make it so" command—it actually executes the infrastructure changes.

**When to use**:
- After reviewing changes with `preview` and confirming they look correct
- Initial deployment of new infrastructure
- Updating existing infrastructure with new configurations
- Applying configuration changes (scaling, updates, feature flags)

**Behavior**:
- Creates the stack if it doesn't exist (no need to run `init` separately)
- Shows a preview of changes (unless you use `--yes` flag)
- Waits for your confirmation before proceeding (unless `--yes` is provided)
- Creates/updates/deletes cloud resources to match your manifest
- Updates Pulumi state to reflect the new infrastructure state
- Rolls back automatically if deployment fails (where provider supports it)

**Usage**:

```bash
openmcf pulumi up -f <manifest-file> [flags]
```

**Aliases**: You can use `update` or `up` interchangeably.

**Examples**:

```bash
# Interactive deployment (will show preview and ask for confirmation)
openmcf pulumi up \
  -f ops/resources/database.yaml

# Non-interactive deployment (CI/CD pipelines)
openmcf pulumi up \
  -f ops/resources/database.yaml \
  --yes

# Deploy with field overrides
openmcf pulumi up \
  -f ops/resources/cache.yaml \
  --set spec.instanceSize=large \
  --set spec.replicas=3

# Deploy using kustomize overlay
openmcf pulumi up \
  --kustomize-dir backend/services/worker \
  --overlay prod \
  --yes
```

**What you'll see**:

```
● Loading manifest...
✔ Manifest loaded
● Validating manifest...
✔ Manifest validated
● Preparing Pulumi execution...
✔ Execution prepared

🤝 Handing off to Pulumi...
   Output below is from Pulumi

Using Pulumi stack from manifest labels: my-org/my-project/prod.GcpCloudSql.main-db

Previewing update (my-org/my-project/prod.GcpCloudSql.main-db):

     Type                             Name                          Plan
 +   pulumi:pulumi:Stack              my-org                 create
 +   ├─ gcp:sql:DatabaseInstance      main-db                       create
 +   ├─ gcp:sql:Database              app-db                        create
 +   └─ gcp:sql:User                  app-user                      create

Resources:
    + 4 to create

Do you want to perform this update? yes
Updating (my-org/my-project/prod.GcpCloudSql.main-db):

     Type                             Name                          Status
 +   pulumi:pulumi:Stack              my-org                 created (3s)
 +   ├─ gcp:sql:DatabaseInstance      main-db                       created (185s)
 +   ├─ gcp:sql:Database              app-db                        created (8s)
 +   └─ gcp:sql:User                  app-user                      created (5s)

Outputs:
    connection_name: "my-project:us-central1:main-db-xyz123"
    database_name:   "app-db"

Resources:
    + 4 created

Duration: 3m21s

✔ Pulumi execution succeeded
```

**Important Notes**:

- **Preview-then-apply workflow**: By default, `up` shows you a preview and waits for confirmation. This is your safety net.
- **Automatic stack creation**: Unlike `preview`, `up` will create the stack if it doesn't exist, so you don't always need to run `init` first.
- **State locking**: Pulumi automatically locks state during updates to prevent concurrent modifications.
- **Failed updates**: If an update fails midway, Pulumi state will reflect the partial changes. You can re-run `up` to continue or use `refresh` to sync state.

---

### `refresh` - Sync State with Reality

**What it does**: Compares Pulumi's state file against the actual resources in your cloud provider and updates the state to match reality. Think of this as "git fetch" for infrastructure—it brings your local understanding up to date with what's actually deployed.

**When to use**:
- After manual changes made outside Pulumi (e.g., via cloud console, CLI, or other tools)
- Before running `up` or `destroy` to ensure state accuracy
- After failed deployments to resynchronize state
- When troubleshooting drift between desired and actual state
- After importing existing resources into Pulumi management

**Behavior**:
- Queries your cloud provider for the current state of managed resources
- Updates Pulumi state to reflect actual resource properties
- Does NOT modify any cloud resources
- Does NOT change your manifest file
- Shows what changed in the state (if anything)
- Detects resources that were deleted outside Pulumi

**Usage**:

```bash
openmcf pulumi refresh -f <manifest-file> [flags]
```

**Examples**:

```bash
# Refresh to sync state after manual changes
openmcf pulumi refresh \
  -f ops/resources/s3-bucket.yaml

# Non-interactive refresh (for automation)
openmcf pulumi refresh \
  -f ops/resources/s3-bucket.yaml \
  --yes

# Refresh before important operations
openmcf pulumi refresh \
  -f ops/resources/production-db.yaml \
  --yes && \
openmcf pulumi up \
  -f ops/resources/production-db.yaml
```

**What you'll see**:

```
● Loading manifest...
✔ Manifest loaded
● Validating manifest...
✔ Manifest validated
● Preparing Pulumi execution...
✔ Execution prepared

🤝 Handing off to Pulumi...
   Output below is from Pulumi

Refreshing (my-org/my-project/prod.AwsS3Bucket.assets):

     Type                    Name                Status       Info
     pulumi:pulumi:Stack     my-org
 ~   └─ aws:s3:Bucket        assets              updated      [diff: ~tags]

Outputs:
    bucket_name: "assets-prod-xyz123"

Resources:
    ~ 1 updated
    1 unchanged

Duration: 5s

✔ Pulumi execution succeeded
```

**Understanding Drift**:

Drift occurs when someone (or something) modifies infrastructure outside of Pulumi:

```
Before refresh:
  Pulumi State: bucket versioning = false

Actual Cloud:
  AWS Console: someone enabled versioning = true

After refresh:
  Pulumi State: bucket versioning = true (synced!)
```

**Next steps after refresh**:
1. If changes match your manifest → You're good, carry on
2. If unexpected changes → Investigate who/what made them
3. If you want to revert manual changes → Update manifest if needed, then run `up`

---

### `destroy` - Teardown Infrastructure

**What it does**: Deletes all cloud resources managed by the Pulumi stack. This is the "rm -rf" of infrastructure—use with caution. The stack itself remains (with empty state) unless you manually delete it.

**When to use**:
- Tearing down temporary environments (dev, testing, ephemeral previews)
- Decommissioning infrastructure that's no longer needed
- Cleaning up after testing or experimentation
- Cost optimization (shutting down unused resources)
- Before major refactoring (destroy old, deploy new)

**Behavior**:
- Shows a preview of resources to be deleted
- Waits for explicit confirmation (unless `--yes` is provided)
- Deletes resources in reverse dependency order (children before parents)
- Updates Pulumi state to reflect deletion
- Leaves the stack itself intact (but with no resources)
- **Cannot be undone** once confirmed

**Usage**:

```bash
openmcf pulumi destroy -f <manifest-file> [flags]
```

**Examples**:

```bash
# Interactive destroy (will ask for confirmation)
openmcf pulumi destroy \
  -f ops/resources/dev-cluster.yaml

# Non-interactive destroy (automation/CI)
openmcf pulumi destroy \
  -f ops/resources/test-environment.yaml \
  --yes

# Destroy temporary environment
openmcf pulumi destroy \
  --kustomize-dir backend/services/api \
  --overlay pr-123 \
  --yes
```

**What you'll see**:

```
● Loading manifest...
✔ Manifest loaded
● Validating manifest...
✔ Manifest validated
● Preparing Pulumi execution...
✔ Execution prepared

🤝 Handing off to Pulumi...
   Output below is from Pulumi

Previewing destroy (my-org/my-project/dev.GkeCluster.test-cluster):

     Type                              Name                      Plan
 -   pulumi:pulumi:Stack               my-org             delete
 -   ├─ gcp:container:Cluster          test-cluster              delete
 -   └─ gcp:container:NodePool         default-pool              delete

Resources:
    - 3 to delete

Do you want to perform this destroy? yes
Destroying (my-org/my-project/dev.GkeCluster.test-cluster):

     Type                              Name                      Status
 -   pulumi:pulumi:Stack               my-org             deleted
 -   ├─ gcp:container:NodePool         default-pool              deleted (90s)
 -   └─ gcp:container:Cluster          test-cluster              deleted (180s)

Resources:
    - 3 deleted

Duration: 4m35s

✔ Pulumi execution succeeded
```

**⚠️ Safety Warnings**:

1. **Permanent deletion**: Most cloud providers permanently delete resources. Some have soft-delete/trash, but don't count on it.
2. **Data loss**: Databases, storage buckets, and other stateful resources will lose their data unless you have backups.
3. **Dependency risk**: If other resources depend on what you're destroying, they may break.
4. **No undo**: Once you confirm, there's no rollback. The resources are gone.

**Best Practices**:

```bash
# ✅ Good: Review before destroying
openmcf pulumi preview -f prod.yaml   # See what exists
openmcf pulumi destroy -f prod.yaml   # Interactive confirmation

# ⚠️ Risky: Blind destruction
openmcf pulumi destroy -f prod.yaml --yes

# ✅ Good: Backup data first
aws s3 sync s3://my-bucket ./backup-$(date +%Y%m%d)/
openmcf pulumi destroy -f s3-bucket.yaml

# ✅ Good: Verify manifest before destroying
cat prod.yaml  # Make absolutely sure this is the right file
openmcf pulumi destroy -f prod.yaml
```

---

### `delete` (or `rm`) - Remove Stack Metadata

**What it does**: Deletes a Pulumi stack and all its configuration/state from the backend. This removes the stack metadata itself, not the cloud resources. Think of this as "deleting the Git repository" for your infrastructure tracking.

**When to use**:
- After destroying all resources and you no longer need the stack tracking
- Cleaning up stack metadata for decommissioned projects
- Removing accidentally created or test stacks
- Freeing up stack names for reuse

**Behavior**:
- Removes the stack from your Pulumi backend (state storage)
- Deletes all stack configuration and history
- Does NOT destroy cloud resources (run `destroy` first if resources exist)
- By default, refuses to delete stacks that still have resources
- With `--force`, removes stack even if resources exist (dangerous!)
- **Cannot be undone** once executed

**Usage**:

```bash
openmcf pulumi delete -f <manifest-file> [flags]
```

**Aliases**: You can use `delete` or `rm` interchangeably.

**Examples**:

```bash
# Standard workflow: destroy resources first, then remove stack
openmcf pulumi destroy \
  -f ops/resources/temp-env.yaml \
  --yes

# After resources are gone, remove the stack metadata
openmcf pulumi delete \
  -f ops/resources/temp-env.yaml

# Using the 'rm' alias
openmcf pulumi rm \
  -f ops/resources/old-stack.yaml

# Force removal (skip resource check) - use with extreme caution
openmcf pulumi delete \
  -f ops/resources/abandoned-stack.yaml \
  --force

# Remove stack via explicit stack FQDN
openmcf pulumi delete \
  -f ops/resources/resource.yaml \
  --stack my-org/my-project/dev.TestStack.old
```

**What you'll see**:

```
● Loading manifest...
✔ Manifest loaded
● Validating manifest...
✔ Manifest validated
● Deleting Pulumi stack...

🤝 Handing off to Pulumi...
   Output below is from Pulumi

Using Pulumi stack from manifest labels: my-org/my-project/dev.TestResource.temp

pulumi module directory: /path/to/module
Removing stack: my-org/my-project/dev.TestResource.temp

Stack 'my-org/my-project/dev.TestResource.temp' has been removed!

✓ Successfully removed stack: my-org/my-project/dev.TestResource.temp

✔ Pulumi execution succeeded
```

**⚠️ Critical Warnings**:

1. **Resources check**: By default, Pulumi refuses to delete stacks that still have resources. This is your safety net.
2. **Destroy first**: Always run `destroy` before `delete` to properly clean up cloud resources.
3. **State loss**: Once deleted, you lose all stack history, outputs, and configuration. No undo.
4. **Force flag danger**: Using `--force` bypasses resource checks. Only use if you're absolutely certain resources are gone or managed elsewhere.
5. **Orphaned resources**: If you force-delete a stack with resources, those resources become orphaned (unmanaged by Pulumi).

**Difference: `destroy` vs `delete`**:

```
destroy:
  - Tears down cloud resources (VMs, databases, etc.)
  - Leaves stack metadata intact
  - Updates state to reflect empty stack
  - Resources are gone, but Pulumi still tracks the stack

delete (rm):
  - Removes stack metadata from backend
  - Does NOT touch cloud resources
  - Pulumi stops tracking this stack entirely
  - Used AFTER destroy to clean up metadata
```

**Recommended Workflow**:

```bash
# Step 1: Verify what resources exist
openmcf pulumi preview -f my-stack.yaml

# Step 2: Destroy the cloud resources
openmcf pulumi destroy -f my-stack.yaml

# Step 3: Verify resources are gone (should show empty stack)
openmcf pulumi preview -f my-stack.yaml

# Step 4: Remove the stack metadata
openmcf pulumi delete -f my-stack.yaml

# Done! Stack and resources are completely gone
```

**When to use `--force`**:

```bash
# Scenario 1: Stack state is corrupted, resources already manually deleted
# You know resources are gone but Pulumi state is wrong
openmcf pulumi delete -f broken-stack.yaml --force

# Scenario 2: Resources were imported/migrated to another stack
# Original stack should no longer manage them
openmcf pulumi delete -f old-stack.yaml --force

# Scenario 3: Test/development stack with resources you don't care about
# (Still not recommended - better to destroy properly)
openmcf pulumi delete -f test-stack.yaml --force
```

**Best Practices**:

```bash
# ✅ Good: Complete cleanup workflow
openmcf pulumi destroy -f stack.yaml --yes
openmcf pulumi delete -f stack.yaml

# ⚠️ Risky: Forcing without verification
openmcf pulumi delete -f stack.yaml --force

# ✅ Good: Verify stack FQDN before deleting
pulumi stack --stack <stack-fqdn>  # Check what's in the stack
openmcf pulumi delete -f stack.yaml

# ✅ Good: Export state before deleting (backup)
pulumi stack export --stack <stack-fqdn> > backup.json
openmcf pulumi delete -f stack.yaml
```

**Troubleshooting**:

**Error: "Stack still has resources"**

```bash
# Problem: Trying to delete stack with resources
# Solution: Destroy resources first
openmcf pulumi destroy -f stack.yaml
openmcf pulumi delete -f stack.yaml

# Or if resources are actually gone (state is wrong)
openmcf pulumi refresh -f stack.yaml  # Sync state
openmcf pulumi delete -f stack.yaml
```

**Error: "Stack not found"**

```bash
# Problem: Stack already deleted or never existed
# Solution: Verify stack FQDN
pulumi stack ls  # List all stacks
# If not listed, it's already gone (nothing to do)
```

---

## Common Flags

All commands support these flags. They're like the universal remote for infrastructure management.

### Manifest Input

**`-f <file>`**: Path to your resource manifest YAML file.

```bash
openmcf pulumi up -f ops/resources/my-resource.yaml
```

**`--kustomize-dir <dir>`** + **`--overlay <name>`**: Use kustomize for environment-specific configurations.

```bash
# Loads kustomize base + overlays/prod
openmcf pulumi up \
  --kustomize-dir backend/services/api \
  --overlay prod
```

**Priority**: `-f` > `--kustomize-dir` + `--overlay`

### Execution Control

**`--module-dir <path>`**: Override the Pulumi module directory (defaults to current directory).

```bash
# Use local development module instead of released version
openmcf pulumi up \
  -f my-resource.yaml \
  --module-dir ~/projects/custom-modules/my-module
```

**`--stack <org>/<project>/<stack>`**: Explicitly specify stack FQDN (overrides manifest label).

```bash
openmcf pulumi up \
  -f resource.yaml \
  --stack my-org/my-project/custom-stack-name
```

**`--yes`**: Auto-approve without confirmation prompts (for CI/CD).

```bash
openmcf pulumi up -f resource.yaml --yes
```

**`--force`**: Force removal of stack even if resources exist (only for `delete`/`rm` command).

```bash
# Use with extreme caution - only when you're certain resources are gone
openmcf pulumi delete -f resource.yaml --force
```

**`--set <key>=<value>`**: Override manifest values at runtime (repeatable flag).

```bash
openmcf pulumi up \
  -f deployment.yaml \
  --set spec.replicas=10 \
  --set spec.container.image.tag=v2.1.0 \
  --set metadata.env=staging
```

### Provider Credentials

**Default Behavior**: Credentials are automatically loaded from environment variables. If you have your cloud provider CLI configured (e.g., `aws configure`, `gcloud auth`, `az login`), no additional credential setup is needed.

```bash
# Works without -p flag if environment variables are set
openmcf pulumi up -f ops/aws-resources/vpc.yaml  # Uses AWS_ACCESS_KEY_ID, etc.
openmcf pulumi up -f ops/gcp-resources/gke.yaml  # Uses GOOGLE_APPLICATION_CREDENTIALS
```

**Explicit Override**: Use `-p` / `--provider-config` to override environment variables with a credentials file.

**`-p, --provider-config <file>`**: Path to provider credential YAML file

```bash
# Override with explicit credentials file
openmcf pulumi up \
  -f ops/aws-resources/vpc.yaml \
  -p ~/.config/openmcf/credentials/aws-prod.yaml

openmcf pulumi up \
  -f ops/gcp-resources/cluster.yaml \
  -p ~/.config/openmcf/credentials/gcp-prod.yaml
```

The CLI auto-detects which provider is needed from your manifest's `apiVersion`. See the [Credentials Guide](/docs/guides/credentials) for environment variable details.

---

## Common Workflows

### First-Time Deployment

```bash
# 1. Initialize the stack (creates state tracking)
openmcf pulumi init -f my-resource.yaml

# 2. Preview what will be created
openmcf pulumi preview -f my-resource.yaml

# 3. Deploy the infrastructure
openmcf pulumi up -f my-resource.yaml
```

**Shortcut**: `up` creates the stack automatically if it doesn't exist:

```bash
# One command to rule them all (for new stacks)
openmcf pulumi up -f my-resource.yaml
```

### Updating Existing Infrastructure

```bash
# 1. Edit your manifest
vim ops/resources/my-app.yaml

# 2. Preview the changes
openmcf pulumi preview -f ops/resources/my-app.yaml

# 3. Apply if changes look good
openmcf pulumi up -f ops/resources/my-app.yaml
```

### Testing Configuration Changes

```bash
# Preview with overrides (no changes to manifest file)
openmcf pulumi preview \
  -f api-deployment.yaml \
  --set spec.replicas=20 \
  --set spec.resources.limits.cpu=4000m

# If it looks good, apply with same overrides
openmcf pulumi up \
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
openmcf pulumi up \
  -f deployment.yaml \
  --set spec.container.image.tag=v1.9.5

# Option 2: Revert manifest to previous version
git checkout HEAD~1 deployment.yaml
openmcf pulumi up -f deployment.yaml

# Option 3: Use a previous Git revision
git show HEAD~5:deployment.yaml > /tmp/previous-deployment.yaml
openmcf pulumi up -f /tmp/previous-deployment.yaml
```

### Syncing After Manual Changes

```bash
# Someone made changes via AWS console, need to sync state

# 1. Refresh to see what changed
openmcf pulumi refresh -f s3-bucket.yaml

# 2. Review the diff
openmcf pulumi preview -f s3-bucket.yaml

# 3. Decide:
#    - Changes match manifest? → Do nothing, state is synced
#    - Changes don't match? → Update manifest or revert via `up`

# 4. If reverting manual changes:
openmcf pulumi up -f s3-bucket.yaml  # Restores manifest config
```

### Multi-Environment Deployment

```bash
# Using kustomize overlays for different environments

# Deploy to dev
openmcf pulumi up \
  --kustomize-dir services/api \
  --overlay dev

# Preview staging changes
openmcf pulumi preview \
  --kustomize-dir services/api \
  --overlay staging

# Deploy to production (with extra caution)
openmcf pulumi preview \
  --kustomize-dir services/api \
  --overlay prod
# Review carefully...
openmcf pulumi up \
  --kustomize-dir services/api \
  --overlay prod
```

### Local Module Development

```bash
# Testing changes to Pulumi module code without publishing

cd ~/projects/openmcf/apis/.../.../iac/pulumi

# Point to local module directory
openmcf pulumi preview \
  -f ~/manifests/test-resource.yaml \
  --module-dir .

# Iterate: edit module code, run preview again
vim module/main.go
openmcf pulumi preview \
  -f ~/manifests/test-resource.yaml \
  --module-dir .

# Deploy with local module
openmcf pulumi up \
  -f ~/manifests/test-resource.yaml \
  --module-dir .
```

### CI/CD Pipeline

```bash
#!/bin/bash
# deploy.sh - Automated deployment script

set -e  # Exit on error

MANIFEST="ops/resources/app-${ENV}.yaml"

echo "🔍 Previewing changes..."
openmcf pulumi preview -f "$MANIFEST" --yes

echo "🚀 Deploying infrastructure..."
openmcf pulumi up -f "$MANIFEST" --yes

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
      
      - name: Setup Pulumi
        run: |
          pulumi login ${{ secrets.PULUMI_BACKEND_URL }}
        env:
          PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }}
      
      - name: Deploy Resources
        run: |
          openmcf pulumi up \
            -f ops/resources/prod-infra.yaml \
            --yes
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
```

---

## Troubleshooting

### Error: "no stack named '...' found"

**Symptom**: Command fails with stack not found error.

**Cause**: The Pulumi stack hasn't been initialized yet.

**Solution**:

```bash
# Option 1: Run init first
openmcf pulumi init -f my-resource.yaml
openmcf pulumi preview -f my-resource.yaml

# Option 2: Use 'up' which auto-creates stack
openmcf pulumi up -f my-resource.yaml
```

### Error: "another update is currently in progress"

**Symptom**: Command fails saying stack is locked.

**Cause**: A previous operation crashed or is still running, leaving the stack locked.

**Solution**:

```bash
# Check if operation is actually running
pulumi stack --stack <stack-fqdn>

# If no operation is running, cancel the lock
pulumi cancel --stack <stack-fqdn>

# Then retry your operation
openmcf pulumi up -f my-resource.yaml
```

### Provider Authentication Failures

**Symptom**: "failed to create provider" or authentication errors.

**Causes**: Missing or invalid cloud provider credentials.

**Solutions**:

**For AWS**:
```bash
# Check credentials
aws sts get-caller-identity

# Or provide provider config file
openmcf pulumi up \
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
openmcf pulumi up -f resource.yaml
```

**For Cloudflare**:
```bash
# Set API token
export CLOUDFLARE_API_TOKEN="your-token-here"
openmcf pulumi up -f resource.yaml
```

### Preview Shows Unexpected Changes

**Symptom**: `preview` shows modifications you didn't make.

**Causes**:
1. Someone made manual changes outside Pulumi
2. Provider API defaults changed
3. Computed values changed upstream

**Solution**:

```bash
# First, sync state with reality
openmcf pulumi refresh -f resource.yaml

# Then preview again
openmcf pulumi preview -f resource.yaml

# Compare against previous state
pulumi stack --show-urns --stack <stack-fqdn>
```

### State Conflict: Resources Already Exist

**Symptom**: "resource already exists" error during deployment.

**Cause**: Resources exist in cloud but not in Pulumi state (created outside Pulumi or state lost).

**Solution**:

```bash
# Option 1: Import existing resources (advanced)
pulumi import <type> <name> <cloud-resource-id> --stack <stack-fqdn>

# Option 2: Manually delete cloud resources
# (Use cloud provider console/CLI to delete conflicting resources)

# Option 3: Use different resource names in manifest
vim my-resource.yaml  # Change metadata.name or resource IDs
```

---

## Best Practices

### 1. **Always Preview Before Applying**

```bash
# ✅ Good: Review changes first
openmcf pulumi preview -f resource.yaml
# Read output, verify changes look correct
openmcf pulumi up -f resource.yaml

# ⚠️ Risky: Blind deployment
openmcf pulumi up -f resource.yaml --yes
```

**Why**: Preview is your safety net. It catches mistakes before they become expensive incidents.

### 2. **Use Version Control for Manifests**

```bash
# ✅ Good: Track changes in Git
git add ops/resources/my-resource.yaml
git commit -m "feat: increase database instance size"
git push
# Deploy via CI/CD or manually

# ❌ Bad: Direct edits without version control
vim /tmp/my-resource.yaml
openmcf pulumi up -f /tmp/my-resource.yaml
```

**Why**: Version control gives you change history, rollback capability, and code review.

### 3. **Refresh Before Important Operations**

```bash
# ✅ Good: Sync state before major changes
openmcf pulumi refresh -f resource.yaml --yes
openmcf pulumi up -f resource.yaml

# ⚠️ Risky: Operating on stale state
# (Someone made manual changes you don't know about)
openmcf pulumi up -f resource.yaml
```

**Why**: Refreshing prevents conflicts and ensures you're working with accurate state.

### 4. **Use Descriptive Stack Names**

```yaml
# ✅ Good: Clear, hierarchical naming
metadata:
  labels:
    pulumi.openmcf.org/stack.name: "my-org/my-project/prod.CloudflareR2Bucket.pipeline-logs"
    #                                       └─────org────┘ └─project──┘ └─────environment.ResourceType.resource-name───┘

# ❌ Bad: Generic, unclear names
metadata:
  labels:
    pulumi.openmcf.org/stack.name: "org1/proj1/stack1"
```

**Why**: Good names make it obvious what infrastructure the stack manages.

### 5. **Test Changes in Lower Environments First**

```bash
# ✅ Good: Progressive deployment
openmcf pulumi up --kustomize-dir services/api --overlay dev
# Test in dev...
openmcf pulumi up --kustomize-dir services/api --overlay staging
# Test in staging...
openmcf pulumi up --kustomize-dir services/api --overlay prod

# ❌ Bad: YOLO to production
openmcf pulumi up --kustomize-dir services/api --overlay prod --yes
```

**Why**: Lower environments catch issues before they impact production.

### 6. **Use `--set` for Temporary Overrides Only**

```bash
# ✅ Good: Quick testing
openmcf pulumi preview \
  -f deployment.yaml \
  --set spec.replicas=1  # Test with minimal resources

# ❌ Bad: Permanent changes via flag
# (6 months later: "Why is prod running 1 replica?!")
openmcf pulumi up \
  -f deployment.yaml \
  --set spec.replicas=1 \
  --yes
```

**Why**: Flags don't persist. Commit important changes to your manifest.

### 7. **Document Provider Credentials**

```bash
# ✅ Good: Document in README
# ops/README.md
# Deploy with:
#   export CLOUDFLARE_API_TOKEN=$(pass cloudflare/api-token)
#   openmcf pulumi up -f r2-bucket.yaml

# ⚠️ Bad: Tribal knowledge
# (New team member: "How do I deploy this?")
```

**Why**: Documentation prevents "works on my machine" situations.

### 8. **Clean Up Unused Stacks**

```bash
# After destroying resources, remove the empty stack
openmcf pulumi destroy -f temp-resource.yaml --yes
pulumi stack rm <stack-fqdn>  # Remove stack metadata

# List all stacks to find abandoned ones
pulumi stack ls
```

**Why**: Stack proliferation makes Pulumi backend harder to manage.

---

## Tips & Tricks

### Quick Stack Status Check

```bash
# View current stack state
pulumi stack --stack <stack-fqdn>

# See all resources in stack
pulumi stack --show-urns --stack <stack-fqdn>

# View outputs
pulumi stack output --stack <stack-fqdn>
```

### Diff Specific Resources

```bash
# Preview changes, grep for specific resource
openmcf pulumi preview -f resource.yaml 2>&1 | grep "aws:s3:Bucket"
```

### Copy Stack Outputs to Clipboard

```bash
# macOS
pulumi stack output connection_string --stack <stack-fqdn> | pbcopy

# Linux
pulumi stack output connection_string --stack <stack-fqdn> | xclip -selection clipboard
```

### Automated Health Checks Post-Deployment

```bash
# deploy-and-verify.sh
openmcf pulumi up -f api-deployment.yaml --yes

# Wait for pods to be ready
kubectl rollout status deployment/api -n production

# Run smoke tests
curl -f https://api.example.com/health || exit 1
```

### Export Stack for Disaster Recovery

```bash
# Export current state
pulumi stack export --stack <stack-fqdn> > stack-backup-$(date +%Y%m%d).json

# Import if state gets corrupted
pulumi stack import --stack <stack-fqdn> < stack-backup-20250105.json
```

---

## Related Documentation

- [Pulumi Concepts](https://www.pulumi.com/docs/intro/concepts/) - Official Pulumi documentation
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

