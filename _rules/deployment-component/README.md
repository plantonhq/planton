# Deployment Component Lifecycle Management

## Overview

This directory contains the complete rule system for managing deployment components in OpenMCF. These rules handle the entire lifecycle from creation to deletion, ensuring components match the **ideal state** defined in `architecture/deployment-component.md`.

## The Seven Lifecycle Operations

OpenMCF provides seven operations for deployment components:

| Operation | Purpose | When to Use |
|-----------|---------|-------------|
| **🔨 Forge** | Create new components | Component doesn't exist |
| **🔍 Audit** | Assess completeness | Check status, find gaps |
| **🔄 Update** | Enhance existing components | Fill gaps, refresh docs, general improvements |
| **✨ Complete** | Auto-improve to target | One-command: audit + fill gaps + verify |
| **🔧 Fix** | Targeted fixes with propagation | Specific bugs, sync issues, consistency fixes |
| **✏️ Rename** | Systematically rename components | Name clarity, remove abstractions, consistency |
| **🗑️ Delete** | Remove components | Obsolete, deprecated, consolidating |

**Key Principles:** 
- Each operation is atomic, well-documented, and follows the ideal state standard
- **Source code is the ultimate source of truth** - documentation must match implementation
- Complete is a convenience wrapper (audit + update + audit)
- Fix ensures consistency across all artifacts (code, docs, examples, tests)

---

## Quick Decision Tree

```
Need to work with a deployment component?
│
├─ Does the component exist?
│  │
│  ├─ NO → Use @forge-openmcf-component
│  │        Creates complete, production-ready component
│  │        Expected result: 95-100% complete
│  │
│  └─ YES → What do you need to do?
│     │
│     ├─ Just checking status?
│     │  └─ Use @audit-openmcf-component
│     │     Shows completion %, identifies gaps
│     │     Generates timestamped report
│     │
│     ├─ Make it production-ready quickly?
│     │  └─ Use @complete-openmcf-component
│     │     Audits + fills all gaps + verifies
│     │     One command to 95%+ completion
│     │
│     ├─ Have specific bug/issue to fix?
│     │  └─ Use @fix-openmcf-component
│     │     Targeted fix with cascading updates
│     │     Ensures code, docs, examples, tests all match
│     │     Source code is truth, docs updated to match
│     │
│     ├─ Need general improvements?
│     │  └─ Use @update-openmcf-component
│     │     Fills gaps, refreshes docs, updates IaC
│     │     6 scenarios: fill-gaps, proto-changed, etc.
│     │
│     ├─ Need to rename component?
│     │  └─ Use @rename-openmcf-component
│     │     Systematic rename across entire codebase
│     │     7 naming patterns, build verification
│     │     Name clarity, remove abstractions
│     │
│     └─ Want to remove?
│        └─ Use @delete-openmcf-component
│           Safe deletion with backup and confirmation
│           Checks for references, creates backup
```

---

## 1. Forge: Create New Components

### Purpose
Bootstrap complete, production-ready deployment components from scratch.

### When to Use
- Creating support for a new cloud provider resource
- Adding a new SaaS platform integration
- Creating a new Kubernetes workload or addon

### What It Creates

✅ **Proto API Definitions** - All 4 proto files with validations and tests
✅ **IaC Modules** - Both Pulumi and Terraform with feature parity
✅ **Documentation** - User-facing, research, and technical docs
✅ **Supporting Files** - Test manifests, debug scripts, build configs
✅ **Registry Entry** - Registered in cloud_resource_kind.proto
✅ **Validation** - Build and test validation passed

**Result:** 95-100% completion score

### Usage

```bash
@forge-openmcf-component <ComponentName> --provider <provider>
```

**Examples:**
```bash
@forge-openmcf-component MongodbAtlas --provider atlas
@forge-openmcf-component GcpStorageBucket --provider gcp
@forge-openmcf-component PostgresKubernetes --provider kubernetes --category workload
```

### Learn More
- **README:** [`forge/README.md`](forge/README.md)
- **Rule:** [`forge/forge-openmcf-component.mdc`](forge/forge-openmcf-component.mdc)
- **Flow Rules:** [`forge/flow/`](forge/flow/)

---

## 2. Audit: Assess Component Completeness

### Purpose
Evaluate components against the ideal state and generate actionable completion reports.

### When to Use
- After forge (verify component was created completely)
- After update (confirm improvements)
- Before update (identify what needs fixing)
- Regular quality checks
- Pre-commit validation

### What It Checks

Audits 9 categories:
1. Cloud Resource Registry
2. Folder Structure
3. Protobuf API Definitions
4. IaC Modules - Pulumi
5. IaC Modules - Terraform
6. Documentation - Research
7. Documentation - User-Facing
8. Supporting Files
9. Nice to Have Items

**Scoring:** Weighted (Critical 40%, Important 40%, Nice-to-Have 20%)

### Usage

```bash
@audit-openmcf-component <ComponentName>
```

**Examples:**
```bash
@audit-openmcf-component MongodbAtlas
@audit-openmcf-component GcpCertManagerCert
```

### Report Output

- **Overall completion percentage** (0-100%)
- **Category-by-category breakdown**
- **Quick wins** (easy improvements)
- **Critical gaps** (blocking issues)
- **Prioritized recommendations**
- **Comparison to complete components**

Reports saved to: `<component>/v1/docs/audit/<timestamp>.md`

### Learn More
- **README:** [`audit/README.md`](audit/README.md)
- **Rule:** [`audit/audit-openmcf-component.mdc`](audit/audit-openmcf-component.mdc)

---

## 3. Update: Enhance Existing Components

### Purpose
Improve existing components by filling gaps, adding features, refreshing docs, or fixing issues.

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
@update-openmcf-component <ComponentName> [--scenario <type>] [--explain "<description>"]
```

**Examples:**
```bash
# Fill gaps from audit
@update-openmcf-component MongodbAtlas --scenario fill-gaps

# Propagate proto changes
@update-openmcf-component GcpCertManagerCert --scenario proto-changed

# Refresh documentation
@update-openmcf-component AwsRdsInstance --scenario refresh-docs

# Update IaC implementation
@update-openmcf-component GcpGkeCluster --scenario update-iac --explain "add multi-region support"

# Fix specific issue
@update-openmcf-component PostgresKubernetes --explain "examples use deprecated field names"
```

### Safety Features
- ✅ Dry-run mode (`--dry-run`)
- ✅ Backup creation (`--backup`)
- ✅ Validation checkpoints
- ✅ Automatic retry (up to 3 times)
- ✅ Conflict detection

### Learn More
- **README:** [`update/README.md`](update/README.md)
- **Rule:** [`update/update-openmcf-component.mdc`](update/update-openmcf-component.mdc)

---

## 4. Complete: Auto-Improve to Production-Ready

### Purpose
One-command workflow that audits a component and automatically fills all gaps to reach target completion score (default 95%).

### When to Use
- Making component production-ready quickly
- Batch improving multiple components
- Quality gates before releases
- Onboarding legacy components
- Following up after forge

### What It Does

**Three-Step Automated Workflow:**
1. **Audit** - Assess current state and identify all gaps
2. **Fill Gaps** - Automatically run update --fill-gaps
3. **Verify** - Re-audit to confirm improvement

**Result:** Before/after comparison showing improvement

### Usage

```bash
@complete-openmcf-component <ComponentName> [flags]
```

**Examples:**
```bash
# Basic usage (target: 95%)
@complete-openmcf-component MongodbAtlas

# Preview without changes
@complete-openmcf-component MongodbAtlas --dry-run

# Custom target score
@complete-openmcf-component PostgresKubernetes --target-score 100

# Batch processing
for component in Comp1 Comp2 Comp3; do
  @complete-openmcf-component $component
done
```

### What Gets Filled

Automatically creates missing items:
- ✅ Terraform module (if missing)
- ✅ Research documentation (if missing)
- ✅ User-facing docs (if missing/incomplete)
- ✅ Examples (if missing/incomplete)
- ✅ Pulumi overview (if missing)
- ✅ Supporting files (manifests, debug scripts)

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
- **Rule:** [`complete/complete-openmcf-component.mdc`](complete/complete-openmcf-component.mdc)

---

## 5. Fix: Targeted Fixes with Cascading Updates

### Purpose
Make targeted fixes to components and automatically propagate changes to all related artifacts (documentation, examples, tests, IaC modules) to ensure complete consistency.

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
5. **Execute Tests** - Component tests, build, full suite
6. **Report** - Show what was fixed and what was propagated

**Five Consistency Checks:**
- Proto ↔ Terraform variables
- Proto ↔ Examples (examples must validate)
- Pulumi ↔ Terraform (feature parity)
- Validations ↔ Tests (every rule tested)
- Documentation ↔ Implementation (docs match reality)

### Usage

```bash
@fix-openmcf-component <ComponentName> --explain "<detailed fix description>"
```

**Examples:**
```bash
# Fix validation logic
@fix-openmcf-component GcpCertManagerCert --explain "primaryDomainName validation should allow wildcards like *.example.com"

# Fix IaC implementation
@fix-openmcf-component AwsRdsInstance --explain "Pulumi hardcodes backup_retention_period instead of using spec field"

# Fix documentation drift
@fix-openmcf-component PostgresKubernetes --explain "examples use deprecated 'database_name' field, should be 'db_identifier'"

# Fix test failures
@fix-openmcf-component MongodbAtlas --explain "spec_test.go expects validation on cluster_tier but spec.proto has no validation rule"
```

### What Gets Updated

**Source Code (if needed):**
- spec.proto (validation rules, fields)
- Pulumi module (deployment logic)
- Terraform module (to maintain parity)
- spec_test.go (validation tests)

**Documentation (always):**
- examples.md (match current API)
- README.md (match current behavior)
- docs/README.md (if architectural)
- IaC READMEs (if usage changed)

**Validation:**
- Component tests: `go test ./apis/.../v1/`
- Build: `go build ./apis/org/openmcf/provider/<provider>/<component>/v1/...`
- Full suite: `go test -v ./apis/org/openmcf/provider/<provider>/<component>/v1/`
- Example validation
- Consistency checks

### Typical Duration

- Documentation-only fix: 2-5 minutes
- Proto + docs fix: 5-10 minutes
- IaC + docs fix: 10-20 minutes
- Complex multi-artifact fix: 20-30 minutes

### Learn More
- **README:** [`fix/README.md`](fix/README.md)
- **Rule:** [`fix/fix-openmcf-component.mdc`](fix/fix-openmcf-component.mdc)

---

## 6. Rename: Systematically Rename Components

### Purpose
Rename deployment components across the entire codebase with comprehensive find-replace patterns, registry updates, and build verification.

### When to Use
- Removing abstractions (e.g., `KubernetesMicroservice` → `KubernetesDeployment`)
- Improving name clarity and accuracy
- Establishing naming consistency
- Preparing for component expansion
- When component name doesn't reflect actual behavior

### Philosophy
**Rename is about semantic truth, not functionality changes.**

Component rename updates names everywhere while preserving all functionality, enum values, and behavior. It ensures names accurately reflect what components do.

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
@rename-openmcf-component
```

The rule will interactively ask for:
1. Old component name (PascalCase)
2. New component name (PascalCase)
3. New ID prefix (optional, press Enter to keep existing)

**Example:**
```bash
@rename-openmcf-component

Old component name: KubernetesMicroservice
New component name: KubernetesDeployment
New ID prefix (current: k8sms): k8sdpl
```

### What Gets Updated

✅ **Component directory** - Copied to new name, old deleted
✅ **All code references** - 7 patterns applied to all files
✅ **Registry** - Enum name and optional ID prefix updated
✅ **Documentation** - All references in `site/public/docs/`
✅ **Build artifacts** - Proto stubs regenerated

### What Gets Preserved

✅ **Enum value** - Registry number unchanged (e.g., 810)
✅ **Provider** - Provider field preserved
✅ **Version** - Remains v1
✅ **Flags** - Special flags like `is_service_kind` preserved
✅ **Metadata** - All other metadata unchanged
✅ **Functionality** - Zero behavioral changes

### Build Pipeline

Rename isn't complete until all phases pass:
1. `make protos` - Regenerate proto stubs
2. `go build ./apis/org/openmcf/provider/<provider>/<component>/v1/...` - Verify compilation
3. `go test -v ./apis/org/openmcf/provider/<provider>/<component>/v1/` - Validate behavior unchanged

**Stops on first failure** for fast feedback.

### Post-Rename

If all tests pass:
- ✅ Automatically invokes `@create-openmcf-changelog`
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

November 2025: All 23 Kubernetes workload components renamed from suffix to prefix pattern:
- `PostgresKubernetes` → `KubernetesPostgres`
- `RedisKubernetes` → `KubernetesRedis`
- ~500 files modified, all builds passed, zero behavioral changes

See: `_changelog/2025-11/2025-11-14-072635-kubernetes-workload-naming-consistency.md`

### Typical Duration

- Simple rename: 1-3 minutes
- Complex rename (large component): 3-7 minutes
- Build pipeline: 30-90 seconds

### Learn More
- **README:** [`rename/README.md`](rename/README.md)
- **Rule:** [`rename/rename-openmcf-component.mdc`](rename/rename-openmcf-component.mdc)
- **Script:** [`rename/_scripts/rename_deployment_component.py`](rename/_scripts/rename_deployment_component.py)

---

## 7. Delete: Remove Components Safely

### Purpose
Completely remove deployment components with safety features to prevent accidents.

### When to Use
- Component is obsolete or deprecated
- Provider discontinued service
- Consolidating similar components
- Cleaning up test components

### Safety Features

🔍 **Dry-Run Mode** - Preview what would be deleted
💾 **Automatic Backup** - Creates timestamped backup
🔎 **Reference Check** - Warns if component is referenced
✋ **Confirmation Required** - Must type component name
📋 **Detailed Report** - Shows exactly what was deleted

### Usage

```bash
@delete-openmcf-component <ComponentName> [flags]
```

**Recommended Workflow:**
```bash
# Step 1: Preview (dry-run)
@delete-openmcf-component ObsoleteComponent --dry-run

# Step 2: Delete with backup
@delete-openmcf-component ObsoleteComponent --backup

# Step 3: Confirm (type: DELETE ObsoleteComponent)

# Step 4: Verify
go build ./apis/.../v1/... && go test -v ./apis/.../v1/
```

**Quick Delete (with caution):**
```bash
@delete-openmcf-component TestComponent --force --backup
```

### What Gets Deleted
- ✅ Component folder (all files)
- ✅ Registry entry (cloud_resource_kind.proto enum)
- ✅ Generated proto stubs (regenerated)

### Learn More
- **README:** [`delete/README.md`](delete/README.md)
- **Rule:** [`delete/delete-openmcf-component.mdc`](delete/delete-openmcf-component.mdc)

---

## Common Workflows

### Workflow 1: Create and Validate

**Option A: Manual (More Control)**
```bash
# 1. Create new component
@forge-openmcf-component NewComponent --provider aws

# 2. Verify completeness
@audit-openmcf-component NewComponent
# Expected: 95-100% complete

# 3. If gaps found, fill them
@update-openmcf-component NewComponent --scenario fill-gaps

# 4. Re-audit to verify
@audit-openmcf-component NewComponent
# Expected: 100% complete
```

**Option B: Automated (Faster)**
```bash
# 1. Create new component
@forge-openmcf-component NewComponent --provider aws

# 2. Auto-complete if any gaps
@complete-openmcf-component NewComponent
# Audits, fills gaps, verifies automatically
# Result: 95-100% complete
```

### Workflow 2: Improve Existing Component

**Option A: Automated (Recommended)**
```bash
# One command to production-ready
@complete-openmcf-component ExistingComponent

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
@audit-openmcf-component ExistingComponent
# Result: 65% complete (missing Terraform, docs)

# 2. Fill identified gaps
@update-openmcf-component ExistingComponent --scenario fill-gaps

# 3. Verify improvement
@audit-openmcf-component ExistingComponent
# Result: 98% complete
```

### Workflow 3: Add Feature to Component

```bash
# 1. Edit spec.proto (add new fields)
vim apis/org/openmcf/provider/gcp/gcpcloudsl/v1/spec.proto

# 2. Propagate changes
@update-openmcf-component GcpCloudSql --scenario proto-changed

# 3. Test changes
# Deploy with iac/hack/manifest.yaml

# 4. Verify no regressions
@audit-openmcf-component GcpCloudSql
# Score should not decrease
```

### Workflow 4: Replace Component

```bash
# 1. Check what needs migration
@audit-openmcf-component OldComponent
# Result: 40% complete, not worth updating

# 2. Create new implementation
@forge-openmcf-component NewComponent --provider aws

# 3. Migrate users (manual step)
# Update references, notify users

# 4. Delete old component
@delete-openmcf-component OldComponent --backup
```

### Workflow 5: Quality Gate (Pre-Commit)

```bash
# Before committing changes to component
@audit-openmcf-component ModifiedComponent

# If score decreased:
# - Investigate what was lost
# - Fix issues
# - Re-audit

# If score maintained or improved:
# - Safe to commit
git add -A
git commit -m "feat: enhance ModifiedComponent"
```

---

## Integration Points

### With Git Workflows

```bash
# Feature branch workflow
git checkout -b feature/new-component

# Create component
@forge-openmcf-component NewComponent --provider gcp

# Validate
@audit-openmcf-component NewComponent

# Commit
git add -A
git commit -m "feat: add NewComponent for GCP"
git push origin feature/new-component
```

### With CI/CD

```yaml
# .github/workflows/component-quality.yml
on: [pull_request]
jobs:
  audit-components:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Audit modified components
        run: |
          # Detect modified components
          # Run audit on each
          # Fail if score < 80%
```

### With Makefiles

```makefile
# Makefile
.PHONY: audit-all
audit-all:
	@for component in $(shell find apis/org/openmcf/provider -name "v1" -type d); do \
		@audit-openmcf-component $$(basename $$(dirname $$component)); \
	done
```

---

## Reference Documents

### Core Documentation
- **Ideal State Definition:** [`../../architecture/deployment-component.md`](../../architecture/deployment-component.md)
  - Defines what "complete" means
  - Provides checklist for all requirements
  - Explains 80/20 principle

### Rule Documentation
- **Forge:** [`forge/README.md`](forge/README.md)
- **Audit:** [`audit/README.md`](audit/README.md)
- **Update:** [`update/README.md`](update/README.md)
- **Delete:** [`delete/README.md`](delete/README.md)

### Component-Specific Documentation
- **Research Document:** `<component>/v1/docs/README.md`
  - **Critical Reference:** Consult this when executing any rule on a component
  - Contains comprehensive research about the component
  - Explains design decisions and 80/20 scoping rationale
  - Documents deployment landscape and best practices
  - Provides context for understanding what the component does and why
  - **Use when:**
    - Updating component (understand current design)
    - Auditing component (assess research quality)
    - Deleting component (understand impact)
    - Making decisions about component behavior

### Flow Rules
- **Forge Flow:** [`forge/flow/`](forge/flow/) - 21 atomic rules for component creation

---

## Best Practices

### For New Components

1. **Always use forge** - Don't create manually
2. **Audit immediately** - Verify forge created everything
3. **Fill any gaps** - Use update if audit shows <95%
4. **Test locally** - Deploy with hack manifest
5. **Document decisions** - Update research docs if needed

### For Existing Components

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
3. **Check references** - Don't break other components
4. **Notify team** - Communicate deletions
5. **Document why** - Record rationale in commit message

---

## Troubleshooting

### "Component not found"

**Check:**
- Component name spelling (case-sensitive)
- Component registered in cloud_resource_kind.proto
- Folder exists at expected path

**Solution:**
```bash
# List all components
grep "^\s*[A-Z]" apis/org/openmcf/shared/cloudresourcekind/cloud_resource_kind.proto
```

### "Audit shows 0% but component exists"

**Check:**
- Files exist but might be empty
- Folder structure matches conventions
- Minimum file sizes met (proto >500 bytes, etc.)

**Solution:**
```bash
# Check file sizes
find apis/org/openmcf/provider/atlas/mongodbatlas/v1 -type f -exec ls -lh {} \;
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
go build ./apis/org/openmcf/provider/<provider>/<component>/v1/...

# Check tests
go test -v ./apis/org/openmcf/provider/<provider>/<component>/v1/
```

### "Delete warns about references"

**Options:**
1. Fix references first (recommended)
2. Use --force (may break builds)
3. Cancel deletion

**Solution:**
```bash
# Find all references
grep -r "ComponentName" apis/
grep -r "ComponentName" docs/

# Update references
# Then delete
```

---

## Success Metrics

### Component Quality

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

- **New components:** ≥95% at creation
- **Production components:** ≥80% minimum
- **Active development:** ≥90% target
- **Deprecated components:** Audit before delete

---

## Examples by Provider

### AWS Components

```bash
@forge-openmcf-component AwsRdsInstance --provider aws
@forge-openmcf-component AwsEksCluster --provider aws
@forge-openmcf-component AwsS3Bucket --provider aws
```

### GCP Components

```bash
@forge-openmcf-component GcpCloudSql --provider gcp
@forge-openmcf-component GcpGkeCluster --provider gcp
@forge-openmcf-component GcpStorageBucket --provider gcp
```

### Kubernetes Components

```bash
@forge-openmcf-component PostgresKubernetes --provider kubernetes --category workload
@forge-openmcf-component RedisKubernetes --provider kubernetes --category workload
@forge-openmcf-component CertManagerKubernetes --provider kubernetes --category addon
```

### SaaS Platform Components

```bash
@forge-openmcf-component MongodbAtlas --provider atlas
@forge-openmcf-component ConfluentKafka --provider confluent
@forge-openmcf-component SnowflakeDatabase --provider snowflake
```

---

## Getting Help

### Documentation
- Read the README for specific operation
- Check ideal state document for requirements
- Review flow rules for implementation details

### Examples
- See complete components (e.g., GcpCertManagerCert)
- Run audit on gold-standard components
- Compare incomplete vs complete components

### Support
- GitHub Discussions for questions
- GitHub Issues for bugs
- Team chat for quick questions

---

## Contributing

When adding new rules or improving existing ones:

1. **Follow existing patterns** - Consistency matters
2. **Update documentation** - Keep READMEs current
3. **Test thoroughly** - Validate on multiple components
4. **Add examples** - Show real usage
5. **Update ideal state** - If requirements change

---

**Ready to start?** Choose the operation you need and follow its README for detailed instructions!

| Operation | Command Template |
|-----------|-----------------|
| **Forge** | `@forge-openmcf-component <Name> --provider <provider>` |
| **Audit** | `@audit-openmcf-component <Name>` |
| **Update** | `@update-openmcf-component <Name> [--scenario <type>]` |
| **Complete** | `@complete-openmcf-component <Name> [--target-score <pct>]` |
| **Fix** | `@fix-openmcf-component <Name> --explain "<fix description>"` |
| **Delete** | `@delete-openmcf-component <Name> --backup` |

