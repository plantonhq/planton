# Delete: Safe Component Removal

## Overview

**Delete** is the rule for safely removing deployment components from the OpenMCF codebase. It provides multiple safety features to prevent accidental deletion and ensure clean removal of all artifacts.

## Why Delete Exists

Components have lifecycles:
- Providers discontinue services
- Components become obsolete
- Better alternatives emerge
- Consolidation is needed
- Test components need cleanup

**Delete makes removal systematic, safe, and complete.**

## Philosophy: Safety First

Deletion is **irreversible** without backups. Delete prioritizes safety:

1. **Preview First** - Dry-run shows what would be deleted
2. **Backup by Default** - Creates timestamped backup
3. **Check References** - Warns about dependencies
4. **Explicit Confirmation** - Must type component name
5. **Detailed Reporting** - Shows exactly what was removed

**Principle:** Make deletion hard to do accidentally, easy to undo.

## When to Use Delete

### ✅ Use Delete When

- **Component is obsolete** - No longer supported, deprecated
- **Provider discontinued** - Cloud provider shut down service
- **Consolidation** - Merged into another component
- **Test cleanup** - Removing POC or experimental components
- **Duplication** - Component serves same purpose as another

### ❌ Don't Use Delete When

- Component needs updates → Use `@update-openmcf-component`
- Component has bugs → Use `@update-openmcf-component`
- Documentation needs fixing → Use `@update-openmcf-component`
- Unsure if needed → Run `@audit-openmcf-component` first
- Component doesn't exist → Nothing to delete

## The Safe Deletion Workflow

### Recommended Process

```bash
# Step 1: Understand current state
@audit-openmcf-component ObsoleteComponent

# Step 2: Preview deletion (dry-run)
@delete-openmcf-component ObsoleteComponent --dry-run

# Step 3: Review what would be deleted

# Step 4: Delete with backup
@delete-openmcf-component ObsoleteComponent --backup

# Step 5: Confirm deletion
# (Type: DELETE ObsoleteComponent)

# Step 6: Verify no issues
go build ./apis/.../v1/...
go test -v ./apis/.../v1/

# Step 7: Commit changes
git add -A
git commit -m "Remove ObsoleteComponent (reason: ...)"
```

## What Gets Deleted

Delete removes **everything** related to a component:

### 1. Component Folder (All Files)

```
apis/org/openmcf/provider/<provider>/<component>/v1/
├── api.proto                    ❌ Deleted
├── spec.proto                   ❌ Deleted
├── stack_input.proto            ❌ Deleted
├── stack_outputs.proto          ❌ Deleted
├── *.pb.go                      ❌ Deleted
├── spec_test.go                 ❌ Deleted
├── README.md                    ❌ Deleted
├── examples.md                  ❌ Deleted
├── docs/
│   └── README.md                ❌ Deleted
└── iac/
    ├── hack/
    │   └── manifest.yaml        ❌ Deleted
    ├── pulumi/
    │   ├── main.go              ❌ Deleted
    │   ├── Pulumi.yaml          ❌ Deleted
    │   ├── Makefile             ❌ Deleted
    │   ├── README.md            ❌ Deleted
    │   ├── overview.md          ❌ Deleted
    │   ├── debug.sh             ❌ Deleted
    │   └── module/
    │       ├── main.go          ❌ Deleted
    │       ├── locals.go        ❌ Deleted
    │       └── outputs.go       ❌ Deleted
    └── tf/
        ├── variables.tf         ❌ Deleted
        ├── provider.tf          ❌ Deleted
        ├── locals.tf            ❌ Deleted
        ├── main.tf              ❌ Deleted
        ├── outputs.tf           ❌ Deleted
        └── README.md            ❌ Deleted
```

**Result:** Entire component folder removed (0 files remain).

### 2. Registry Entry

```protobuf
// Before deletion
enum CloudResourceKind {
  ...
  MongodbAtlas = 51 [(kind_meta) = {    ❌ Deleted
    provider: atlas                      ❌ Deleted
    version: v1                          ❌ Deleted
    id_prefix: "mdbatl"                  ❌ Deleted
  }];                                    ❌ Deleted
  SnowflakeDatabase = 52 [(kind_meta) = {
    provider: snowflake
    version: v1
    id_prefix: "snowdb"
  }];
  ...
}

// After deletion
enum CloudResourceKind {
  ...
  // MongodbAtlas removed
  SnowflakeDatabase = 52 [(kind_meta) = {
    provider: snowflake
    version: v1
    id_prefix: "snowdb"
  }];
  ...
}
```

**Result:** Enum entry completely removed from `cloud_resource_kind.proto`.

### 3. Generated Proto Stubs

After deletion, running `make protos` removes stale .pb.go files for deleted component.

## Flags and Options

### --dry-run (Always Use First)

**Preview without deleting:**

```bash
@delete-openmcf-component MongodbAtlas --dry-run
```

**Output:**
```
🔍 Dry-Run: MongodbAtlas Deletion

Component Info:
  Provider: atlas
  Enum Value: 51
  ID Prefix: mdbatl
  Path: apis/org/openmcf/provider/atlas/mongodbatlas/v1/

Would Delete:
  📁 Component folder
     23 files, 450 KB total
     
  📝 Registry entry
     cloud_resource_kind.proto: MongodbAtlas = 51

Would Check:
  🔎 References in other files
     - Go imports
     - Proto imports
     - Documentation
     - Examples
     
Would Create (if --backup used):
  💾 Backup folder
     mongodbatlas-backup-YYYY-MM-DD-HHMMSS/

Summary:
  Files to delete: 23
  Estimated time: 5-10 seconds
  Reversible: Yes (with backup)

No files will be modified (dry-run mode)

To proceed:
  @delete-openmcf-component MongodbAtlas --backup
```

### --backup (Strongly Recommended)

**Create backup before deleting:**

```bash
@delete-openmcf-component MongodbAtlas --backup
```

**Creates:**
```
apis/org/openmcf/provider/atlas/
├── mongodbatlas/                         # Original (will be deleted)
└── mongodbatlas-backup-2025-11-13-143022/  # Backup (preserved)
    ├── v1/
    │   ├── ... (all files)
    └── enum_entry.txt                    # Saved enum entry
```

**Restore if needed:**
```bash
# Restore component
cp -r mongodbatlas-backup-2025-11-13-143022/v1 mongodbatlas/

# Restore enum entry
cat mongodbatlas-backup-2025-11-13-143022/enum_entry.txt
# Manually add to cloud_resource_kind.proto

# Regenerate stubs
make protos
```

### --force (Use with Caution)

**Skip confirmation prompt:**

```bash
@delete-openmcf-component TestComponent --force --backup
```

**When to use:**
- Scripting/automation
- Absolutely certain
- Already verified safe

**When NOT to use:**
- First time deleting
- Unsure about references
- Component might be needed

### --skip-references

**Skip reference checking (faster but risky):**

```bash
@delete-openmcf-component TestComponent --skip-references --force
```

**Warning:** May break builds if component is used elsewhere.

## Reference Checking

Delete automatically searches for references:

### What It Searches

**Go Code:**
```go
import "org/openmcf/provider/atlas/mongodbatlas/v1"  // ⚠️ Reference found
```

**Proto Files:**
```protobuf
import "org/openmcf/provider/atlas/mongodbatlas/v1/api.proto";  // ⚠️ Reference found
```

**Documentation:**
```markdown
See [MongodbAtlas](./mongodbatlas/) for SaaS examples.  // ⚠️ Reference found
```

**Configuration:**
```yaml
kind: MongodbAtlas  // ⚠️ Reference found
```

### If References Found

```
⚠️  Warning: MongodbAtlas is referenced in 3 files

Critical References (will break build):
  1. apis/org/openmcf/provider/atlas/backup/v1/spec.proto:15
     import "org/openmcf/provider/atlas/mongodbatlas/v1/api.proto";
     → Must remove or update import
     
  2. docs/examples/database-comparison.md:45
     See MongodbAtlas for managed NoSQL
     → Update documentation

Non-Critical References:
  3. docs/changelog/2024-03-15.md:12
     Added MongodbAtlas support
     → Historical reference, can leave as-is

Recommendations:
  1. Fix critical references before deletion
  2. Update or remove import in backup component
  3. Update database comparison documentation

Options:
  - Fix references first (recommended)
  - Use --force to delete anyway (may break build)
  - Cancel deletion

Proceed? (y/n): _
```

## Confirmation Process

Delete requires explicit confirmation:

```
🗑️  Ready to Delete: MongodbAtlas

Summary:
  Provider: atlas
  Enum: MongodbAtlas = 51
  Files: 23 files (450 KB)
  Backup: Yes (mongodbatlas-backup-2025-11-13-143022)
  References: 3 found (2 critical)

⚠️  This action is IRREVERSIBLE without backup!

To confirm deletion, type the component name exactly:
DELETE MongodbAtlas

Type here: _
```

**Must type exactly:** `DELETE MongodbAtlas`

**Typos rejected:**
- `delete MongodbAtlas` ❌
- `DELETE mongodbatlas` ❌
- `MongodbAtlas` ❌
- `DELETE` ❌

**Only accepts:** `DELETE MongodbAtlas` ✅

## Deletion Report

After successful deletion:

```
✅ Deletion Complete: MongodbAtlas

What Was Deleted:
  ✅ Component folder
     Path: apis/org/openmcf/provider/atlas/mongodbatlas/v1/
     Files: 23 deleted
     Size: 450 KB freed
     
  ✅ Registry entry
     Removed: MongodbAtlas = 51
     From: cloud_resource_kind.proto
     
  ✅ Proto stubs regenerated
     Command: make protos
     Status: Success

Backup Created:
  💾 Location: mongodbatlas-backup-2025-11-13-143022/
  📦 Contents: 23 files + enum entry
  ⏰ Created: 2025-11-13 14:30:22
  
  To restore:
    cp -r mongodbatlas-backup-2025-11-13-143022/v1 mongodbatlas/

References Found:
  ⚠️  2 critical references require updates
  ℹ️  1 non-critical reference (historical)
  
  Details:
    1. backup/v1/spec.proto - Remove import
    2. docs/database-comparison.md - Update text

Build Status:
  ⚠️  Not verified (references may cause build errors)
  
  Recommended:
    go build ./apis/.../v1/...  # Check for import errors
    go test -v ./apis/.../v1/   # Check for test failures

Next Steps:
  1. Fix critical references (2 files)
  2. Run: go build ./apis/.../v1/... && go test -v ./apis/.../v1/
  3. Commit changes:
     git add -A
     git commit -m "Remove MongodbAtlas component"

Duration: 8 seconds
Status: ✅ Complete
```

## Common Scenarios

### Scenario 1: Clean Test Component Removal

```bash
# Test component with no references
@delete-openmcf-component TestCloudResourceGeneric --force --backup

# Output:
# ✅ No references found
# ✅ Deleted successfully
# ✅ Build still passes

go build ./apis/.../v1/... && go test -v ./apis/.../v1/  # ✅ All pass
```

### Scenario 2: Remove Obsolete Component

```bash
# Old implementation being replaced
@delete-openmcf-component OldPostgresKubernetes --dry-run

# Review references (find 5 references)
# Update references to new component
# ...edit files...

# Delete after fixing references
@delete-openmcf-component OldPostgresKubernetes --backup

# Verify
go build ./apis/.../v1/... && go test -v ./apis/.../v1/  # ✅ All pass
```

### Scenario 3: Provider Discontinuation

```bash
# Provider shut down service
@delete-openmcf-component DiscontinuedService --backup

# References found in docs (historical)
# Decision: Keep historical references

# Force delete with acknowledgment
# Type: DELETE DiscontinuedService

# Update docs to note service discontinued
vim docs/providers/old-provider.md
# Add: "Service discontinued as of 2025-11-13"

git add -A
git commit -m "Remove DiscontinuedService (provider shut down service)"
```

### Scenario 4: Consolidation

```bash
# Consolidated two similar components into one
# Deleting old component

# First: Ensure migration complete
grep -r "OldComponentName" .  # Find any usage

# Second: Delete old component
@delete-openmcf-component OldComponentName --backup

# Third: Update documentation
# Add migration guide explaining the consolidation

git add -A
git commit -m "Remove OldComponentName (consolidated into NewComponentName)"
```

## Error Handling

### Component Not Found

```
❌ Error: MongodbAtlas not found

Searched:
  ✓ cloud_resource_kind.proto - No enum entry
  ✓ File system - No folder at expected path

Possible reasons:
  1. Component name misspelled (check exact case)
  2. Component already deleted
  3. Component was never created

Similar components:
  - MongodbKubernetes ✓ Exists
  - ConfluentKafka ✓ Exists

Did you mean one of these?
```

### Permission Denied

```
❌ Error: Permission denied

Cannot delete: apis/org/openmcf/provider/atlas/mongodbatlas/v1/
Reason: Directory not writable

Fix:
  chmod -R u+w apis/org/openmcf/provider/atlas/mongodbatlas/
  
Then retry:
  @delete-openmcf-component MongodbAtlas --backup
```

### References Block Deletion

```
❌ Error: Critical references prevent safe deletion

Found 2 blocking references:
  1. backup/v1/spec.proto:15 - Import dependency
  2. other-component/v1/spec.proto:23 - Type dependency

Cannot proceed without --force (not recommended)

Recommended action:
  1. Fix references first
  2. Then delete safely

Or force delete (may break build):
  @delete-openmcf-component MongodbAtlas --force --backup
```

## Restoring Deleted Components

### From Backup (Recommended)

```bash
# List available backups
ls apis/org/openmcf/provider/<provider>/*-backup-*/

# Restore component folder
cp -r mongodbatlas-backup-2025-11-13-143022/v1 mongodbatlas/

# Restore enum entry
cat mongodbatlas-backup-2025-11-13-143022/enum_entry.txt
# Copy the enum entry text
vim apis/org/openmcf/shared/cloudresourcekind/cloud_resource_kind.proto
# Paste enum entry in correct location (numeric order)

# Regenerate proto stubs
make protos

# Verify restoration
go build ./apis/.../v1/... && go test -v ./apis/.../v1/
```

### From Git History

```bash
# Find when component was deleted
git log --all --oneline -- apis/org/openmcf/provider/atlas/mongodbatlas/

# Example output:
# abc1234 Remove MongodbAtlas component
# def5678 Update MongodbAtlas documentation
# ...

# Restore from commit before deletion
git checkout def5678 -- apis/org/openmcf/provider/atlas/mongodbatlas/

# Restore enum entry
git show def5678:apis/org/openmcf/shared/cloudresourcekind/cloud_resource_kind.proto | grep -A 5 "MongodbAtlas"
# Manually add to current cloud_resource_kind.proto

# Regenerate and verify
make protos && go build ./apis/.../v1/... && go test -v ./apis/.../v1/
```

## Best Practices

### Before Deleting

- [ ] Run audit to understand component
- [ ] Search for references (`grep -r "ComponentName" .`)
- [ ] Notify team of deletion intent
- [ ] Document reason for deletion
- [ ] Use --dry-run to preview
- [ ] Always use --backup flag

### During Deletion

- [ ] Read confirmation carefully
- [ ] Type component name exactly
- [ ] Watch for reference warnings
- [ ] Review deletion report
- [ ] Keep terminal output (for records)

### After Deletion

- [ ] Run `go build ./apis/.../v1/...` (check for errors)
- [ ] Run `go test -v ./apis/.../v1/` (check for failures)
- [ ] Fix any broken references
- [ ] Update related documentation
- [ ] Commit with descriptive message
- [ ] Keep backup for at least 30 days

## Safety Checklist

Before confirming deletion:

- [ ] Component is truly obsolete (not just needs updates)
- [ ] No active production usage
- [ ] Team is aware and approves
- [ ] Migration path exists (if applicable)
- [ ] Backup will be created (--backup flag)
- [ ] References documented or fixed
- [ ] Commit history shows component is safe to remove

## Troubleshooting

### "Cannot delete: Directory not empty"

```bash
# Force remove if needed
rm -rf apis/org/openmcf/provider/<provider>/<component>/

# Or fix permissions first
chmod -R u+w apis/org/openmcf/provider/<provider>/<component>/
```

### "Build fails after deletion"

```bash
# Find what broke
go build ./apis/.../v1/... 2>&1 | grep "error"

# Common issues:
# 1. Unremoved imports
grep -r "mongodbatlas" .

# 2. Type references
grep -r "MongodbAtlas" apis/

# Fix imports and references
# Then rebuild
go build ./apis/.../v1/... && go test -v ./apis/.../v1/
```

### "Need to restore deleted component"

```bash
# From backup
cp -r component-backup-*/v1 component/

# Restore enum entry (from backup or git)
# Edit cloud_resource_kind.proto

# Regenerate
make protos && go build ./apis/.../v1/... && go test -v ./apis/.../v1/
```

## Success Metrics

Successful deletion:

- ✅ Component folder removed (verified with `ls`)
- ✅ Enum entry removed (verified in proto file)
- ✅ Build succeeds (`go build ./apis/.../v1/...` passes)
- ✅ Tests pass (`go test -v ./apis/.../v1/` passes)
- ✅ Backup created (can be restored)
- ✅ References updated or documented
- ✅ Changes committed with clear message

## Related Commands

- `@audit-openmcf-component` - Check component status before deletion
- `@complete-openmcf-component` - Auto-improve to 95%+ (consider before deleting)
- `@fix-openmcf-component` - Fix specific issues (consider before deleting)
- `@forge-openmcf-component` - Create new component
- `@update-openmcf-component` - Update existing component

## Questions?

- Check ideal state: `architecture/deployment-component.md`
- Review delete rule: `_rules/deployment-component/delete/delete-openmcf-component.mdc`
- See examples: Run `--dry-run` on any component

---

**Remember:** Always `--dry-run` first, always `--backup` when deleting, and always verify with `go build ./apis/.../v1/... && go test -v ./apis/.../v1/`!
