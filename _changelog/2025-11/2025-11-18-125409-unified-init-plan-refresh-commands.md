# Unified Init/Plan/Refresh Commands with Provisioner Auto-Detection

**Date**: November 18, 2025

## Summary

Added three new unified root-level commands (`init`, `plan`, `refresh`) that complete the infrastructure lifecycle management suite alongside the existing `apply` and `destroy` commands. These commands provide a kubectl-like experience by automatically detecting the IaC provisioner from manifest labels and routing to the appropriate CLI (Pulumi, Tofu, or Terraform). Users can now manage their entire infrastructure lifecycle with a consistent, simple command interface regardless of the underlying provisioner.

## Problem Statement

Following the introduction of unified `apply` and `destroy` commands, users still had to switch to provisioner-specific commands for other lifecycle operations:

```bash
# Unified commands (existing)
planton apply -f manifest.yaml
planton destroy -f manifest.yaml

# But still needed provisioner-specific commands
planton pulumi init --manifest manifest.yaml
planton pulumi preview --manifest manifest.yaml
planton pulumi refresh --manifest manifest.yaml

planton tofu init --manifest manifest.yaml
planton tofu plan --manifest manifest.yaml
planton tofu refresh --manifest manifest.yaml
```

### Pain Points

- **Incomplete Abstraction**: Unified commands only covered apply/destroy, forcing users back to provisioner-specific syntax for other operations
- **Inconsistent CI/CD**: Preview/plan operations in pipelines still required provisioner-aware logic
- **Mental Context Switching**: Developers had to remember different command patterns for different lifecycle stages
- **Documentation Complexity**: Guides had to explain both unified and provisioner-specific approaches
- **Barrier to Entry**: New users faced a steeper learning curve with mixed command patterns

## Solution

Extended the unified command pattern to cover the complete infrastructure lifecycle by adding three new root-level commands that follow the same auto-detection pattern as `apply` and `destroy`:

1. **`init`** - Initialize backend or stack
2. **`plan`** (alias: `preview`) - Preview infrastructure changes
3. **`refresh`** - Sync state with cloud reality

All three commands:
- Read the `planton.dev/provisioner` label from manifests
- Auto-route to the appropriate provisioner implementation
- Support interactive provisioner selection when label is missing
- Maintain full backward compatibility with existing provisioner-specific commands

### Complete Unified Lifecycle

Users can now manage their entire infrastructure lifecycle with consistent commands:

```bash
planton init -f manifest.yaml       # Initialize
planton plan -f manifest.yaml       # Preview (or 'preview')
planton apply -f manifest.yaml      # Deploy
planton refresh -f manifest.yaml    # Sync state
planton destroy -f manifest.yaml    # Teardown (or 'delete')
```

## Implementation Details

### New Command Files

Created three new command files following the established pattern from `apply.go` and `destroy.go`:

**1. `cmd/planton/root/init.go`**

Initializes infrastructure backend or stack:

```go
func initHandler(cmd *cobra.Command, args []string) {
    // Load, validate manifest
    // Detect provisioner from label or prompt
    // Route to appropriate provisioner
    switch provType {
    case provisioner.ProvisionerTypePulumi:
        initWithPulumi(...)  // Calls pulumistack.Init()
    case provisioner.ProvisionerTypeTofu:
        initWithTofu(...)    // Calls tofumodule.TofuInit()
    }
}
```

**Key behaviors**:
- Pulumi: Creates stack if it doesn't exist (idempotent operation)
- Tofu: Initializes backend, downloads provider plugins
- Supports Tofu-specific flags: `--backend-type`, `--backend-config`

**2. `cmd/planton/root/plan.go`**

Previews infrastructure changes without applying them:

```go
func planWithPulumi(...) {
    // Use update operation with isUpdatePreview=true
    pulumistack.Run(moduleDir, stackFqdn, targetManifestPath,
        pulumi.PulumiOperationType_update, 
        true,   // isUpdatePreview
        false,  // isAutoApprove
        valueOverrides, showDiff, providerConfigOptions...)
}

func planWithTofu(...) {
    tofumodule.RunCommand(moduleDir, targetManifestPath,
        terraform.TerraformOperationType_plan,
        valueOverrides, true, isDestroyPlan, providerConfigOptions...)
}
```

**Key features**:
- Alias: `preview` for Pulumi-style experience
- Pulumi: Runs preview mode of update operation
- Tofu: Runs plan operation
- Supports `--destroy` flag for destroy preview (Tofu)
- Supports `--diff` flag for detailed diffs (Pulumi)

**3. `cmd/planton/root/refresh.go`**

Syncs state with cloud reality without modifying resources:

```go
func refreshWithPulumi(...) {
    pulumistack.Run(moduleDir, stackFqdn, targetManifestPath,
        pulumi.PulumiOperationType_refresh,
        false, true, valueOverrides, showDiff, providerConfigOptions...)
}

func refreshWithTofu(...) {
    tofumodule.RunCommand(moduleDir, targetManifestPath,
        terraform.TerraformOperationType_refresh,
        valueOverrides, true, false, providerConfigOptions...)
}
```

**Key characteristics**:
- Read-only operation (does not modify cloud resources)
- Always auto-approves (no confirmation needed)
- Queries cloud provider for current resource state
- Updates state file to match reality

### Command Registration

Updated `cmd/planton/root.go` to register the three new commands:

```go
rootCmd.AddCommand(
    root.Apply,
    root.Destroy,
    root.Init,      // New
    root.LoadManifest,
    root.Plan,      // New (with 'preview' alias)
    root.Pulumi,
    root.Refresh,   // New
    root.Tofu,
    root.ValidateManifest,
    root.Version,
)
```

### Shared Implementation Pattern

All five unified commands follow the same execution flow:

```
1. Parse flags and resolve manifest path
   ├─ Support -f/--manifest flag
   ├─ Support --kustomize-dir + --overlay
   └─ Support --input-dir

2. Apply value overrides (--set flags)

3. Validate manifest

4. Detect provisioner
   ├─ Read planton.dev/provisioner label
   ├─ If missing: prompt user interactively
   └─ Default to Pulumi if no input

5. Prepare provider configs

6. Route to provisioner implementation
   ├─ Pulumi: Call appropriate pulumistack function
   ├─ Tofu: Call tofumodule function
   └─ Terraform: Return "not yet implemented"

7. Print success/failure with appropriate helper
```

### Supported Flags

All commands support common flags plus provisioner-specific options:

**Common Flags (All Commands)**:
- `-f, --manifest` - Manifest file path (kubectl-style)
- `--kustomize-dir` - Kustomize base directory
- `--overlay` - Kustomize overlay name
- `--set` - Field overrides (key=value pairs)
- `--module-dir` - Provisioner module directory

**Pulumi-Specific**:
- `--stack` - Stack FQDN override
- `--diff` - Show detailed diffs (plan, refresh, apply, destroy)

**Tofu/Terraform-Specific**:
- `--destroy` - Destroy plan mode (plan command only)
- `--backend-type` - Backend type (init command only)
- `--backend-config` - Backend configuration (init command only)

**Provider Credentials** (All Commands):
- `--aws-provider-config`
- `--azure-provider-config`
- `--gcp-provider-config`
- `--kubernetes-provider-config`
- `--cloudflare-provider-config`
- `--confluent-provider-config`
- `--atlas-provider-config`
- `--snowflake-provider-config`

## Benefits

### 1. Complete Lifecycle Unification

Users now have unified commands for every infrastructure operation:

| Operation | Unified Command | Pulumi Equivalent | Tofu Equivalent |
|-----------|----------------|-------------------|-----------------|
| Initialize | `init` | `pulumi init` | `tofu init` |
| Preview | `plan` / `preview` | `pulumi preview` | `tofu plan` |
| Deploy | `apply` | `pulumi up` | `tofu apply` |
| Sync | `refresh` | `pulumi refresh` | `tofu refresh` |
| Destroy | `destroy` / `delete` | `pulumi destroy` | `tofu destroy` |

### 2. Simplified CI/CD Pipelines

Pull request pipelines can use unified commands:

```yaml
# .gitlab-ci.yml - works for any provisioner
stages:
  - preview
  - deploy

preview:
  stage: preview
  script:
    - planton plan -f deployment.yaml
  only:
    - merge_requests

deploy:
  stage: deploy
  script:
    - planton apply -f deployment.yaml --yes
  only:
    - main
```

### 3. Consistent Documentation

Documentation can focus on unified commands without provisioner-specific branches:

```bash
# Getting Started - One Pattern for All
planton init -f app.yaml
planton plan -f app.yaml
planton apply -f app.yaml
```

### 4. Better Developer Experience

Developers can focus on infrastructure intent rather than tool syntax:

```bash
# Before: Need to remember provisioner
cd pulumi-resources/
planton pulumi preview --manifest db.yaml --stack org/proj/env

cd ../tofu-resources/
planton tofu plan --manifest vpc.yaml

# After: Same commands everywhere
planton plan -f db.yaml
planton plan -f vpc.yaml
```

### 5. Aliasing for Different Backgrounds

Command aliases support different user preferences:

```bash
# Terraform/Tofu users
planton plan -f app.yaml

# Pulumi users
planton preview -f app.yaml

# kubectl users
planton delete -f app.yaml
```

## Usage Examples

### Example 1: Complete Lifecycle Workflow

```bash
# Full infrastructure lifecycle with unified commands
cd my-service/

# 1. Initialize backend/stack (first time or after cache cleanup)
planton init -f database.yaml

# 2. Preview changes before applying
planton plan -f database.yaml

# 3. Deploy infrastructure
planton apply -f database.yaml --yes

# 4. After manual console changes, sync state
planton refresh -f database.yaml

# 5. Preview changes again
planton plan -f database.yaml

# 6. Destroy when done
planton destroy -f database.yaml
```

### Example 2: Multi-Environment Deployment

```bash
# Deploy across environments with unified commands
for env in dev staging prod; do
    echo "Deploying to $env..."
    planton init --kustomize-dir services/api --overlay $env
    planton plan --kustomize-dir services/api --overlay $env
    planton apply --kustomize-dir services/api --overlay $env --yes
done
```

### Example 3: CI/CD with Preview Gates

```bash
#!/bin/bash
# deploy.sh - CI/CD deployment script

set -e

# Initialize (idempotent - safe to run every time)
planton init -f deployment.yaml

# In pull requests: preview only
if [ "$CI_PIPELINE_SOURCE" = "merge_request_event" ]; then
    echo "Previewing changes for PR..."
    planton plan -f deployment.yaml \
        --set spec.image.tag="$CI_COMMIT_SHA"
    exit 0
fi

# In main branch: deploy
echo "Deploying to $CI_ENVIRONMENT_NAME..."
planton apply -f deployment.yaml \
    --set spec.image.tag="$CI_COMMIT_SHA" \
    --set metadata.labels.environment="$CI_ENVIRONMENT_NAME" \
    --yes
```

### Example 4: OpenTofu with S3 Backend

```yaml
# manifest.yaml
apiVersion: aws.planton.dev/v1
kind: AwsVpc
metadata:
  name: production-vpc
  labels:
    planton.dev/provisioner: tofu
    terraform.planton.dev/backend.type: s3
    terraform.planton.dev/backend.object: vpc/prod.tfstate
spec:
  cidrBlock: 10.0.0.0/16
```

```bash
# Initialize with S3 backend
planton init -f manifest.yaml \
    --backend-type s3 \
    --backend-config bucket=my-terraform-state \
    --backend-config key=vpc/prod.tfstate \
    --backend-config region=us-west-2

# Plan changes
planton plan -f manifest.yaml

# Apply
planton apply -f manifest.yaml --auto-approve
```

### Example 5: Destroy Preview (Tofu)

```bash
# Preview what would be destroyed without actually destroying
planton plan -f app.yaml --destroy

# Review the output, then destroy if acceptable
planton destroy -f app.yaml --auto-approve
```

## Documentation Updates

### Updated Files

**1. `site/public/docs/cli/cli-reference.md`**

- Added `init`, `plan`, and `refresh` to command tree
- Created detailed sections for each command with usage examples
- Documented provisioner routing behavior
- Updated flag documentation

**2. `site/public/docs/cli/unified-commands.md`**

- Renamed from "Unified Apply/Destroy Commands" to "Unified Commands"
- Expanded to cover complete infrastructure lifecycle (5 commands)
- Added sections for `init`, `plan`/`preview`, and `refresh`
- Updated all examples to show complete workflows
- Added CI/CD pipeline examples with preview gates
- Expanded best practices from 5 to 9 practices
- Updated migration guide with all command mappings
- Enhanced benefits section emphasizing lifecycle coverage

### Documentation Highlights

**Complete Lifecycle Examples**: Shows init → plan → apply → refresh → destroy workflows

**CI/CD Integration**: Demonstrates preview gates in pull request pipelines

**Command Aliases**: Documents `plan`/`preview` and `destroy`/`delete` aliases

**Provisioner Routing**: Clear explanation of how commands route to different provisioners

**Flag Documentation**: Comprehensive tables showing which flags work with which commands

**Best Practices**: 9 practices including:
- Following complete lifecycle
- Always previewing in CI/CD
- Using refresh after manual changes
- Leveraging command aliases
- Organizing environments with kustomize

## Impact

### Developer Workflow

**Before**: Mixed command patterns
```bash
# Pulumi workflow
planton pulumi init --manifest app.yaml --stack org/proj/dev
planton pulumi preview --manifest app.yaml --stack org/proj/dev
planton apply -f app.yaml  # Unified
planton pulumi refresh --manifest app.yaml --stack org/proj/dev
planton destroy -f app.yaml  # Unified
```

**After**: Consistent unified commands
```bash
# Works for any provisioner
planton init -f app.yaml
planton plan -f app.yaml
planton apply -f app.yaml
planton refresh -f app.yaml
planton destroy -f app.yaml
```

### CI/CD Pipelines

- Simplified pipeline definitions
- No provisioner-specific logic needed
- Easier to template and reuse
- Preview gates work uniformly across all provisioners

### Documentation & Training

- Single command reference for all operations
- Reduced learning curve for new team members
- Consistent examples across all guides
- No need to teach multiple provisioner syntaxes

### Backward Compatibility

- All existing provisioner-specific commands continue to work
- Zero breaking changes
- Users can migrate gradually at their own pace
- Mixed usage (unified + provisioner-specific) is fully supported

## Testing & Verification

### Build Verification

- ✅ All commands compile successfully
- ✅ No linter errors
- ✅ Binary builds for all platforms (darwin-amd64, darwin-arm64, linux)

### Command Verification

```bash
# Verify all commands registered
$ planton --help | grep -E "(init|plan|refresh)"
  init              initialize backend/stack using the provisioner...
  plan              preview infrastructure changes using the provisioner...
  refresh           sync state with cloud reality using the provisioner...

# Verify aliases
$ planton preview --help | head -1
Preview infrastructure changes by automatically routing...

# Verify plan command
$ planton plan --help
Usage:
  planton plan [flags]

Aliases:
  plan, preview
```

### Flag Verification

- ✅ Common flags work across all commands
- ✅ Pulumi-specific flags (`--stack`, `--diff`) function correctly
- ✅ Tofu-specific flags (`--destroy`, `--backend-type`, `--backend-config`) work as expected
- ✅ Provider credential flags available on all commands

## Related Work

### Previous Changelogs

- **2025-11-18: Unified Apply/Destroy Commands** - Introduced the unified command pattern for deploy and teardown operations
- This changelog completes the unified command suite by adding init, plan, and refresh

### Provisioner Auto-Detection

All unified commands leverage the existing provisioner detection infrastructure:

- `pkg/iac/provisioner/provisioner.go` - Provisioner type detection
- `pkg/iac/provisionerlabels/labels.go` - Label constants
- `internal/cli/prompt/prompt.go` - Interactive provisioner selection

### Underlying Implementations

Commands delegate to existing provisioner implementations:

**Pulumi**:
- `pkg/iac/pulumi/pulumistack/init.go`
- `pkg/iac/pulumi/pulumistack/run.go` (with operation types)

**Tofu**:
- `pkg/iac/tofu/tofumodule/tofu_init.go`
- `pkg/iac/tofu/tofumodule/run_command.go`

## Future Enhancements

### Planned Improvements

1. **Terraform Support**: Currently shows "not yet implemented" - add full Terraform support
2. **Auto-detection Fallback**: Detect provisioner from resource Kind when label is missing
3. **Watch Mode**: Add `--watch` flag for continuous reconciliation
4. **Unified Diff Command**: Add standalone `planton diff` command
5. **State Management**: Add unified commands for state operations (export, import, etc.)

### Potential Additions

- **Cost Estimation**: Preview cost impact of changes (Infracost integration)
- **Policy Checks**: Integrate policy-as-code validation (OPA, Sentinel)
- **Approval Workflows**: Built-in approval gates for production deployments
- **Rollback Support**: Unified rollback command to revert to previous state

## Breaking Changes

**None**. This feature is completely additive and maintains full backward compatibility.

## Migration Guidance

### Recommended Migration Path

1. **Add Provisioner Label** to manifests:
   ```yaml
   metadata:
     labels:
       planton.dev/provisioner: pulumi  # or tofu
   ```

2. **Update CI/CD Pipelines** gradually:
   ```yaml
   # Update one stage at a time
   - planton plan -f app.yaml              # New
   - planton pulumi up --manifest app.yaml # Old (still works)
   ```

3. **Update Documentation** and runbooks to reference unified commands

4. **Train Team** on the new command patterns

5. **Monitor** for any issues during transition

### No Forced Migration

- Existing commands continue to work indefinitely
- Teams can adopt unified commands at their own pace
- Mixed usage is fully supported

---

**Status**: ✅ Production Ready  
**Timeline**: Implemented November 18, 2025 (1 day development + documentation)  
**Code Changed**: 4 files added, 1 file modified, ~1000 lines total

