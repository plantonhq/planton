# Delete: Safe Kind Removal

## Overview

**Delete** is the rule for safely removing kinds from the Planton codebase. It provides multiple safety features to prevent accidental deletion and ensure clean removal of all artifacts.

## Why Delete Exists

Kinds have lifecycles:
- Providers discontinue services
- Kinds become obsolete
- Better alternatives emerge
- Consolidation is needed
- Test kinds need cleanup

**Delete makes removal systematic, safe, and complete.**

## Philosophy: Safety First

Deletion is **irreversible** without backups. Delete prioritizes safety:

1. **Preview First** - Dry-run shows what would be deleted
2. **Backup by Default** - Creates timestamped backup
3. **Check References** - Warns about dependencies
4. **Explicit Confirmation** - Must type kind name
5. **Detailed Reporting** - Shows exactly what was removed

**Principle:** Make deletion hard to do accidentally, easy to undo.

## When to Use Delete

### ✅ Use Delete When

- **Kind is obsolete** - No longer supported, deprecated
- **Provider discontinued** - Cloud provider shut down service
- **Consolidation** - Merged into another kind
- **Test cleanup** - Removing POC or experimental kinds
- **Duplication** - Kind serves same purpose as another

### ❌ Don't Use Delete When

- Kind needs updates → Use `@update-catalog-kind`
- Kind has bugs → Use `@update-catalog-kind`
- Documentation needs fixing → Use `@update-catalog-kind`
- Unsure if needed → Run `@audit-catalog-kind` first
- Kind doesn't exist → Nothing to delete

## The Safe Deletion Workflow

### Recommended Process

```bash
# Step 1: Understand current state
@audit-catalog-kind ObsoleteKind

# Step 2: Preview deletion (dry-run)
@delete-catalog-kind ObsoleteKind --dry-run

# Step 3: Review what would be deleted

# Step 4: Delete with backup
@delete-catalog-kind ObsoleteKind --backup

# Step 5: Confirm deletion
# (Type: DELETE ObsoleteKind)

# Step 6: Verify no issues
go build ./catalog/...
go test -v ./catalog/...

# Step 7: Commit changes
git add -A
git commit -m "Remove ObsoleteKind (reason: ...)"
```

## What Gets Deleted

Delete removes **everything** related to a kind:

### 1. Kind Folder (All Files)

```
catalog/<provider>/<kind>/
├── README.md                    ❌ Deleted
├── catalog.md                   ❌ Deleted
├── logo.svg                     ❌ Deleted
├── GUIDE.md                     ❌ Deleted (if present)
├── <version>/                   (e.g., v1alpha1/)
│   ├── api.proto                ❌ Deleted
│   ├── spec.proto               ❌ Deleted
│   ├── input.proto              ❌ Deleted
│   ├── outputs.proto            ❌ Deleted
│   ├── *.pb.go                  ❌ Deleted
│   ├── spec_test.go             ❌ Deleted
│   ├── BUILD.bazel              ❌ Deleted
│   └── reference.md             ❌ Deleted
├── presets/
│   └── *.yaml + *.md            ❌ Deleted
├── e2e/
│   └── manifest.yaml (+ variants, scenarios/)  ❌ Deleted
└── iac/
    ├── pulumi/
    │   ├── main.go              ❌ Deleted
    │   ├── Pulumi.yaml          ❌ Deleted
    │   ├── README.md            ❌ Deleted
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

**Result:** Entire kind folder removed (0 files remain).

### 2. Registry Entry

```protobuf
// Before deletion
enum CatalogKind {
  ...
  CloudflareD1Database = 7005 [(kind_meta) = {    ❌ Deleted
    provider: cloudflare                            ❌ Deleted
    version: v1                                     ❌ Deleted
    id_prefix: "cfd1db"                             ❌ Deleted
  }];                                               ❌ Deleted
  CloudflareZeroTrustAccessApplication = 7006 [(kind_meta) = {
    provider: cloudflare
    version: v1
    id_prefix: "cfzta"
  }];
  ...
}

// After deletion
enum CatalogKind {
  ...
  // CloudflareD1Database removed
  CloudflareZeroTrustAccessApplication = 7006 [(kind_meta) = {
    provider: cloudflare
    version: v1
    id_prefix: "cfzta"
  }];
  ...
}
```

**Result:** Enum entry completely removed from `catalog_kind.proto`.

### 3. Generated Proto Stubs

After deletion, running `make protos` removes stale .pb.go files for deleted kind.

## Flags and Options

### --dry-run (Always Use First)

**Preview without deleting:**

```bash
@delete-catalog-kind CloudflareD1Database --dry-run
```

**Output:**
```
🔍 Dry-Run: CloudflareD1Database Deletion

Kind Info:
  Provider: cloudflare
  Enum Value: 7005
  ID Prefix: cfd1db
  Path: catalog/cloudflare/cloudflared1database/v1alpha1/

Would Delete:
  📁 Kind folder
     23 files, 450 KB total
     
  📝 Registry entry
     catalog_kind.proto: CloudflareD1Database = 7005

Would Check:
  🔎 References in other files
     - Go imports
     - Proto imports
     - Documentation
     - Examples
     
Would Create (if --backup used):
  💾 Backup folder
     cloudflared1database-backup-YYYY-MM-DD-HHMMSS/

Summary:
  Files to delete: 23
  Estimated time: 5-10 seconds
  Reversible: Yes (with backup)

No files will be modified (dry-run mode)

To proceed:
  @delete-catalog-kind CloudflareD1Database --backup
```

### --backup (Strongly Recommended)

**Create backup before deleting:**

```bash
@delete-catalog-kind CloudflareD1Database --backup
```

**Creates:**
```
catalog/cloudflare/
├── cloudflared1database/                         # Original (will be deleted)
└── cloudflared1database-backup-2025-11-13-143022/  # Backup (preserved)
    ├── v1/
    │   ├── ... (all files)
    └── enum_entry.txt                    # Saved enum entry
```

**Restore if needed:**
```bash
# Restore kind
cp -r cloudflared1database-backup-2025-11-13-143022/v1 cloudflared1database/

# Restore enum entry
cat cloudflared1database-backup-2025-11-13-143022/enum_entry.txt
# Manually add to catalog_kind.proto

# Regenerate stubs
make protos
```

### --force (Use with Caution)

**Skip confirmation prompt:**

```bash
@delete-catalog-kind TestKind --force --backup
```

**When to use:**
- Scripting/automation
- Absolutely certain
- Already verified safe

**When NOT to use:**
- First time deleting
- Unsure about references
- Kind might be needed

### --skip-references

**Skip reference checking (faster but risky):**

```bash
@delete-catalog-kind TestKind --skip-references --force
```

**Warning:** May break builds if kind is used elsewhere.

## Reference Checking

Delete automatically searches for references:

### What It Searches

**Go Code:**
```go
import "catalog/cloudflare/cloudflared1database/v1alpha1"  // ⚠️ Reference found
```

**Proto Files:**
```protobuf
import "catalog/cloudflare/cloudflared1database/v1alpha1/api.proto";  // ⚠️ Reference found
```

**Documentation:**
```markdown
See [CloudflareD1Database](./cloudflared1database/) for serverless database examples.  // ⚠️ Reference found
```

**Configuration:**
```yaml
kind: CloudflareD1Database  // ⚠️ Reference found
```

### If References Found

```
⚠️  Warning: CloudflareD1Database is referenced in 2 files

Critical References (will break build):
  1. catalog/cloudflare/backup/v1alpha1/spec.proto:15
     import "catalog/cloudflare/cloudflared1database/v1alpha1/api.proto";
     → Must remove or update import

Non-Critical References:
  2. _changelog/2025-11/2025-11-13-091500-added-cloudflared1database.md:12
     Added CloudflareD1Database support
     → Historical reference, can leave as-is

Recommendations:
  1. Fix critical references before deletion
  2. Update or remove import in backup kind

Options:
  - Fix references first (recommended)
  - Use --force to delete anyway (may break build)
  - Cancel deletion

Proceed? (y/n): _
```

## Confirmation Process

Delete requires explicit confirmation:

```
🗑️  Ready to Delete: CloudflareD1Database

Summary:
  Provider: cloudflare
  Enum: CloudflareD1Database = 7005
  Files: 23 files (450 KB)
  Backup: Yes (cloudflared1database-backup-2025-11-13-143022)
  References: 2 found (1 critical)

⚠️  This action is IRREVERSIBLE without backup!

To confirm deletion, type the kind name exactly:
DELETE CloudflareD1Database

Type here: _
```

**Must type exactly:** `DELETE CloudflareD1Database`

**Typos rejected:**
- `delete CloudflareD1Database` ❌
- `DELETE cloudflared1database` ❌
- `CloudflareD1Database` ❌
- `DELETE` ❌

**Only accepts:** `DELETE CloudflareD1Database` ✅

## Deletion Report

After successful deletion:

```
✅ Deletion Complete: CloudflareD1Database

What Was Deleted:
  ✅ Kind folder
     Path: catalog/cloudflare/cloudflared1database/v1alpha1/
     Files: 23 deleted
     Size: 450 KB freed
     
  ✅ Registry entry
     Removed: CloudflareD1Database = 7005
     From: catalog_kind.proto
     
  ✅ Proto stubs regenerated
     Command: make protos
     Status: Success

Backup Created:
  💾 Location: cloudflared1database-backup-2025-11-13-143022/
  📦 Contents: 23 files + enum entry
  ⏰ Created: 2025-11-13 14:30:22
  
  To restore:
    cp -r cloudflared1database-backup-2025-11-13-143022 catalog/cloudflare/cloudflared1database

References Found:
  ⚠️  1 critical reference requires updates
  ℹ️  1 non-critical reference (historical)
  
  Details:
    1. backup/v1alpha1/spec.proto - Remove import

Build Status:
  ⚠️  Not verified (references may cause build errors)
  
  Recommended:
    go build ./catalog/...  # Check for import errors
    go test -v ./catalog/...   # Check for test failures

Next Steps:
  1. Fix critical references (1 file)
  2. Run: go build ./catalog/... && go test -v ./catalog/...
  3. Commit changes:
     git add -A
     git commit -m "Remove CloudflareD1Database kind"

Duration: 8 seconds
Status: ✅ Complete
```

## Common Scenarios

### Scenario 1: Clean Test Kind Removal

```bash
# Test kind with no references
@delete-catalog-kind TestCatalogKindGeneric --force --backup

# Output:
# ✅ No references found
# ✅ Deleted successfully
# ✅ Build still passes

go build ./catalog/... && go test -v ./catalog/...  # ✅ All pass
```

### Scenario 2: Remove Obsolete Kind

```bash
# Old implementation being replaced
@delete-catalog-kind OldKubernetesPostgres --dry-run

# Review references (find 5 references)
# Update references to new kind
# ...edit files...

# Delete after fixing references
@delete-catalog-kind OldKubernetesPostgres --backup

# Verify
go build ./catalog/... && go test -v ./catalog/...  # ✅ All pass
```

### Scenario 3: Provider Discontinuation

```bash
# Provider shut down service
@delete-catalog-kind DiscontinuedService --backup

# References found in docs (historical)
# Decision: Keep historical references

# Force delete with acknowledgment
# Type: DELETE DiscontinuedService

git add -A
git commit -m "Remove DiscontinuedService (provider shut down service)"
```

### Scenario 4: Consolidation

```bash
# Consolidated two similar kinds into one
# Deleting old kind

# First: Ensure migration complete
grep -r "OldKindName" .  # Find any usage

# Second: Delete old kind
@delete-catalog-kind OldKindName --backup

# Third: Update documentation
# Add migration guide explaining the consolidation

git add -A
git commit -m "Remove OldKindName (consolidated into NewKindName)"
```

## Error Handling

### Kind Not Found

```
❌ Error: CloudflareD1Database not found

Searched:
  ✓ catalog_kind.proto - No enum entry
  ✓ File system - No folder at expected path

Possible reasons:
  1. Kind name misspelled (check exact case)
  2. Kind already deleted
  3. Kind was never created

Similar kinds:
  - CloudflareKvNamespace ✓ Exists
  - CloudflareR2Bucket ✓ Exists

Did you mean one of these?
```

### Permission Denied

```
❌ Error: Permission denied

Cannot delete: catalog/cloudflare/cloudflared1database/v1alpha1/
Reason: Directory not writable

Fix:
  chmod -R u+w catalog/cloudflare/cloudflared1database/
  
Then retry:
  @delete-catalog-kind CloudflareD1Database --backup
```

### References Block Deletion

```
❌ Error: Critical references prevent safe deletion

Found 2 blocking references:
  1. backup/v1alpha1/spec.proto:15 - Import dependency
  2. other-kind/v1alpha1/spec.proto:23 - Type dependency

Cannot proceed without --force (not recommended)

Recommended action:
  1. Fix references first
  2. Then delete safely

Or force delete (may break build):
  @delete-catalog-kind CloudflareD1Database --force --backup
```

## Restoring Deleted Kinds

### From Backup (Recommended)

```bash
# List available backups
ls catalog/<provider>/*-backup-*/

# Restore kind folder
cp -r cloudflared1database-backup-2025-11-13-143022/v1 cloudflared1database/

# Restore enum entry
cat cloudflared1database-backup-2025-11-13-143022/enum_entry.txt
# Copy the enum entry text
vim shared/catalogkind/catalog_kind.proto
# Paste enum entry in correct location (numeric order)

# Regenerate proto stubs
make protos

# Verify restoration
go build ./catalog/... && go test -v ./catalog/...
```

### From Git History

```bash
# Find when kind was deleted
git log --all --oneline -- catalog/cloudflare/cloudflared1database/

# Example output:
# abc1234 Remove CloudflareD1Database kind
# def5678 Update CloudflareD1Database documentation
# ...

# Restore from commit before deletion
git checkout def5678 -- catalog/cloudflare/cloudflared1database/

# Restore enum entry
git show def5678:shared/catalogkind/catalog_kind.proto | grep -A 5 "CloudflareD1Database"
# Manually add to current catalog_kind.proto

# Regenerate and verify
make protos && go build ./catalog/... && go test -v ./catalog/...
```

## Best Practices

### Before Deleting

- [ ] Run audit to understand kind
- [ ] Search for references (`grep -r "KindName" .`)
- [ ] Notify team of deletion intent
- [ ] Document reason for deletion
- [ ] Use --dry-run to preview
- [ ] Always use --backup flag

### During Deletion

- [ ] Read confirmation carefully
- [ ] Type kind name exactly
- [ ] Watch for reference warnings
- [ ] Review deletion report
- [ ] Keep terminal output (for records)

### After Deletion

- [ ] Run `go build ./catalog/...` (check for errors)
- [ ] Run `go test -v ./catalog/...` (check for failures)
- [ ] Fix any broken references
- [ ] Update related documentation
- [ ] Commit with descriptive message
- [ ] Keep backup for at least 30 days

## Safety Checklist

Before confirming deletion:

- [ ] Kind is truly obsolete (not just needs updates)
- [ ] No active production usage
- [ ] Team is aware and approves
- [ ] Migration path exists (if applicable)
- [ ] Backup will be created (--backup flag)
- [ ] References documented or fixed
- [ ] Commit history shows kind is safe to remove

## Troubleshooting

### "Cannot delete: Directory not empty"

```bash
# Force remove if needed
rm -rf catalog/<provider>/<kind>/

# Or fix permissions first
chmod -R u+w catalog/<provider>/<kind>/
```

### "Build fails after deletion"

```bash
# Find what broke
go build ./catalog/... 2>&1 | grep "error"

# Common issues:
# 1. Unremoved imports
grep -r "cloudflared1database" .

# 2. Type references
grep -r "CloudflareD1Database" catalog/

# Fix imports and references
# Then rebuild
go build ./catalog/... && go test -v ./catalog/...
```

### "Need to restore deleted kind"

```bash
# From backup
cp -r kind-backup-*/v1 kind/

# Restore enum entry (from backup or git)
# Edit catalog_kind.proto

# Regenerate
make protos && go build ./catalog/... && go test -v ./catalog/...
```

## Success Metrics

Successful deletion:

- ✅ Kind folder removed (verified with `ls`)
- ✅ Enum entry removed (verified in proto file)
- ✅ Build succeeds (`go build ./catalog/...` passes)
- ✅ Tests pass (`go test -v ./catalog/...` passes)
- ✅ Backup created (can be restored)
- ✅ References updated or documented
- ✅ Changes committed with clear message

## Related Commands

- `@audit-catalog-kind` - Check kind status before deletion
- `@complete-catalog-kind` - Auto-improve to 95%+ (consider before deleting)
- `@fix-catalog-kind` - Fix specific issues (consider before deleting)
- `@forge-catalog-kind` - Create new kind
- `@update-catalog-kind` - Update existing kind

## Questions?

- Check ideal state: `architecture/catalog-kind.md`
- Review delete rule: `_rules/catalog-kind/delete/delete-catalog-kind.mdc`
- See examples: Run `--dry-run` on any kind

---

**Remember:** Always `--dry-run` first, always `--backup` when deleting, and always verify with `go build ./catalog/... && go test -v ./catalog/...`!
