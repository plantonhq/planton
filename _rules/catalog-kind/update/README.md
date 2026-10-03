# Update: Enhance Kinds

## Overview

**Update** is the rule for enhancing existing kinds. It intelligently determines what needs updating based on your intent, audit results, or explicit instructions, then orchestrates the appropriate forge rules or targeted edits to bring the kind to ideal state.

## Why Update Exists

Kinds evolve over time:
- Requirements change
- Providers add new features
- Best practices evolve
- Initial implementation was incomplete
- Documentation becomes outdated
- Examples need refreshing

**Update makes evolution systematic and safe.**

## When to Use Update

### ✅ Use Update When

- **Filling gaps** - Audit shows kind is incomplete
- **Adding features** - New proto fields, new resources, new capabilities
- **Fixing issues** - Broken tests, build errors, outdated examples
- **Refreshing docs** - New research, updated best practices
- **Maintaining consistency** - Proto changed, need to sync Terraform/examples
- **Enhancing quality** - Improve code, add tests, expand documentation

### ❌ Don't Use Update When

- Kind doesn't exist → Use `@forge-catalog-kind`
- Want to remove kind → Use `@delete-catalog-kind`
- Just checking status → Use `@audit-catalog-kind`
- Kind is perfect → No update needed!

## The Six Update Scenarios

Update handles six distinct scenarios, each with its own workflow:

### 1. Fill Gaps (Audit-Driven)

**Trigger:** Audit shows <100% completion

```bash
@update-catalog-kind CloudflareD1Database --scenario fill-gaps
```

**Process:**
1. Reads audit report (or runs audit)
2. Identifies missing files/features
3. Runs specific forge rules to fill gaps
4. Validates results

**Example:** Audit shows missing Terraform module (70% complete)
- Runs rules 012-013 to create Terraform module
- Validates Terraform works
- Result: 95% complete

### 2. Proto Schema Changed

**Trigger:** You modified spec.proto, need to propagate changes

```bash
@update-catalog-kind GcpCertManagerCert --scenario proto-changed
```

**Process:**
1. Regenerates proto stubs: `make protos` (.pb.go files)
2. Validates kind tests: `go test ./catalog/<provider>/<kind>/...`
3. Updates Terraform variables.tf to match spec.proto
4. Updates presets and the e2e manifest to use new fields
5. Runs build validation: `go build ./catalog/...`
6. Runs full test validation: `go test -v ./catalog/...`

**Example:** Added `enable_ssl` field to spec.proto
- Runs `make protos` to regenerate stubs with new field
- Runs kind tests to validate buf.validate rules
- Adds `enable_ssl` variable to Terraform
- Updates examples to show SSL usage
- Runs `go build ./catalog/...` and `go test -v ./catalog/...` for full validation
- Result: Everything consistent with new schema

### 3. Refresh Documentation

**Trigger:** Documentation is outdated or incomplete

```bash
@update-catalog-kind KubernetesPostgres --scenario refresh-docs
```

**Process:**
1. Refreshes the kind guide (GUIDE.md)
2. Updates user-facing docs (README.md (kind root))
3. Refreshes examples with current patterns
4. Updates IaC documentation

**Example:** Provider released new features, docs mention old approach
- Researches current best practices
- Regenerates comprehensive docs
- Updates examples
- Result: Documentation reflects current state

### 4. Update IaC Implementation

**Trigger:** Need to modify Pulumi or Terraform deployment logic

```bash
@update-catalog-kind AwsRdsInstance --scenario update-iac --explain "add multi-AZ support"
```

**Process:**
1. Analyzes current implementation
2. Updates Pulumi module based on explanation
3. Updates Terraform module for feature parity
4. Runs build validation: `go build ./catalog/...`
5. Updates tests
6. Runs E2E tests
7. Runs full test validation: `go test -v ./catalog/...`

**Example:** Adding multi-region support
- Modifies Pulumi to create regional resources
- Runs `go build ./catalog/...` to validate compilation
- Mirrors changes in Terraform
- Updates tests for multi-region scenarios
- Runs `go test -v ./catalog/...` for full validation
- Result: Both IaC modules support multi-region

### 5. Fix Specific Issue

**Trigger:** Targeted fix needed

```bash
@update-catalog-kind GcpCertManagerCert --explain "presets use deprecated field names"
```

**Process:**
1. Analyzes issue description
2. Identifies affected files
3. Makes targeted fixes
4. Validates fixes

**Example:** Examples reference old field names
- Scans presets and the e2e manifest for deprecated fields
- Updates to current field names
- Validates examples against schema
- Result: Examples work correctly

### 6. Auto (Let AI Decide)

**Trigger:** Not sure which scenario applies

```bash
@update-catalog-kind CloudflareD1Database
```

**Process:**
1. Runs quick audit
2. Asks clarifying questions
3. Determines best scenario
4. Proceeds with that scenario

**Example:** User not sure what's needed
- Audit shows 70% complete (missing docs)
- AI suggests "fill-gaps" scenario
- Executes documentation generation
- Result: 95% complete

## Typical Workflows

### Workflow 1: Audit → Update → Audit

```bash
# 1. Check current state
@audit-catalog-kind CloudflareD1Database
# Result: 65% complete (missing Terraform, docs)

# 2. Fill gaps
@update-catalog-kind CloudflareD1Database --scenario fill-gaps
# Runs rules 012-013, 019, validation

# 3. Verify improvement
@audit-catalog-kind CloudflareD1Database
# Result: 98% complete
```

### Workflow 2: Edit Proto → Update

```bash
# 1. Manually edit spec.proto
# Added: bool enable_monitoring = 15;

# 2. Propagate changes
@update-catalog-kind GcpCloudSql --scenario proto-changed

# 3. Test changes
# Deploy with e2e manifest to verify
```

### Workflow 3: Provider Update → Refresh Docs

```bash
# 1. Provider released new features
# Your docs mention old approach

# 2. Refresh documentation
@update-catalog-kind AwsRdsInstance --scenario refresh-docs

# 3. Review generated docs
# Check GUIDE.md reflects current best practices
```

### Workflow 4: Feature Request → Update IaC

```bash
# 1. Need to add custom VPC support
@update-catalog-kind GcpGkeCluster --scenario update-iac --explain "add support for custom VPC with private IP ranges"

# 2. Review generated code
# Check both Pulumi and Terraform modules

# 3. Test deployment
# Use e2e manifest with custom VPC config
```

## Flags and Options

### Core Flags

| Flag | Purpose | Example |
|------|---------|---------|
| `--scenario` | Specify update type | `--scenario fill-gaps` |
| `--explain` | Describe what to update | `--explain "add SSL support"` |
| `--dry-run` | Preview changes | `--dry-run` |
| `--backup` | Create backup first | `--backup` |
| `--resume-from` | Resume from rule number | `--resume-from 012` |

### Scenario Values

- `fill-gaps` - Fill missing items from audit
- `proto-changed` - Propagate proto schema changes
- `refresh-docs` - Update documentation
- `update-iac` - Modify deployment logic
- `auto` - Let AI determine (default)

### Flag Combinations

```bash
# Preview gap-filling
@update-catalog-kind CloudflareD1Database --scenario fill-gaps --dry-run

# Update IaC with backup
@update-catalog-kind GcpCertManagerCert --scenario update-iac --explain "add DNS validation" --backup

# Auto-determine with explanation
@update-catalog-kind KubernetesPostgres --explain "examples need updating to show volume configuration"
```

## Safety Features

### 1. Dry-Run Mode

Preview changes before applying:

```bash
@update-catalog-kind CloudflareD1Database --scenario fill-gaps --dry-run
```

**Output:**
```
📋 Update Plan for CloudflareD1Database

Current State: 65% complete
Missing Items:
  ❌ iac/tf/ (Terraform module)
  ❌ GUIDE.md (kind guide)
  ⚠️  presets/ (incomplete)

Planned Actions:
  1. Run rules 012-013 → Create Terraform module
  2. Run rule 019 → Write GUIDE.md
  3. Enhance presets/ → Add 2 more presets

Estimated Duration: 10-15 minutes
Estimated Files Modified: 15

Run without --dry-run to apply changes.
```

### 2. Backup Before Update

Create timestamped backup:

```bash
@update-catalog-kind GcpCertManagerCert --scenario proto-changed --backup
```

**Creates:**
```
catalog/gcp/gcpcertmanagercert/v1alpha1/.backup-2025-11-13-143022/
├── spec.proto
├── api.proto
├── iac/
└── ... (all files before update)
```

**Restore if needed:**
```bash
cp -r /tmp/planton-backups/<kind>-backup-2025-11-13-143022/* catalog/<provider>/<kind>/
```

### 3. Validation Checkpoints

Update validates after major changes with specific commands:

| Checkpoint | Command | Validates | Fails If |
|------------|---------|-----------|----------|
| After proto changes | `make protos` | Proto compiles, stubs generated | Import errors, syntax errors |
| Kind tests | `go test ./catalog/...` | buf.validate rules work | Any spec_test.go failure |
| After Go/Pulumi changes | `go build ./catalog/...` | Complete build succeeds | Compilation errors |
| After doc updates | Validation | Examples work | Invalid YAML, wrong fields |
| Final validation | `go test -v ./catalog/...` | Full test suite passes | Any test failure |

**Build and Test Execution:**
Update always runs these commands in sequence:
```bash
# 1. If proto changed: Regenerate stubs
make protos

# 2. Always: Validate kind tests (validates buf.validate rules)
go test ./catalog/<provider>/<kind>/...

# 3. If Go/Pulumi code changed: Verify complete build
go build ./catalog/...

# 4. Always: Verify all tests pass
go test -v ./catalog/...
```
This ensures spec_test.go correctly validates all validation rules in spec.proto and the complete build succeeds.

### 4. Automatic Retry

Each operation retries up to 3 times:
- Build errors → auto-fix and retry
- Test failures → analyze and retry
- Syntax errors → fix and retry

### 5. Conflict Detection

If update would overwrite custom changes:
```
⚠️  Conflict detected in iac/pulumi/module/main.go

Your custom changes:
  - Added custom validation logic (line 45-60)

Update wants to:
  - Regenerate main.go based on new spec

Options:
  1. Skip this file (keep your changes)
  2. Overwrite (lose your changes)
  3. Show diff and merge manually
  4. Cancel update

Choice: _
```

## Progress Tracking

Update provides detailed progress:

```
🔄 Updating CloudflareD1Database

Scenario: fill-gaps
Current: 65% → Target: 95%+

Phase 1: Create Terraform Module
[1/7] ✅ Generated variables.tf (matches spec.proto)
[2/7] ✅ Generated provider.tf
[3/7] ✅ Generated locals.tf
[4/7] ✅ Generated main.tf (creates cluster + database)
[5/7] ✅ Generated outputs.tf (maps to outputs.proto)
[6/7] ✅ Generated README.md
[7/7] ✅ Passed terraform validate

Phase 2: Generate Documentation
[8/8] ✅ Wrote GUIDE.md (kind guide)

Phase 3: Validation
[10/10] ✅ Build passed (go build ./catalog/...)
[11/11] ✅ Tests passed (go test -v ./catalog/...)

✅ Update complete!

Summary:
  Before: 65% complete
  After: 98% complete
  Improvement: +33%
  Files created: 12
  Files modified: 3
  Duration: 12 minutes

Next Steps:
  1. Review generated files
  2. Run: @audit-catalog-kind CloudflareD1Database
  3. Test with: e2e/manifest.yaml
  4. Commit changes
```

## Error Handling

### Common Errors

**Error: Kind not found**
```
❌ CloudflareD1Database not found in catalog_kind.proto

Did you mean:
  - MongodbKubernetes
  - CloudflareD1Database (check spelling)

Or create new:
  @forge-catalog-kind CloudflareD1Database --provider cloudflare
```

**Error: Nothing to update**
```
✅ CloudflareD1Database is already 100% complete

Audit shows all items present:
  ✅ Proto files
  ✅ IaC modules
  ✅ Documentation
  ✅ Tests passing

If you need specific changes, use:
  --scenario update-iac --explain "..."
```

**Error: Proto regeneration failed**
```
❌ Failed to regenerate proto stubs

Error: buf: undefined message MyCustomType

Fix:
  1. Check spec.proto syntax
  2. Ensure all message types are defined
  3. Run: make protos
  4. Resume: @update-catalog-kind CloudflareD1Database --resume-from <rule-number>
```

### Recovery

If update fails:
1. Error message shows what succeeded
2. Suggestion for fix provided
3. Resume from failure point:
   ```bash
   @update-catalog-kind CloudflareD1Database --resume-from <rule-number>
   ```

## Integration Examples

### With Audit

```bash
# Daily workflow
@audit-catalog-kind MyKind    # Check status
@update-catalog-kind MyKind --scenario fill-gaps  # Fill gaps
@audit-catalog-kind MyKind    # Verify improvement
```

### With Forge

```bash
# Initial creation might be incomplete
@forge-catalog-kind NewKind --provider aws
# Result: 70% (documentation might be minimal)

# Fill remaining gaps
@update-catalog-kind NewKind --scenario fill-gaps
# Result: 98% (full documentation generated)
```

### Continuous Improvement

```bash
# Week 1: Create kind
@forge-catalog-kind MyKind --provider gcp

# Week 2: Add features
@update-catalog-kind MyKind --scenario proto-changed
# (after adding fields to spec.proto)

# Week 3: Provider releases new features
@update-catalog-kind MyKind --scenario refresh-docs

# Week 4: Enhance deployment
@update-catalog-kind MyKind --scenario update-iac --explain "add auto-scaling"
```

## Best Practices

### Before Update

1. ✅ **Run audit** - Know current state
2. ✅ **Commit changes** - Clean git state
3. ✅ **Use dry-run** - Preview changes
4. ✅ **Backup if unsure** - Use --backup flag

### During Update

1. ✅ **Monitor progress** - Watch for errors
2. ✅ **Be specific** - Clear --explain descriptions
3. ✅ **Trust validation** - Errors show immediately
4. ✅ **Don't interrupt** - Let it complete

### After Update

1. ✅ **Review diffs** - Check what changed
2. ✅ **Run audit** - Verify improvements
3. ✅ **Test locally** - Deploy with e2e manifest
4. ✅ **Commit meaningfully** - Good commit message

## Tips

### Choosing the Right Scenario

| Situation | Scenario |
|-----------|----------|
| Audit shows gaps | `fill-gaps` |
| You edited proto | `proto-changed` |
| Docs outdated | `refresh-docs` |
| Need new feature | `update-iac` |
| Specific fix | Use `--explain` |
| Not sure | Let it auto-determine |

### Writing Good --explain Descriptions

**Bad:**
```
--explain "fix stuff"
--explain "update"
--explain "make it better"
```

**Good:**
```
--explain "add support for custom VPC with private IP ranges"
--explain "presets use deprecated 'database_name' field, should use 'db_identifier'"
--explain "Pulumi module doesn't set tags on resources, add standard tags"
```

### Dealing with Custom Changes

If you've customized generated code:

**Option 1: Keep customizations**
- Choose "skip" when conflicts detected
- Manually merge if needed

**Option 2: Regenerate**
- Document your customizations first
- Let update regenerate
- Reapply customizations after

**Option 3: Fork patterns**
- Move custom logic to separate files
- Update won't touch non-standard files

## Troubleshooting

### Update Takes Too Long

**If stuck:**
1. Check progress messages (shows current step)
2. Wait for timeout (3 retries per rule)
3. Cancel and investigate: Ctrl+C
4. Check logs for specific error

### Build Fails After Update

**Debug:**
```bash
cd catalog/<provider>/<kind>/v1
make protos    # Regenerate stubs
go build       # Check Go errors
go test -v ./catalog/...      # Run tests
```

### Examples Don't Work After Update

**Fix:**
1. Check presets and the e2e manifest use current field names
2. Validate examples against schema:
   ```bash
   planton validate --manifest examples.yaml
   ```
3. Update examples manually if needed

### Audit Score Didn't Improve

**Check:**
1. Update completed successfully?
2. Which gaps were addressed?
3. Run audit with verbose:
   ```bash
   @audit-catalog-kind MyKind --verbose
   ```

## Success Metrics

Good update outcomes:

- ✅ Audit score improved
- ✅ All tests pass
- ✅ Build succeeds
- ✅ No regressions
- ✅ Documentation current
- ✅ Examples work
- ✅ Kind deployable

## Related Commands

- `@forge-catalog-kind` - Create new kind
- `@audit-catalog-kind` - Check completion status
- `@complete-catalog-kind` - Auto-improve to 95%+ (audit + update + audit)
- `@fix-catalog-kind` - Targeted fixes with cascading updates
- `@delete-catalog-kind` - Remove kind

## Questions?

- Check ideal state: `architecture/catalog-kind.md`
- Review forge rules: `_rules/catalog-kind/forge/flow/`
- Run audit to see examples of complete kinds

---

**Ready to update?** Run `@update-catalog-kind <KindName>` to get started!

