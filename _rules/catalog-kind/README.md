# Kind Lifecycle Management

## Overview

This directory contains the complete rule system for managing kinds in Planton. These rules handle the entire lifecycle from creation to deletion, ensuring kinds match the **ideal state** defined in `architecture/catalog-kind.md`.

## The Seven Lifecycle Operations

Planton provides seven operations for kinds:

| Operation | Purpose | When to Use |
|-----------|---------|-------------|
| **🔨 Forge** | Create new kinds | Kind doesn't exist |
| **🔍 Audit** | Assess completeness | Check status, find gaps |
| **🔄 Update** | Enhance existing kinds | Fill gaps, refresh docs, general improvements |
| **✨ Complete** | Auto-improve to target | One-command: audit + fill gaps + verify |
| **🔧 Fix** | Targeted fixes with propagation | Specific bugs, sync issues, consistency fixes |
| **✏️ Rename** | Systematically rename kinds | Name clarity, remove abstractions, consistency |
| **🗑️ Delete** | Remove kinds | Obsolete, deprecated, consolidating |

**Key Principles:** 
- Each operation is atomic, well-documented, and follows the ideal state standard
- **Source code is the ultimate source of truth** - documentation must match implementation
- Complete is a convenience wrapper (audit + update + audit)
- Fix ensures consistency across all artifacts (code, docs, examples, tests)

---

## Quick Decision Tree

```
Need to work with a kind?
│
├─ Does the kind exist?
│  │
│  ├─ NO → Use @forge-catalog-kind
│  │        Creates complete, production-ready kind
│  │        Expected result: 95-100% complete
│  │
│  └─ YES → What do you need to do?
│     │
│     ├─ Just checking status?
│     │  └─ Use @audit-catalog-kind
│     │     Shows completion %, identifies gaps
│     │     Generates timestamped report
│     │
│     ├─ Make it production-ready quickly?
│     │  └─ Use @complete-catalog-kind
│     │     Audits + fills all gaps + verifies
│     │     One command to 95%+ completion
│     │
│     ├─ Have specific bug/issue to fix?
│     │  └─ Use @fix-catalog-kind
│     │     Targeted fix with cascading updates
│     │     Ensures code, docs, examples, tests all match
│     │     Source code is truth, docs updated to match
│     │
│     ├─ Need general improvements?
│     │  └─ Use @update-catalog-kind
│     │     Fills gaps, refreshes docs, updates IaC
│     │     6 scenarios: fill-gaps, proto-changed, etc.
│     │
│     ├─ Need to rename kind?
│     │  └─ Use @rename-catalog-kind
│     │     Systematic rename across entire codebase
│     │     7 naming patterns, build verification
│     │     Name clarity, remove abstractions
│     │
│     └─ Want to remove?
│        └─ Use @delete-catalog-kind
│           Safe deletion with backup and confirmation
│           Checks for references, creates backup
```

---

## 1. Forge: Create New Kinds

### Purpose
Bootstrap complete, production-ready kinds from scratch.

### When to Use
- Creating support for a new cloud provider resource
- Adding a new SaaS platform integration
- Creating a new Kubernetes workload or addon

### What It Creates

✅ **Proto API Definitions** - All 4 contract protos (`api.proto`, `spec.proto`, `input.proto`, `outputs.proto`) with validations and tests
✅ **IaC Modules** - Both Pulumi and Terraform with feature parity
✅ **Documentation** - `README.md`, `catalog.md`, module READMEs
✅ **Supporting Files** - Presets, e2e test manifests, build configs
✅ **Registry Entry** - Registered in catalog_kind.proto
✅ **Validation** - Build and test validation passed

**Result:** 95-100% completion score

### Usage

```bash
@forge-catalog-kind <KindName> --provider <provider>
```

**Examples:**
```bash
@forge-catalog-kind CloudflareD1Database --provider cloudflare
@forge-catalog-kind GcpStorageBucket --provider gcp
@forge-catalog-kind KubernetesPostgres --provider kubernetes --category workload
```

### Learn More
- **README:** [`forge/README.md`](forge/README.md)
- **Rule:** [`forge/forge-catalog-kind.mdc`](forge/forge-catalog-kind.mdc)
- **Flow Rules:** [`forge/flow/`](forge/flow/)

---

## 2. Audit: Assess Kind Completeness

### Purpose
Evaluate kinds against the ideal state and generate actionable completion reports.

### When to Use
- After forge (verify kind was created completely)
- After update (confirm improvements)
- Before update (identify what needs fixing)
- Regular quality checks
- Pre-commit validation

### What It Checks

File-set presence and absence (folder structure, required files, forbidden residue) is machine-enforced by the `pkg/anatomy` conformance gate (CI lane `lint.kind-anatomy.yaml`) -- the audit points there instead of re-checking inventories. The audit's own judgment covers content quality:

1. Catalog Kind Registry
2. Anatomy Conformance (gate pointer)
3. Protobuf API Definitions
4. IaC Modules - Pulumi
5. IaC Modules - Terraform
6. Documentation - User-Facing (README.md, catalog.md, GUIDE.md when present)
7. Presets and E2E Coverage
8. Nice to Have Items

**Scoring:** Weighted (Critical 40%, Important 40%, Nice-to-Have 20%)

### Usage

```bash
@audit-catalog-kind <KindName>
```

**Examples:**
```bash
@audit-catalog-kind CloudflareD1Database
@audit-catalog-kind GcpCertManagerCert
```

### Report Output

- **Overall completion percentage** (0-100%)
- **Category-by-category breakdown**
- **Quick wins** (easy improvements)
- **Critical gaps** (blocking issues)
- **Prioritized recommendations**
- **Comparison to complete kinds**

The report is presented in the session -- kind directories carry no audit artifacts (the anatomy gate keeps the kind file set closed).

### Learn More
- **README:** [`audit/README.md`](audit/README.md)
- **Rule:** [`audit/audit-catalog-kind.mdc`](audit/audit-catalog-kind.mdc)

---

## 3. Update: Enhance Existing Kinds

### Purpose
Improve existing kinds by filling gaps, adding features, refreshing docs, or fixing issues.

### When to Use
- Filling gaps identified by audit
- Adding new fields to proto schema
- Refreshing outdated documentation
- Modifying IaC deployment logic
- Fixing specific issues

### Update Scenarios

| Scenario | Use When | Command |
|----------|----------|---------|
| **Fill Gaps** | Audit shows <100% | `--scenario fill-gaps` |
| **Proto Changed** | Modified spec.proto | `--scenario proto-changed` |
| **Refresh Docs** | Docs outdated | `--scenario refresh-docs` |
| **Update IaC** | Change deployment | `--scenario update-iac` |
| **Fix Issue** | Specific problem | `--explain "..."` |
| **Auto** | Not sure | Let AI decide |

### Usage

```bash
@update-catalog-kind <KindName> [--scenario <type>] [--explain "<description>"]
```

**Examples:**
```bash
# Fill gaps from audit
@update-catalog-kind CloudflareD1Database --scenario fill-gaps

# Propagate proto changes
@update-catalog-kind GcpCertManagerCert --scenario proto-changed

# Refresh documentation
@update-catalog-kind AwsRdsInstance --scenario refresh-docs

# Update IaC implementation
@update-catalog-kind GcpGkeCluster --scenario update-iac --explain "add multi-region support"

# Fix specific issue
@update-catalog-kind KubernetesPostgres --explain "presets use deprecated field names"
```

### Safety Features
- ✅ Dry-run mode (`--dry-run`)
- ✅ Backup creation (`--backup`)
- ✅ Validation checkpoints
- ✅ Automatic retry (up to 3 times)
- ✅ Conflict detection

### Learn More
- **README:** [`update/README.md`](update/README.md)
- **Rule:** [`update/update-catalog-kind.mdc`](update/update-catalog-kind.mdc)

---

## 4. Complete: Auto-Improve to Production-Ready

### Purpose
One-command workflow that audits a kind and automatically fills all gaps to reach target completion score (default 95%).

### When to Use
- Making kind production-ready quickly
- Batch improving multiple kinds
- Quality gates before releases
- Onboarding legacy kinds
- Following up after forge

### What It Does

**Three-Step Automated Workflow:**
1. **Audit** - Assess current state and identify all gaps
2. **Fill Gaps** - Automatically run update --fill-gaps
3. **Verify** - Re-audit to confirm improvement

**Result:** Before/after comparison showing improvement

### Usage

```bash
@complete-catalog-kind <KindName> [flags]
```

**Examples:**
```bash
# Basic usage (target: 95%)
@complete-catalog-kind CloudflareD1Database

# Preview without changes
@complete-catalog-kind CloudflareD1Database --dry-run

# Custom target score
@complete-catalog-kind KubernetesPostgres --target-score 100

# Batch processing
for kind in Comp1 Comp2 Comp3; do
  @complete-catalog-kind $kind
done
```

### What Gets Filled

Automatically creates missing items:
- ✅ Terraform module (if missing)
- ✅ User-facing docs (README.md, catalog.md -- if missing/incomplete)
- ✅ Presets with .md sidecars (if missing/incomplete)
- ✅ Module READMEs (if missing)
- ✅ Supporting files (e2e manifests)

**Note:** Only fills gaps, doesn't modify existing files

### Typical Results

| Starting Score | Target | Duration | Result |
|----------------|--------|----------|--------|
| 40-60% | 95% | 30-40 min | 95-98% |
| 60-80% | 95% | 15-25 min | 95-98% |
| 80-94% | 95% | 5-15 min | 95-100% |
| 95%+ | 95% | 30 sec | Already complete |

### Learn More
- **README:** [`complete/README.md`](complete/README.md)
- **Rule:** [`complete/complete-catalog-kind.mdc`](complete/complete-catalog-kind.mdc)

---

## 5. Fix: Targeted Fixes with Cascading Updates

### Purpose
Make targeted fixes to kinds and automatically propagate changes to all related artifacts (documentation, examples, tests, IaC modules) to ensure complete consistency.

### Core Philosophy
**Source code is the ultimate source of truth.** Documentation describes code, code doesn't describe documentation.

### When to Use
- Fixing specific bugs in proto schema, IaC modules, or validation logic
- Correcting incorrect behavior with documentation updates
- Synchronizing artifacts when they've drifted (examples out of date)
- Fixing test failures and validation logic
- Restoring feature parity between Pulumi and Terraform

### What It Does

**Six-Step Workflow:**
1. **Analyze** - Understand the fix needed and read current source code
2. **Fix Source Code** - Make changes to proto, IaC modules, tests
3. **Propagate to Docs** - Update all documentation to match new code
4. **Validate Consistency** - Run 5 consistency checks
5. **Execute Tests** - Kind tests, build, full suite
6. **Report** - Show what was fixed and what was propagated

**Five Consistency Checks:**
- Proto ↔ Terraform variables
- Proto ↔ Examples (examples must validate)
- Pulumi ↔ Terraform (feature parity)
- Validations ↔ Tests (every rule tested)
- Documentation ↔ Implementation (docs match reality)

### Usage

```bash
@fix-catalog-kind <KindName> --explain "<detailed fix description>"
```

**Examples:**
```bash
# Fix validation logic
@fix-catalog-kind GcpCertManagerCert --explain "primaryDomainName validation should allow wildcards like *.example.com"

# Fix IaC implementation
@fix-catalog-kind AwsRdsInstance --explain "Pulumi hardcodes backup_retention_period instead of using spec field"

# Fix documentation drift
@fix-catalog-kind KubernetesPostgres --explain "presets use deprecated 'database_name' field, should be 'db_identifier'"

# Fix test failures
@fix-catalog-kind CloudflareD1Database --explain "spec_test.go expects validation on location but spec.proto has no validation rule"
```

### What Gets Updated

**Source Code (if needed):**
- spec.proto (validation rules, fields)
- Pulumi module (deployment logic)
- Terraform module (to maintain parity)
- spec_test.go (validation tests)

**Documentation (always):**
- README.md and catalog.md (match current API and behavior)
- Presets (match current API)
- GUIDE.md (if judgment changed)
- IaC READMEs (if usage changed)

**Validation:**
- Kind tests: `go test ./catalog/<provider>/<kind>/...`
- Build: `go build ./catalog/<provider>/<kind>/...`
- Full suite: `go test -v ./catalog/<provider>/<kind>/...`
- Preset validation
- Consistency checks

### Typical Duration

- Documentation-only fix: 2-5 minutes
- Proto + docs fix: 5-10 minutes
- IaC + docs fix: 10-20 minutes
- Complex multi-artifact fix: 20-30 minutes

### Learn More
- **README:** [`fix/README.md`](fix/README.md)
- **Rule:** [`fix/fix-catalog-kind.mdc`](fix/fix-catalog-kind.mdc)

---

## 6. Rename: Systematically Rename Kinds

### Purpose
Rename kinds across the entire codebase with comprehensive find-replace patterns, registry updates, and build verification.

### When to Use
- Removing abstractions (e.g., `KubernetesMicroservice` → `KubernetesDeployment`)
- Improving name clarity and accuracy
- Establishing naming consistency
- Preparing for kind expansion
- When kind name doesn't reflect actual behavior

### Philosophy
**Rename is about semantic truth, not functionality changes.**

Kind rename updates names everywhere while preserving all functionality, enum values, and behavior. It ensures names accurately reflect what kinds do.

### The Seven Naming Patterns

Rename applies comprehensive replacement patterns:

1. **PascalCase** - `KubernetesMicroservice` → `KubernetesDeployment`
2. **camelCase** - `kubernetesMicroservice` → `kubernetesDeployment`
3. **UPPER_SNAKE_CASE** - `KUBERNETES_MICROSERVICE` → `KUBERNETES_DEPLOYMENT`
4. **snake_case** - `kubernetes_microservice` → `kubernetes_deployment`
5. **kebab-case** - `kubernetes-microservice` → `kubernetes-deployment`
6. **Space separated** - `"kubernetes microservice"` → `"kubernetes deployment"`
7. **lowercase** - `kubernetesmicroservice` → `kubernetesdeployment`

### Usage

```bash
@rename-catalog-kind
```

The rule will interactively ask for:
1. Old kind name (PascalCase)
2. New kind name (PascalCase)
3. New ID prefix (optional, press Enter to keep existing)

**Example:**
```bash
@rename-catalog-kind

Old kind name: KubernetesMicroservice
New kind name: KubernetesDeployment
New ID prefix (current: k8sms): k8sdpl
```

### What Gets Updated

✅ **Kind directory** - Copied to new name, old deleted
✅ **All code references** - 7 patterns applied to all files
✅ **Registry** - Enum name and optional ID prefix updated
✅ **Build artifacts** - Proto stubs regenerated

### What Gets Preserved

✅ **Enum value** - Registry number unchanged (e.g., 810)
✅ **Provider** - Provider field preserved
✅ **Version** - Version directory unchanged (e.g., v1alpha1)
✅ **Flags** - Special flags like `is_service_kind` preserved
✅ **Metadata** - All other metadata unchanged
✅ **Functionality** - Zero behavioral changes

### Build Pipeline

Rename isn't complete until all phases pass:
1. `make protos` - Regenerate proto stubs
2. `go build ./catalog/<provider>/<kind>/...` - Verify compilation
3. `go test -v ./catalog/<provider>/<kind>/...` - Validate behavior unchanged

**Stops on first failure** for fast feedback.

### Post-Rename

If all tests pass:
- ✅ Automatically invokes `@create-planton-changelog`
- ✅ Provides commit message template
- ✅ Shows next steps

### Safety

**Git-based safety** - No backups created, relies on git:
```bash
git status  # Check before
git diff    # Review after
git reset --hard HEAD  # Rollback if needed
```

**Target handling** - Automatically deletes target directory if it exists.

### Real-World Example

November 2025: All 23 Kubernetes workload kinds renamed from suffix to prefix pattern:
- `PostgresKubernetes` → `KubernetesPostgres`
- `RedisKubernetes` → `KubernetesRedis`
- ~500 files modified, all builds passed, zero behavioral changes

See: `_changelog/2025-11/2025-11-14-072635-kubernetes-workload-naming-consistency.md`

### Typical Duration

- Simple rename: 1-3 minutes
- Complex rename (large kind): 3-7 minutes
- Build pipeline: 30-90 seconds

### Learn More
- **README:** [`rename/README.md`](rename/README.md)
- **Rule:** [`rename/rename-catalog-kind.mdc`](rename/rename-catalog-kind.mdc)
- **Script:** [`rename/_scripts/rename_kind.py`](rename/_scripts/rename_kind.py)

---

## 7. Delete: Remove Kinds Safely

### Purpose
Completely remove kinds with safety features to prevent accidents.

### When to Use
- Kind is obsolete or deprecated
- Provider discontinued service
- Consolidating similar kinds
- Cleaning up test kinds

### Safety Features

🔍 **Dry-Run Mode** - Preview what would be deleted
💾 **Automatic Backup** - Creates timestamped backup
🔎 **Reference Check** - Warns if kind is referenced
✋ **Confirmation Required** - Must type kind name
📋 **Detailed Report** - Shows exactly what was deleted

### Usage

```bash
@delete-catalog-kind <KindName> [flags]
```

**Recommended Workflow:**
```bash
# Step 1: Preview (dry-run)
@delete-catalog-kind ObsoleteKind --dry-run

# Step 2: Delete with backup
@delete-catalog-kind ObsoleteKind --backup

# Step 3: Confirm (type: DELETE ObsoleteKind)

# Step 4: Verify
go build ./catalog/... && go test ./catalog/...
```

**Quick Delete (with caution):**
```bash
@delete-catalog-kind TestKind --force --backup
```

### What Gets Deleted
- ✅ Kind folder (all files)
- ✅ Registry entry (catalog_kind.proto enum)
- ✅ Generated proto stubs (regenerated)

### Learn More
- **README:** [`delete/README.md`](delete/README.md)
- **Rule:** [`delete/delete-catalog-kind.mdc`](delete/delete-catalog-kind.mdc)

---

## Common Workflows

### Workflow 1: Create and Validate

**Option A: Manual (More Control)**
```bash
# 1. Create new kind
@forge-catalog-kind NewKind --provider aws

# 2. Verify completeness
@audit-catalog-kind NewKind
# Expected: 95-100% complete

# 3. If gaps found, fill them
@update-catalog-kind NewKind --scenario fill-gaps

# 4. Re-audit to verify
@audit-catalog-kind NewKind
# Expected: 100% complete
```

**Option B: Automated (Faster)**
```bash
# 1. Create new kind
@forge-catalog-kind NewKind --provider aws

# 2. Auto-complete if any gaps
@complete-catalog-kind NewKind
# Audits, fills gaps, verifies automatically
# Result: 95-100% complete
```

### Workflow 2: Improve Existing Kind

**Option A: Automated (Recommended)**
```bash
# One command to production-ready
@complete-catalog-kind ExistingKind

# Automatically:
# - Audits current state (65%)
# - Fills all gaps (Terraform, docs, etc.)
# - Re-audits (98%)
# - Reports improvement (+33%)

# Duration: ~18 minutes
```

**Option B: Manual (More Control)**
```bash
# 1. Check current state
@audit-catalog-kind ExistingKind
# Result: 65% complete (missing Terraform, docs)

# 2. Fill identified gaps
@update-catalog-kind ExistingKind --scenario fill-gaps

# 3. Verify improvement
@audit-catalog-kind ExistingKind
# Result: 98% complete
```

### Workflow 3: Add Feature to Kind

```bash
# 1. Edit spec.proto (add new fields)
vim catalog/gcp/gcpcloudsql/v1alpha1/spec.proto

# 2. Propagate changes
@update-catalog-kind GcpCloudSql --scenario proto-changed

# 3. Test changes
# Deploy with e2e/manifest.yaml

# 4. Verify no regressions
@audit-catalog-kind GcpCloudSql
# Score should not decrease
```

### Workflow 4: Replace Kind

```bash
# 1. Check what needs migration
@audit-catalog-kind OldKind
# Result: 40% complete, not worth updating

# 2. Create new implementation
@forge-catalog-kind NewKind --provider aws

# 3. Migrate users (manual step)
# Update references, notify users

# 4. Delete old kind
@delete-catalog-kind OldKind --backup
```

### Workflow 5: Quality Gate (Pre-Commit)

```bash
# Before committing changes to kind
@audit-catalog-kind ModifiedKind

# If score decreased:
# - Investigate what was lost
# - Fix issues
# - Re-audit

# If score maintained or improved:
# - Safe to commit
git add -A
git commit -m "feat: enhance ModifiedKind"
```

---

## Integration Points

### With Git Workflows

```bash
# Feature branch workflow
git checkout -b feature/new-kind

# Create kind
@forge-catalog-kind NewKind --provider gcp

# Validate
@audit-catalog-kind NewKind

# Commit
git add -A
git commit -m "feat: add NewKind for GCP"
git push origin feature/new-kind
```

### With CI/CD

```yaml
# .github/workflows/kind-quality.yml
on: [pull_request]
jobs:
  audit-kinds:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Audit modified kinds
        run: |
          # Detect modified kinds
          # Run audit on each
          # Fail if score < 80%
```

### With Makefiles

```makefile
# Makefile
.PHONY: audit-all
audit-all:
	@for kind in $(shell find catalog -mindepth 2 -maxdepth 2 -type d -not -name "_*" -not -name "aa_*"); do \
		@audit-catalog-kind $$(basename $$kind); \
	done
```

---

## Reference Documents

### Core Documentation
- **Ideal State Definition:** [`../../architecture/catalog-kind.md`](../../architecture/catalog-kind.md)
  - Defines what "complete" means
  - Provides checklist for all requirements
  - Explains the provider parity standard

### Rule Documentation
- **Forge:** [`forge/README.md`](forge/README.md)
- **Audit:** [`audit/README.md`](audit/README.md)
- **Update:** [`update/README.md`](update/README.md)
- **Delete:** [`delete/README.md`](delete/README.md)

### Kind-Specific Documentation
- **Durable Judgment Sources:** `<kind>/GUIDE.md` (when present), `<kind>/README.md`, `<kind>/catalog.md`
  - **Critical Reference:** Consult these when executing any rule on a kind
  - GUIDE.md carries authored operational judgment: design decisions, parity accounting (pinned provider version, compositions, recorded exclusions), conventions, and trade-offs
  - README.md and catalog.md carry the user-facing behavior contract
  - Research documents are never committed -- anything worth keeping lives in these files
  - **Use when:**
    - Updating kind (understand current design)
    - Auditing kind (assess documentation quality)
    - Deleting kind (understand impact)
    - Making decisions about kind behavior

### Flow Rules
- **Forge Flow:** [`forge/flow/`](forge/flow/) - 21 atomic rules for kind creation

---

## Best Practices

### For New Kinds

1. **Always use forge** - Don't create manually
2. **Audit immediately** - Verify forge created everything
3. **Fill any gaps** - Use update if audit shows <95%
4. **Test locally** - Deploy with the e2e manifest
5. **Document decisions** - Update GUIDE.md if judgment changed

### For Existing Kinds

1. **Audit first** - Understand current state
2. **Prioritize critical gaps** - Fix blockers first
3. **Update systematically** - Use appropriate scenario
4. **Re-audit after changes** - Verify improvements
5. **Track progress** - Keep audit reports for history

### For Quality Assurance

1. **Regular audits** - Weekly or monthly health checks
2. **Trend tracking** - Compare audit scores over time
3. **Pre-commit gates** - Audit before committing
4. **Team standards** - Minimum 80% for production
5. **Documentation** - Keep audit reports in git

### For Deletions

1. **Always use dry-run first** - Preview deletion
2. **Always use --backup** - Create safety net
3. **Check references** - Don't break other kinds
4. **Notify team** - Communicate deletions
5. **Document why** - Record rationale in commit message

---

## Troubleshooting

### "Kind not found"

**Check:**
- Kind name spelling (case-sensitive)
- Kind registered in catalog_kind.proto
- Folder exists at expected path

**Solution:**
```bash
# List all kinds
grep "^\s*[A-Z]" shared/catalogkind/catalog_kind.proto
```

### "Audit shows 0% but kind exists"

**Check:**
- Files exist but might be empty
- Folder structure matches conventions
- Minimum file sizes met (proto >500 bytes, etc.)

**Solution:**
```bash
# Check file sizes
find catalog/cloudflare/cloudflared1database -type f -exec ls -lh {} \;
```

### "Update fails with build errors"

**Check:**
- Proto syntax is valid
- Import paths are correct
- Generated stubs are current

**Solution:**
```bash
# Regenerate stubs
make protos

# Check build
go build ./catalog/<provider>/<kind>/...

# Check tests
go test -v ./catalog/<provider>/<kind>/...
```

### "Delete warns about references"

**Options:**
1. Fix references first (recommended)
2. Use --force (may break builds)
3. Cancel deletion

**Solution:**
```bash
# Find all references
grep -r "KindName" catalog/
grep -r "KindName" docs/

# Update references
# Then delete
```

---

## Success Metrics

### Kind Quality

- **100%** = Perfect, production-ready
- **95-99%** = Excellent, minor polish possible
- **80-94%** = Good, some improvements recommended
- **60-79%** = Fair, significant work needed
- **<60%** = Poor, major work or reconsider

### Team Productivity

- **Time to create:** <30 minutes (forge)
- **Time to audit:** <1 minute
- **Time to update:** 10-60 minutes (depends on scenario)
- **Time to delete:** <2 minutes

### Quality Standards

- **New kinds:** ≥95% at creation
- **Production kinds:** ≥80% minimum
- **Active development:** ≥90% target
- **Deprecated kinds:** Audit before delete

---

## Examples by Provider

### AWS Kinds

```bash
@forge-catalog-kind AwsRdsInstance --provider aws
@forge-catalog-kind AwsEksCluster --provider aws
@forge-catalog-kind AwsS3Bucket --provider aws
```

### GCP Kinds

```bash
@forge-catalog-kind GcpCloudSql --provider gcp
@forge-catalog-kind GcpGkeCluster --provider gcp
@forge-catalog-kind GcpStorageBucket --provider gcp
```

### Kubernetes Kinds

```bash
@forge-catalog-kind KubernetesPostgres --provider kubernetes --category workload
@forge-catalog-kind KubernetesValkey --provider kubernetes --category workload
@forge-catalog-kind KubernetesCertManager --provider kubernetes --category addon
```

### SaaS Platform Kinds

```bash
@forge-catalog-kind CloudflareD1Database --provider cloudflare
@forge-catalog-kind Auth0Client --provider auth0
@forge-catalog-kind DigitalOceanBucket --provider digitalocean
```

---

## Getting Help

### Documentation
- Read the README for specific operation
- Check ideal state document for requirements
- Review flow rules for implementation details

### Examples
- See complete kinds (e.g., GcpCertManagerCert)
- Run audit on gold-standard kinds
- Compare incomplete vs complete kinds

### Support
- GitHub Discussions for questions
- GitHub Issues for bugs
- Team chat for quick questions

---

## Contributing

When adding new rules or improving existing ones:

1. **Follow existing patterns** - Consistency matters
2. **Update documentation** - Keep READMEs current
3. **Test thoroughly** - Validate on multiple kinds
4. **Add examples** - Show real usage
5. **Update ideal state** - If requirements change

---

**Ready to start?** Choose the operation you need and follow its README for detailed instructions!

| Operation | Command Template |
|-----------|-----------------|
| **Forge** | `@forge-catalog-kind <Name> --provider <provider>` |
| **Audit** | `@audit-catalog-kind <Name>` |
| **Update** | `@update-catalog-kind <Name> [--scenario <type>]` |
| **Complete** | `@complete-catalog-kind <Name> [--target-score <pct>]` |
| **Fix** | `@fix-catalog-kind <Name> --explain "<fix description>"` |
| **Delete** | `@delete-catalog-kind <Name> --backup` |

