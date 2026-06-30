# Pulumi Commands Reference

> **📖 For the best reading experience**, view this documentation on the [Planton website](https://planton.dev/docs/cli/pulumi-commands).

Your complete guide to managing infrastructure with `planton pulumi` commands.

---

## Overview

Think of Pulumi as your infrastructure's version control system. Just as Git lets you commit, preview diffs, and push code changes, Pulumi's lifecycle commands let you initialize, preview, deploy, refresh, and destroy infrastructure. The `planton` CLI wraps these Pulumi operations with manifest-driven workflows, giving you a consistent experience across all cloud providers.

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

**Stack**: A deployment instance with its own state. The stack name follows the format `<org>/<project>/<stack>` (e.g., `planton/planton/prod.CloudflareR2Bucket.my-bucket`). Your manifest contains this in its labels.

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
- Reads the stack FQDN from your manifest's `pulumi.planton.dev/stack.name` label
- Creates the stack in your configured Pulumi backend
- If the stack already exists, gracefully skips initialization (idempotent operation)
- Does NOT create any cloud resources—it only prepares the state tracking

**Usage**:

```bash
planton pulumi init --manifest <manifest-file> [flags]
```

**Examples**:

```bash
# Initialize a Cloudflare R2 bucket stack
planton pulumi init \
  --manifest ops/cloud-resources/prod/r2-bucket.yaml

# Initialize using kustomize overlay (for projects using kustomize)
planton pulumi init \
  --kustomize-dir backend/services/api \
  --overlay prod

# Initialize with explicit module directory (for development/testing)
planton pulumi init \
  --manifest ops/resources/vpc.yaml \
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

Using Pulumi stack from manifest labels: planton/planton/prod.CloudflareR2Bucket.pipeline-logs

pulumi module directory: /path/to/module
Initializing stack: planton/planton/prod.CloudflareR2Bucket.pipeline-logs

Created stack 'planton/planton/prod.CloudflareR2Bucket.pipeline-logs'

✓ Successfully initialized stack: planton/planton/prod.CloudflareR2Bucket.pipeline-logs

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
planton pulumi preview --manifest <manifest-file> [flags]
```

**Examples**:

```bash
# Preview changes for a Kubernetes deployment
planton pulumi preview \
  --manifest services/api/deployment.yaml

# Preview with field overrides (useful for testing different configurations)
planton pulumi preview \
  --manifest services/api/deployment.yaml \
  --set spec.replicas=5 \
  --set spec.container.image.tag=v2.0.0

# Preview using kustomize (common for multi-environment setups)
planton pulumi preview \
  --kustomize-dir backend/services/api \
  --overlay staging
```

**Reading the output**:

```
Previewing update (planton/prod.CloudflareR2Bucket.logs):

     Type                        Name                Plan       Info
 +   pulumi:pulumi:Stack         planton       create
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
planton pulumi up --manifest <manifest-file> [flags]
```

**Aliases**: You can use `update` or `up` interchangeably.

**Examples**:

```bash
# Interactive deployment (will show preview and ask for confirmation)
planton pulumi up \
  --manifest ops/resources/database.yaml

# Non-interactive deployment (CI/CD pipelines)
planton pulumi up \
  --manifest ops/resources/database.yaml \
  --yes

# Deploy with field overrides
planton pulumi up \
  --manifest ops/resources/cache.yaml \
  --set spec.instanceSize=large \
  --set spec.replicas=3

# Deploy using kustomize overlay
planton pulumi up \
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

Using Pulumi stack from manifest labels: planton/planton/prod.GcpCloudSql.main-db

Previewing update (planton/planton/prod.GcpCloudSql.main-db):

     Type                             Name                          Plan
 +   pulumi:pulumi:Stack              planton                 create
 +   ├─ gcp:sql:DatabaseInstance      main-db                       create
 +   ├─ gcp:sql:Database              app-db                        create
 +   └─ gcp:sql:User                  app-user                      create

Resources:
    + 4 to create

Do you want to perform this update? yes
Updating (planton/planton/prod.GcpCloudSql.main-db):

     Type                             Name                          Status
 +   pulumi:pulumi:Stack              planton                 created (3s)
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
planton pulumi refresh --manifest <manifest-file> [flags]
```

**Examples**:

```bash
# Refresh to sync state after manual changes
planton pulumi refresh \
  --manifest ops/resources/s3-bucket.yaml

# Non-interactive refresh (for automation)
planton pulumi refresh \
  --manifest ops/resources/s3-bucket.yaml \
  --yes

# Refresh before important operations
planton pulumi refresh \
  --manifest ops/resources/production-db.yaml \
  --yes && \
planton pulumi up \
  --manifest ops/resources/production-db.yaml
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

Refreshing (planton/planton/prod.AwsS3Bucket.assets):

     Type                    Name                Status       Info
     pulumi:pulumi:Stack     planton
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
planton pulumi destroy --manifest <manifest-file> [flags]
```

**Examples**:

```bash
# Interactive destroy (will ask for confirmation)
planton pulumi destroy \
  --manifest ops/resources/dev-cluster.yaml

# Non-interactive destroy (automation/CI)
planton pulumi destroy \
  --manifest ops/resources/test-environment.yaml \
  --yes

# Destroy temporary environment
planton pulumi destroy \
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

Previewing destroy (planton/planton/dev.GkeCluster.test-cluster):

     Type                              Name                      Plan
 -   pulumi:pulumi:Stack               planton             delete
 -   ├─ gcp:container:Cluster          test-cluster              delete
 -   └─ gcp:container:NodePool         default-pool              delete

Resources:
    - 3 to delete

Do you want to perform this destroy? yes
Destroying (planton/planton/dev.GkeCluster.test-cluster):

     Type                              Name                      Status
 -   pulumi:pulumi:Stack               planton             deleted
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
planton pulumi preview --manifest prod.yaml   # See what exists
planton pulumi destroy --manifest prod.yaml   # Interactive confirmation

# ⚠️ Risky: Blind destruction
planton pulumi destroy --manifest prod.yaml --yes

# ✅ Good: Backup data first
aws s3 sync s3://my-bucket ./backup-$(date +%Y%m%d)/
planton pulumi destroy --manifest s3-bucket.yaml

# ✅ Good: Verify manifest before destroying
cat prod.yaml  # Make absolutely sure this is the right file
planton pulumi destroy --manifest prod.yaml
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
planton pulumi delete --manifest <manifest-file> [flags]
```

**Aliases**: You can use `delete` or `rm` interchangeably.

**Examples**:

```bash
# Standard workflow: destroy resources first, then remove stack
planton pulumi destroy \
  --manifest ops/resources/temp-env.yaml \
  --yes

# After resources are gone, remove the stack metadata
planton pulumi delete \
  --manifest ops/resources/temp-env.yaml

# Using the 'rm' alias
planton pulumi rm \
  --manifest ops/resources/old-stack.yaml

# Force removal (skip resource check) - use with extreme caution
planton pulumi delete \
  --manifest ops/resources/abandoned-stack.yaml \
  --force

# Remove stack via explicit stack FQDN
planton pulumi delete \
  --manifest ops/resources/resource.yaml \
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

Using Pulumi stack from manifest labels: planton/planton/dev.TestResource.temp

pulumi module directory: /path/to/module
Removing stack: planton/planton/dev.TestResource.temp

Stack 'planton/planton/dev.TestResource.temp' has been removed!

✓ Successfully removed stack: planton/planton/dev.TestResource.temp

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
planton pulumi preview --manifest my-stack.yaml

# Step 2: Destroy the cloud resources
planton pulumi destroy --manifest my-stack.yaml

# Step 3: Verify resources are gone (should show empty stack)
planton pulumi preview --manifest my-stack.yaml

# Step 4: Remove the stack metadata
planton pulumi delete --manifest my-stack.yaml

# Done! Stack and resources are completely gone
```

**When to use `--force`**:

```bash
# Scenario 1: Stack state is corrupted, resources already manually deleted
# You know resources are gone but Pulumi state is wrong
planton pulumi delete --manifest broken-stack.yaml --force

# Scenario 2: Resources were imported/migrated to another stack
# Original stack should no longer manage them
planton pulumi delete --manifest old-stack.yaml --force

# Scenario 3: Test/development stack with resources you don't care about
# (Still not recommended - better to destroy properly)
planton pulumi delete --manifest test-stack.yaml --force
```

**Best Practices**:

```bash
# ✅ Good: Complete cleanup workflow
planton pulumi destroy --manifest stack.yaml --yes
planton pulumi delete --manifest stack.yaml

# ⚠️ Risky: Forcing without verification
planton pulumi delete --manifest stack.yaml --force

# ✅ Good: Verify stack FQDN before deleting
pulumi stack --stack <stack-fqdn>  # Check what's in the stack
planton pulumi delete --manifest stack.yaml

# ✅ Good: Export state before deleting (backup)
pulumi stack export --stack <stack-fqdn> > backup.json
planton pulumi delete --manifest stack.yaml
```

**Troubleshooting**:

**Error: "Stack still has resources"**

```bash
# Problem: Trying to delete stack with resources
# Solution: Destroy resources first
planton pulumi destroy --manifest stack.yaml
planton pulumi delete --manifest stack.yaml

# Or if resources are actually gone (state is wrong)
planton pulumi refresh --manifest stack.yaml  # Sync state
planton pulumi delete --manifest stack.yaml
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

**`--manifest <file>`**: Path to your resource manifest YAML file.

```bash
planton pulumi up --manifest ops/resources/my-resource.yaml
```

**`--kustomize-dir <dir>`** + **`--overlay <name>`**: Use kustomize for environment-specific configurations.

```bash
# Loads kustomize base + overlays/prod
planton pulumi up \
  --kustomize-dir backend/services/api \
  --overlay prod
```

**`--input-dir <dir>`**: Directory containing `target.yaml` and credential files (legacy pattern).

```bash
planton pulumi up --input-dir ops/stacks/prod-database/
```

**Priority**: `--manifest` > `--input-dir` > `--kustomize-dir` + `--overlay`

### Execution Control

**`--module-dir <path>`**: Override the Pulumi module directory (defaults to current directory).

```bash
# Use local development module instead of released version
planton pulumi up \
  --manifest my-resource.yaml \
  --module-dir ~/projects/custom-modules/my-module
```

**`--stack <org>/<project>/<stack>`**: Explicitly specify stack FQDN (overrides manifest label).

```bash
planton pulumi up \
  --manifest resource.yaml \
  --stack my-org/my-project/custom-stack-name
```

**`--yes`**: Auto-approve without confirmation prompts (for CI/CD).

```bash
planton pulumi up --manifest resource.yaml --yes
```

**`--force`**: Force removal of stack even if resources exist (only for `delete`/`rm` command).

```bash
# Use with extreme caution - only when you're certain resources are gone
planton pulumi delete --manifest resource.yaml --force
```

**`--set <key>=<value>`**: Override manifest values at runtime (repeatable flag).

```bash
planton pulumi up \
  --manifest deployment.yaml \
  --set spec.replicas=10 \
  --set spec.container.image.tag=v2.1.0 \
  --set metadata.env=staging
```

### Credential Injection

These flags inject provider credentials into the stack input (alternative to environment variables):

- **`--aws-credential <file>`**: Path to AWS credential YAML
- **`--azure-credential <file>`**: Path to Azure credential YAML
- **`--gcp-credential <file>`**: Path to GCP credential YAML
- **`--kubernetes-cluster <file>`**: Path to Kubernetes cluster credential YAML
- **`--confluent-credential <file>`**: Path to Confluent Cloud credential YAML
- **`--docker-credential <file>`**: Path to Docker registry credential YAML
- **`--mongodb-atlas-credential <file>`**: Path to MongoDB Atlas credential YAML
- **`--snowflake-credential <file>`**: Path to Snowflake credential YAML

**Example**:

```bash
planton pulumi up \
  --manifest ops/aws-resources/vpc.yaml \
  --aws-credential ~/.config/planton/credentials/aws-prod.yaml
```

---

## Common Workflows

### First-Time Deployment

```bash
# 1. Initialize the stack (creates state tracking)
planton pulumi init --manifest my-resource.yaml

# 2. Preview what will be created
planton pulumi preview --manifest my-resource.yaml

# 3. Deploy the infrastructure
planton pulumi up --manifest my-resource.yaml
```

**Shortcut**: `up` creates the stack automatically if it doesn't exist:

```bash
# One command to rule them all (for new stacks)
planton pulumi up --manifest my-resource.yaml
```

### Updating Existing Infrastructure

```bash
# 1. Edit your manifest
vim ops/resources/my-app.yaml

# 2. Preview the changes
planton pulumi preview --manifest ops/resources/my-app.yaml

# 3. Apply if changes look good
planton pulumi up --manifest ops/resources/my-app.yaml
```

### Testing Configuration Changes

```bash
# Preview with overrides (no changes to manifest file)
planton pulumi preview \
  --manifest api-deployment.yaml \
  --set spec.replicas=20 \
  --set spec.resources.limits.cpu=4000m

# If it looks good, apply with same overrides
planton pulumi up \
  --manifest api-deployment.yaml \
  --set spec.replicas=20 \
  --set spec.resources.limits.cpu=4000m

# Later, commit the changes to manifest
vim api-deployment.yaml  # Make changes permanent
```

### Emergency Rollback

```bash
# Scenario: v2.0.0 deployment has issues, need to roll back to v1.9.5

# Option 1: Override the current manifest
planton pulumi up \
  --manifest deployment.yaml \
  --set spec.container.image.tag=v1.9.5

# Option 2: Revert manifest to previous version
git checkout HEAD~1 deployment.yaml
planton pulumi up --manifest deployment.yaml

# Option 3: Use a previous Git revision
git show HEAD~5:deployment.yaml > /tmp/previous-deployment.yaml
planton pulumi up --manifest /tmp/previous-deployment.yaml
```

### Syncing After Manual Changes

```bash
# Someone made changes via AWS console, need to sync state

# 1. Refresh to see what changed
planton pulumi refresh --manifest s3-bucket.yaml

# 2. Review the diff
planton pulumi preview --manifest s3-bucket.yaml

# 3. Decide:
#    - Changes match manifest? → Do nothing, state is synced
#    - Changes don't match? → Update manifest or revert via `up`

# 4. If reverting manual changes:
planton pulumi up --manifest s3-bucket.yaml  # Restores manifest config
```

### Multi-Environment Deployment

```bash
# Using kustomize overlays for different environments

# Deploy to dev
planton pulumi up \
  --kustomize-dir services/api \
  --overlay dev

# Preview staging changes
planton pulumi preview \
  --kustomize-dir services/api \
  --overlay staging

# Deploy to production (with extra caution)
planton pulumi preview \
  --kustomize-dir services/api \
  --overlay prod
# Review carefully...
planton pulumi up \
  --kustomize-dir services/api \
  --overlay prod
```

### Local Module Development

```bash
# Testing changes to Pulumi module code without publishing

cd ~/projects/planton/apis/.../.../iac/pulumi

# Point to local module directory
planton pulumi preview \
  --manifest ~/manifests/test-resource.yaml \
  --module-dir .

# Iterate: edit module code, run preview again
vim module/main.go
planton pulumi preview \
  --manifest ~/manifests/test-resource.yaml \
  --module-dir .

# Deploy with local module
planton pulumi up \
  --manifest ~/manifests/test-resource.yaml \
  --module-dir .
```

### CI/CD Pipeline

```bash
#!/bin/bash
# deploy.sh - Automated deployment script

set -e  # Exit on error

MANIFEST="ops/resources/app-${ENV}.yaml"

echo "🔍 Previewing changes..."
planton pulumi preview --manifest "$MANIFEST" --yes

echo "🚀 Deploying infrastructure..."
planton pulumi up --manifest "$MANIFEST" --yes

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
          planton pulumi up \
            --manifest ops/resources/prod-infra.yaml \
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
planton pulumi init --manifest my-resource.yaml
planton pulumi preview --manifest my-resource.yaml

# Option 2: Use 'up' which auto-creates stack
planton pulumi up --manifest my-resource.yaml
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
planton pulumi up --manifest my-resource.yaml
```

### Error: "failed to load stack-input"

**Symptom**: Module fails to load configuration from manifest.

**Causes**:
1. Manifest YAML is malformed
2. Required fields are missing
3. Validation constraints not met

**Solution**:

```bash
# Validate manifest structure
cat my-resource.yaml | yq .

# Check for required fields
grep -E "(apiVersion|kind|metadata|spec)" my-resource.yaml

# Validate with Pulumi directly
planton pulumi preview --manifest my-resource.yaml
# Error message will show specific validation failures
```

### Provider Authentication Failures

**Symptom**: "failed to create provider" or authentication errors.

**Causes**: Missing or invalid cloud provider credentials.

**Solutions**:

**For AWS**:
```bash
# Check credentials
aws sts get-caller-identity

# Or provide credential file
planton pulumi up \
  --manifest resource.yaml \
  --aws-credential ~/.aws/credentials-prod.yaml
```

**For GCP**:
```bash
# Check credentials
gcloud auth list
gcloud config get-value project

# Or set environment variable
export GOOGLE_APPLICATION_CREDENTIALS=~/gcp-key.json
planton pulumi up --manifest resource.yaml
```

**For Cloudflare**:
```bash
# Set API token
export CLOUDFLARE_API_TOKEN="your-token-here"
planton pulumi up --manifest resource.yaml
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
planton pulumi refresh --manifest resource.yaml

# Then preview again
planton pulumi preview --manifest resource.yaml

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
planton pulumi preview --manifest resource.yaml
# Read output, verify changes look correct
planton pulumi up --manifest resource.yaml

# ⚠️ Risky: Blind deployment
planton pulumi up --manifest resource.yaml --yes
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
planton pulumi up --manifest /tmp/my-resource.yaml
```

**Why**: Version control gives you change history, rollback capability, and code review.

### 3. **Refresh Before Important Operations**

```bash
# ✅ Good: Sync state before major changes
planton pulumi refresh --manifest resource.yaml --yes
planton pulumi up --manifest resource.yaml

# ⚠️ Risky: Operating on stale state
# (Someone made manual changes you don't know about)
planton pulumi up --manifest resource.yaml
```

**Why**: Refreshing prevents conflicts and ensures you're working with accurate state.

### 4. **Use Descriptive Stack Names**

```yaml
# ✅ Good: Clear, hierarchical naming
metadata:
  labels:
    pulumi.planton.dev/stack.name: "planton/planton/prod.CloudflareR2Bucket.pipeline-logs"
    #                                       └─────org────┘ └─project──┘ └─────environment.ResourceType.resource-name───┘

# ❌ Bad: Generic, unclear names
metadata:
  labels:
    pulumi.planton.dev/stack.name: "org1/proj1/stack1"
```

**Why**: Good names make it obvious what infrastructure the stack manages.

### 5. **Test Changes in Lower Environments First**

```bash
# ✅ Good: Progressive deployment
planton pulumi up --kustomize-dir services/api --overlay dev
# Test in dev...
planton pulumi up --kustomize-dir services/api --overlay staging
# Test in staging...
planton pulumi up --kustomize-dir services/api --overlay prod

# ❌ Bad: YOLO to production
planton pulumi up --kustomize-dir services/api --overlay prod --yes
```

**Why**: Lower environments catch issues before they impact production.

### 6. **Use `--set` for Temporary Overrides Only**

```bash
# ✅ Good: Quick testing
planton pulumi preview \
  --manifest deployment.yaml \
  --set spec.replicas=1  # Test with minimal resources

# ❌ Bad: Permanent changes via flag
# (6 months later: "Why is prod running 1 replica?!")
planton pulumi up \
  --manifest deployment.yaml \
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
#   planton pulumi up --manifest r2-bucket.yaml

# ⚠️ Bad: Tribal knowledge
# (New team member: "How do I deploy this?")
```

**Why**: Documentation prevents "works on my machine" situations.

### 8. **Clean Up Unused Stacks**

```bash
# After destroying resources, remove the empty stack
planton pulumi destroy --manifest temp-resource.yaml --yes
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
planton pulumi preview --manifest resource.yaml 2>&1 | grep "aws:s3:Bucket"
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
planton pulumi up --manifest api-deployment.yaml --yes

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
- [Planton Architecture](../../README.md) - Understanding the CLI architecture
- [Provider Configuration Guide](../../pkg/iac/pulumi/pulumimodule/provider/README.md) - Setting up cloud provider credentials
- [Manifest Schema Reference](../../apis/README.md) - Available resource types and their schemas

---

## Getting Help

**Found a bug?** [Open an issue](https://github.com/plantonhq/planton/issues)

**Need support?** Check existing issues or discussions

**Contributing?** Pull requests welcome! See [CONTRIBUTING.md](../../CONTRIBUTING.md)

---

**Remember**: Infrastructure as code is code. Apply the same discipline you'd apply to application code—version control, testing, code review, and automation. Your infrastructure deserves it. 🚀

