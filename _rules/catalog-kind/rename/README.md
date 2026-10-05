# Rename: Systematic Kind Renaming

## Overview

The Rename system provides automated, comprehensive renaming of kinds across the entire Planton codebase. It handles all aspects of a rename: file operations, find-replace patterns, registry updates, documentation changes, and build verification.

**Core Philosophy**: Rename is about naming accuracy and clarity, not functionality changes. When a kind's name doesn't accurately reflect what it does, renaming restores semantic truth to the codebase.

## When to Use Rename

### Good Reasons to Rename

**✅ Remove Abstractions**

Example: `KubernetesMicroservice` → `KubernetesDeployment`

The original name "Microservice" is an abstraction. The kind creates a Kubernetes Deployment resource, so "Deployment" is more accurate. This rename:
- Removes unnecessary abstraction
- Clarifies what Kubernetes resource gets created
- Sets stage for `KubernetesStatefulSet` (which would be confusing alongside "Microservice")
- Aligns with existing pattern (`KubernetesCronJob` accurately describes the resource)

**✅ Establish Consistency**

Example: Suffix pattern → Prefix pattern (from 2025-11 workload refactoring)

All 23 Kubernetes workload kinds were renamed from `PostgresKubernetes` to `KubernetesPostgres` to establish consistent prefix pattern across the codebase. This:
- Improved visual grouping (all Kubernetes resources sort together)
- Aligned with Kubernetes ecosystem conventions (`KubeProxy`, `KubeDNS`)
- Made naming consistent with addon operators

**✅ Improve Clarity**

Example: `AwsEcsService` → `AwsEcsFargateService`

If a kind specifically targets Fargate, the name should reflect that specificity.

**✅ Prepare for Expansion**

Example: Rename before adding similar kinds

Before adding `KubernetesStatefulSet`, rename `KubernetesMicroservice` to `KubernetesDeployment` so the two can coexist with clear semantic distinction.

### Poor Reasons to Rename

**❌ Fixing Functionality**

If the kind behavior is wrong, use `@fix-catalog-kind` instead. Rename doesn't change behavior.

**❌ Completing Implementation**

If the kind is incomplete, use `@complete-catalog-kind` or `@update-catalog-kind`. Rename assumes the kind works correctly, just has wrong name.

**❌ Typos in Documentation**

Direct file edits are faster for documentation fixes. Rename is for systematic name changes across the entire codebase.

**❌ Just Because**

Every rename has a cost (user manifest updates, potential confusion). Only rename when there's a clear semantic improvement.

## Architecture

### Kinds

```
rename/
├── rename-catalog-kind.mdc    # Cursor rule (interactive workflow)
├── README.md                                 # This file
└── _scripts/
    └── rename_kind.py       # Python script (execution)
```

### Data Flow

```
User Invocation
      ↓
@rename-catalog-kind (Cursor Rule)
      ↓
Interactive Questions (old name, new name, ID prefix)
      ↓
Confirmation Summary
      ↓
rename_kind.py (Python Script)
      ↓
   [Validate]
      ↓
   [Copy Directory]
      ↓
   [Apply 7 Replacement Patterns]
      ↓
   [Update Registry]
      ↓
   [Delete Old Directory]
      ↓
   [Run: make protos]
      ↓
   [Run: go build ./catalog/...]
      ↓
   [Run: go test -v ./catalog/...]
      ↓
JSON Output (success + metrics)
      ↓
@create-planton-changelog (if success)
      ↓
Commit Guidance
```

## The Seven Naming Patterns

The rename system applies seven comprehensive replacement patterns to cover every naming convention in the codebase:

### 1. PascalCase

```
KubernetesMicroservice → KubernetesDeployment
```

**Used in**:
- Proto message types (`message KubernetesDeployment`)
- Go struct types (`type KubernetesDeployment struct`)
- Enum values (`CatalogKind_KubernetesDeployment`)

**Critical for**:
- Proto definitions
- Go type declarations
- Registry entries

### 2. camelCase

```
kubernetesMicroservice → kubernetesDeployment
```

**Used in**:
- Go variable names (`var deployment *kubernetesDeployment`)
- JSON field names (in some contexts)
- JavaScript/TypeScript identifiers

**Critical for**:
- Variable declarations
- Function parameters
- Method receivers (when camelCase convention used)

### 3. UPPER_SNAKE_CASE

```
KUBERNETES_MICROSERVICE → KUBERNETES_DEPLOYMENT
```

**Used in**:
- Environment variable names
- Go constants (`const KUBERNETES_DEPLOYMENT`)
- Proto enum zero values

**Critical for**:
- Configuration keys
- Constant definitions
- Environment variables

### 4. snake_case

```
kubernetes_microservice → kubernetes_deployment
```

**Used in**:
- Proto field names
- Database column names
- Internal references

**Critical for**:
- Proto field declarations
- Some naming conventions in config files

### 5. kebab-case

```
kubernetes-microservice → kubernetes-deployment
```

**Used in**:
- CLI flags (`--kubernetes-deployment`)
- URLs and slugs
- Some file names

**Critical for**:
- Command-line interfaces
- Documentation URLs
- Hyphenated identifiers

### 6. Space Separated (Quoted)

```
"kubernetes microservice" → "kubernetes deployment"
```

**Used in**:
- Documentation prose
- User-facing strings
- Comments and descriptions
- Log messages

**Critical for**:
- Human-readable text
- Documentation
- Help text

### 7. lowercase (No Delimiters)

```
kubernetesmicroservice → kubernetesdeployment
```

**Used in**:
- Directory names
- Go package names
- Proto package paths
- Import paths

**Critical for**:
- File system operations
- Package declarations
- Import statements

### Pattern Ordering

Patterns are applied in order of specificity (most specific first):

1. PascalCase (most unique)
2. camelCase
3. UPPER_SNAKE_CASE
4. snake_case
5. kebab-case
6. Space separated
7. lowercase (least specific, catches any remaining)

This ordering prevents incorrect replacements (e.g., lowercase pattern wouldn't incorrectly match inside PascalCase identifiers).

## File Operations

### What Gets Copied

```
catalog/{provider}/{old_folder}/
                                                    ↓
catalog/{provider}/{new_folder}/
```

**Everything in the kind directory**:
- `README.md`, `catalog.md`, `logo.svg`, `GUIDE.md` (kind root docs)
- `<version>/` (e.g., `v1alpha1/` -- the versioned contract)
  - `*.proto` (all four contract protos)
  - `*.pb.go` (generated stubs, will be regenerated)
  - `spec_test.go`, `BUILD.bazel`, `reference.md`
- `presets/` (preset manifests + .md sidecars)
- `e2e/` (test manifest, variants, scenarios)
- `iac/` (IaC modules)
  - `pulumi/` (Pulumi implementation)
  - `tf/` (Terraform implementation)

### What Gets Replaced In

**New Kind Directory** (all files)

Every file in the new kind directory gets processed for all 7 naming patterns.

### What Gets Updated

**catalog_kind.proto**

The registry entry gets updated:

```protobuf
// Before
KubernetesMicroservice = 810 [(kind_meta) = {
  provider: kubernetes
  version: "v1alpha1"
  id_prefix: "k8sms"
  is_service_kind: true
  kubernetes_meta: {
    category: workload
    namespace_prefix: "service"
  }
}];

// After
KubernetesDeployment = 810 [(kind_meta) = {
  provider: kubernetes
  version: "v1alpha1"
  id_prefix: "k8sdpl"              // Updated if new prefix provided
  is_service_kind: true              // Preserved
  kubernetes_meta: {
    category: workload               // Preserved
    namespace_prefix: "service"      // Preserved
  }
}];
```

**What's preserved**:
- Enum value (810)
- Provider (kubernetes)
- Version ("v1alpha1")
- All flags (`is_service_kind`)
- All metadata (`kubernetes_meta`)

**What's updated**:
- Enum name (KubernetesMicroservice → KubernetesDeployment)
- ID prefix (only if explicitly provided)

### What Gets Deleted

After successful copy and replacement:
- Old kind directory is deleted
- Old proto stubs are removed by `make protos`

## Build Pipeline

The rename isn't complete until all three build phases pass:

### Phase 1: make protos

**Purpose**: Regenerate proto stubs

**What it does**:
- Removes old `.pb.go` files
- Generates new `.pb.go` files with new names
- Updates all proto cross-references

**Why it can fail**:
- Proto syntax errors introduced by replacements
- Import path issues
- Message name conflicts

### Phase 2: go build ./catalog/...

**Purpose**: Compile entire codebase

**What it does**:
- Compiles all Go packages
- Verifies all imports resolve
- Type-checks all code

**Why it can fail**:
- Undefined identifiers (missed replacements)
- Import path errors
- Type mismatches

### Phase 3: go test -v ./catalog/...

**Purpose**: Run test suite

**What it does**:
- Executes all unit tests
- Validates behavior unchanged
- Checks validation rules

**Why it can fail**:
- Hard-coded strings in tests
- Test expectations changed
- Logic errors introduced

### Stop on First Failure

The pipeline stops immediately on first failure:

```
make protos → ✅ Success → Continue
go build ./catalog/...  → ❌ Failed  → STOP (show error)
go test -v ./catalog/...   → (not reached)
```

This provides fast feedback and prevents cascading errors.

## Safety and Recovery

### Git-Based Safety

**No backups are created**. The rename operation relies on git:

```bash
# Before rename
git status  # Should be clean

# After rename
git diff    # Review changes
git log     # See what changed

# If needed
git reset --hard HEAD  # Complete rollback
```

**Why no backups?**

1. **Git is the backup** - Full version control history
2. **Cleaner operations** - No temporary directories
3. **Standard practice** - Industry convention
4. **Faster execution** - No extra file operations

### Rollback Process

If rename fails or introduces issues:

```bash
# Complete rollback (before commit)
git reset --hard HEAD

# Review what was attempted
git diff HEAD~1  # If already committed

# Selective rollback
git checkout HEAD -- path/to/file
```

### Target Directory Handling

If target directory already exists:

```
Warning: Target directory exists: kubernetesdeployment/
Deleting existing target directory...
Proceeding with rename...
```

The script automatically deletes it. **No confirmation required**.

**Why?**
- Ensures clean slate
- Prevents merge conflicts
- Predictable behavior

**If this is problematic:**
- Check `git status` first
- Commit or stash any changes
- Then run rename

## Usage Walkthrough

### Step-by-Step Example

**Scenario**: Rename `KubernetesMicroservice` to `KubernetesDeployment`

#### Step 1: Preparation

```bash
# Ensure clean git state
git status

# Ensure tests pass
cd /path/to/planton
go test -v ./catalog/...
```

#### Step 2: Invoke Rule

```
@rename-catalog-kind
```

#### Step 3: Answer Questions

```
Old kind name (PascalCase): KubernetesMicroservice
New kind name (PascalCase): KubernetesDeployment
New ID prefix (current: k8sms, press Enter to keep): k8sdpl
```

#### Step 4: Review Summary

```
Rename Summary
==============
Old Kind: KubernetesMicroservice
New Kind: KubernetesDeployment
Old Folder: kubernetesmicroservice/
New Folder: kubernetesdeployment/
Provider: kubernetes/workload
Enum Value: 810 (preserved)
Old ID Prefix: k8sms
New ID Prefix: k8sdpl

Replacement Patterns (7 patterns):
  1. KubernetesMicroservice → KubernetesDeployment
  2. kubernetesMicroservice → kubernetesDeployment
  3. KUBERNETES_MICROSERVICE → KUBERNETES_DEPLOYMENT
  4. kubernetes_microservice → kubernetes_deployment
  5. kubernetes-microservice → kubernetes-deployment
  6. "kubernetes microservice" → "kubernetes deployment"
  7. kubernetesmicroservice → kubernetesdeployment

Proceed with rename? (yes/no):
```

#### Step 5: Confirm

```
yes
```

#### Step 6: Watch Execution

```
Validating...
  ✓ Old kind found in registry
  ✓ Old directory exists
  ✓ New kind name available

Copying directory...
  ✓ 247 files copied

Applying replacements...
  ✓ 1834 replacements in kind directory

Updating registry...
  ✓ Enum name updated
  ✓ ID prefix updated

Deleting old directory...
  ✓ Old kind removed

Running build pipeline...
  ✓ make protos (23s)
  ✓ go build ./catalog/... (34s)
  ✓ go test -v ./catalog/... (18s)

✅ Rename completed successfully!
```

#### Step 7: Review Results

```json
{
  "success": true,
  "old_kind": "KubernetesMicroservice",
  "new_kind": "KubernetesDeployment",
  "files_modified": 247,
  "replacements_made": 1901,
  "duration_seconds": 75.3
}
```

#### Step 8: Changelog Creation

```
All tests passed! Creating changelog...

@create-planton-changelog
```

The rule automatically invokes changelog creation.

#### Step 9: Review and Commit

```bash
# Review changes
git diff

# Review files changed
git status

# Commit
git add -A
git commit -m "refactor(kubernetes): rename KubernetesMicroservice to KubernetesDeployment

Removes abstraction - 'Microservice' doesn't accurately describe the
Kubernetes Deployment resource that gets created. This rename:
- Clarifies what Kubernetes resource is created
- Sets stage for KubernetesStatefulSet introduction
- Maintains naming consistency with KubernetesCronJob

Preserves:
- Enum value: 810
- All functionality
- Deployment behavior

Breaking change: User manifests must update kind field."

# Push
git push origin main
```

## Common Patterns

### Pattern 1: Abstraction Removal

**Before**: Name represents abstract concept
**After**: Name represents concrete implementation

Examples:
- `KubernetesMicroservice` → `KubernetesDeployment`
- `AwsDatabaseCluster` → `AwsAuroraCluster`
- `CloudStorage` → `AwsS3Bucket`

### Pattern 2: Consistency Establishment

**Before**: Inconsistent naming across similar resources
**After**: Consistent pattern

Examples:
- `PostgresKubernetes` → `KubernetesPostgres` (prefix consistency)
- `CertManagerKubernetes` → `CertManager` (suffix removal)

### Pattern 3: Specificity Addition

**Before**: Generic name
**After**: Specific implementation detail

Examples:
- `AwsEcsService` → `AwsEcsFargateService`
- `KubernetesDatabase` → `KubernetesPostgres`

### Pattern 4: Preparation for Expansion

**Before**: Name blocks future additions
**After**: Name allows for siblings

Examples:
- Before adding `KubernetesStatefulSet`, rename `KubernetesMicroservice` to `KubernetesDeployment`
- Before adding `AwsLambdaContainer`, rename generic `AwsLambda` to `AwsLambdaZip`

## Integration with Other Lifecycle Operations

### Rename + Audit

```bash
# Audit reveals naming issues
@audit-catalog-kind KubernetesMicroservice

# Result: Name is misleading abstraction (80% score, naming flagged)

# Rename to fix
@rename-catalog-kind
```

### Rename + Complete

**Order**: Rename first, then complete

```bash
# 1. Fix the name
@rename-catalog-kind

# 2. Fill gaps
@complete-catalog-kind KubernetesDeployment
```

**Why this order?**
- Rename establishes correct identity
- Complete fills missing artifacts under correct name
- Avoids wasted work on wrong name

### Rename + Fix

Rename can be part of a fix:

```bash
@fix-catalog-kind KubernetesMicroservice \
  --explain "Rename to KubernetesDeployment and fix validation"

# Fix rule may invoke rename internally
```

### Rename + Forge

Don't rename and forge simultaneously:

```bash
# ❌ Wrong: Rename then immediately forge
@rename-catalog-kind
@forge-catalog-kind KubernetesStatefulSet

# ✅ Right: Verify rename first
@rename-catalog-kind
[verify rename succeeded]
@forge-catalog-kind KubernetesStatefulSet
```

## Troubleshooting

### Error: Kind Not Found

```
Error: Kind KubernetesMicroservice not found in catalog_kind.proto
```

**Cause**: Typo in kind name or kind doesn't exist

**Solution**:
```bash
# List all Kubernetes kinds
grep "kubernetes" shared/catalogkind/catalog_kind.proto | grep "= [0-9]"

# Check exact spelling
```

### Error: New Kind Already Exists

```
Error: Kind KubernetesDeployment already exists in catalog_kind.proto
```

**Cause**: Target name is taken

**Solution**:
```bash
# Check if it's from a previous failed rename
ls catalog/kubernetes/workload/

# If it's a leftover, delete it
@delete-catalog-kind KubernetesDeployment --force

# Try rename again
```

### Error: Build Failed

```
Error: go build ./catalog/... failed
Exit code: 1
Output: undefined: kubernetesMicroservice.SomeType
```

**Cause**: Replacement missed some references

**Solution**:
```bash
# Rollback
git reset --hard HEAD

# Find missed references
grep -r "kubernetesMicroservice" .

# Two options:
# 1. Fix manually and commit
# 2. Improve replacement patterns in script
```

### Error: Tests Failed

```
Error: go test -v ./catalog/... failed
Exit code: 1
Output: Test "TestKubernetesMicroservice" expects old name
```

**Cause**: Hard-coded test expectations

**Solution**:
```bash
# Rollback
git reset --hard HEAD

# Update test expectations manually
# Then try rename again
```

### Warning: Target Exists

```
Warning: Target directory exists: kubernetesdeployment/
Deleting existing target directory...
```

**Not an error** - Script handles this automatically

**If problematic**:
```bash
# Check if target has uncommitted changes
git status kubernetesdeployment/

# If so, commit or stash first
git add kubernetesdeployment/
git commit -m "temp: save changes"

# Then rename
```

## Real-World Case Study: Workload Naming Refactoring

In November 2025, all 23 Kubernetes workload kinds were renamed:

**Before**: Suffix pattern
- `ArgocdKubernetes`, `PostgresKubernetes`, `RedisKubernetes`, etc.

**After**: Prefix pattern
- `KubernetesArgocd`, `KubernetesPostgres`, `KubernetesRedis`, etc.

**Scope**:
- 23 kinds renamed
- ~500 files modified
- ~15,000 lines changed
- All builds passed
- Zero behavioral changes

**Process**:
1. Created shell scripts for batch renaming
2. Applied 7 naming patterns systematically
3. Ran protos/build/test after each rename
4. Committed all changes together
5. Created comprehensive changelog

**Lessons Learned**:
- Systematic approach scales to large refactorings
- Build verification catches errors immediately
- Comprehensive patterns ensure completeness
- Zero-behavior-change is achievable with care

**Reference**: `_changelog/2025-11/2025-11-14-072635-kubernetes-workload-naming-consistency.md`

## Best Practices

### Before Rename

- ✅ Understand why the rename is needed
- ✅ Ensure tests pass on current name
- ✅ Commit or stash any uncommitted work
- ✅ Choose a clear, unambiguous new name
- ✅ Check if ID prefix should change

### During Rename

- ✅ Answer questions carefully
- ✅ Review summary thoroughly
- ✅ Watch build output for issues
- ✅ Don't interrupt the process

### After Rename

- ✅ Review git diff comprehensively
- ✅ Run tests manually as extra verification
- ✅ Create changelog documenting motivation
- ✅ Write informative commit message
- ✅ Update any external references (outside repo)

## Success Criteria

A rename is successful when:

- ✅ Kind directory renamed
- ✅ All 7 naming patterns applied
- ✅ Registry updated correctly
- ✅ Old directory deleted
- ✅ `make protos` passes
- ✅ `go build ./catalog/...` passes
- ✅ `go test -v ./catalog/...` passes
- ✅ Changelog created
- ✅ Changes committed

## Reference

- **Cursor Rule**: `rename-catalog-kind.mdc`
- **Python Script**: `_scripts/rename_kind.py`
- **Architecture**: `../../../architecture/catalog-kind.md`
- **Related Rules**: `audit`, `complete`, `fix`, `forge`, `delete`

---

**Remember**: Rename is about semantic truth. Change names when they don't accurately reflect what the kind does, but preserve all functionality and behavior.

