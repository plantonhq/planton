# Audit: Assess Component Completeness

## Overview

**Audit** is the diagnostic rule that evaluates deployment components against the ideal state defined in `architecture/deployment-component.md`. It generates comprehensive, actionable reports showing exactly what's complete, what's missing, and how to achieve 100% completion.

## Why Audit Exists

You can't improve what you don't measure. Audit provides:
- **Objective assessment** - Clear completion percentage
- **Gap identification** - Exactly what's missing
- **Prioritized recommendations** - What to fix first
- **Historical tracking** - See improvement over time
- **Quality assurance** - Validate forge/update results

**Audit makes quality visible and actionable.**

## When to Use Audit

### ✅ Use Audit When

- **After forge** - Verify component was created completely
- **After update** - Confirm improvements were made
- **Before update** - Identify what needs fixing
- **Regular checks** - Periodic quality reviews
- **Quality gates** - Before committing or releasing
- **Understanding state** - New to a component
- **Comparing components** - See which are most complete

### Audit Use Cases

**Quality Assurance:**
```bash
# After creating component
@forge-openmcf-component NewComponent --provider aws
@audit-openmcf-component NewComponent  # Expect 95-100%
```

**Gap Identification:**
```bash
# Check existing component
@audit-openmcf-component MongodbAtlas  # Shows 65% complete
# Report lists missing items
@update-openmcf-component MongodbAtlas --scenario fill-gaps
```

**Progress Tracking:**
```bash
# Monthly audit of all components
for component in $(list-all-components); do
  @audit-openmcf-component $component
done
# Track improvement over time
```

**Pre-Commit Validation:**
```bash
# Before committing changes
@audit-openmcf-component ModifiedComponent
# Ensure score didn't decrease
```

## What Audit Checks

Audit evaluates **9 categories** against the ideal state:

### 1. Cloud Resource Registry (Critical - 4.44%)

- [ ] Enum entry exists in `cloud_resource_kind.proto`
- [ ] Enum value in correct provider range
- [ ] Unique `id_prefix`
- [ ] Complete metadata (provider, version, id_prefix)
- [ ] Kubernetes metadata (if applicable)

### 2. Folder Structure (Critical - 4.44%)

- [ ] Correct provider hierarchy
- [ ] Kubernetes category if applicable (addon/workload/config)
- [ ] Lowercase folder naming
- [ ] v1/ subfolder exists

### 3. Protobuf API Definitions (Critical - 22.20%)

**Proto Files (13.32%):**
- [ ] `api.proto` exists (>500 bytes)
- [ ] `spec.proto` exists (>500 bytes)
- [ ] `stack_input.proto` exists (>300 bytes)
- [ ] `stack_outputs.proto` exists (>300 bytes)

**Generated Stubs (3.33%):**
- [ ] `api.pb.go` exists
- [ ] `spec.pb.go` exists
- [ ] `stack_input.pb.go` exists
- [ ] `stack_outputs.pb.go` exists

**Test File Presence (2.77%):**
- [ ] `spec_test.go` exists (>500 bytes)
- [ ] File contains actual test functions
- [ ] File imports testing framework

**Test Execution (2.78%):**
- [ ] Tests compile without errors
- [ ] Tests execute when running: `go test ./apis/.../provider/<provider>/<component>/v1/`
- [ ] All tests pass (no failures)
- [ ] Tests validate all buf.validate rules are correct

**Critical:** Components with failing tests are considered incomplete. Test execution is mandatory for production-readiness.

### 4. IaC Modules - Pulumi (Critical - 13.32%)

**Module Files:**
- [ ] `iac/pulumi/module/main.go` exists
- [ ] `iac/pulumi/module/locals.go` exists
- [ ] `iac/pulumi/module/outputs.go` exists

**Entrypoint Files:**
- [ ] `iac/pulumi/main.go` exists
- [ ] `iac/pulumi/Pulumi.yaml` exists
- [ ] `iac/pulumi/Makefile` exists

### 5. IaC Modules - Terraform (Critical - 4.44%)

- [ ] `iac/tf/variables.tf` exists (>1KB)
- [ ] `iac/tf/provider.tf` exists
- [ ] `iac/tf/locals.tf` exists
- [ ] `iac/tf/main.tf` exists (>1KB)
- [ ] `iac/tf/outputs.tf` exists

### 6. Documentation - Research (Important - 13.34%)

- [ ] `docs/README.md` exists
- [ ] File is substantial (>10KB for comprehensive)
- [ ] Contains landscape analysis
- [ ] Contains best practices

### 7. Documentation - User-Facing (Important - 13.33%)

- [ ] `README.md` exists at v1/ level (>2KB)
- [ ] `examples.md` exists (>1KB with multiple examples)

### 8. Supporting Files (Important - 13.33%)

**Pulumi:**
- [ ] `iac/pulumi/README.md` exists

**Terraform:**
- [ ] `iac/tf/README.md` exists

**Helpers:**
- [ ] `iac/hack/manifest.yaml` exists
- [ ] `iac/pulumi/debug.sh` exists

### 9. Nice to Have (20%)

Bonus points for extra polish.

## Scoring System

### Weighted Scoring

```
Total Score = Critical (40%) + Important (40%) + Nice-to-Have (20%)

Where:
- Critical = 9 items × 4.44% each = 40%
- Important = 6 items × 6.67% each = 40%
- Nice-to-Have = 3 items × 6.67% each = 20%
```

### Score Interpretation

| Score | Status | Meaning |
|-------|--------|---------|
| **100%** | Fully Complete | Production-ready, all items present |
| **80-99%** | Functionally Complete | Minor improvements needed |
| **60-79%** | Partially Complete | Significant work remaining |
| **40-59%** | Skeleton Exists | Major implementation needed |
| **<40%** | Early Stage | Just started or abandoned |

### Example Scoring

**Component A:**
- Critical items: 38% / 40% (missing 1 test)
- Important items: 35% / 40% (missing research doc)
- Nice-to-have: 15% / 20% (missing some polish)
- **Total: 88% - Functionally Complete**

**Component B:**
- Critical items: 40% / 40% (all present)
- Important items: 40% / 40% (all present)
- Nice-to-have: 20% / 20% (all present)
- **Total: 100% - Fully Complete**

## Audit Report Structure

### Report Header

```markdown
# Audit Report: MongodbAtlas

**Audit Date:** 2025-11-13 14:30:22
**Component Kind:** MongodbAtlas
**Provider:** atlas
**Component Path:** `apis/org/openmcf/provider/atlas/mongodbatlas/v1/`
**Enum Value:** 51
**ID Prefix:** mdbatl
```

### Overall Score

```markdown
## Overall Completion Score

**Score: 65%**

████████░░ 65% Complete

**Status:** Partially Complete
```

### Summary Table

```markdown
## Summary by Category

| Category | Weight | Score | Status |
|----------|--------|-------|--------|
| Cloud Resource Registry | 4.44% | 4.44% | ✅ |
| Folder Structure | 4.44% | 4.44% | ✅ |
| Protobuf API Definitions | 17.76% | 17.76% | ✅ |
| IaC Modules - Pulumi | 13.32% | 13.32% | ✅ |
| IaC Modules - Terraform | 4.44% | 0.00% | ❌ |
| Documentation - Research | 13.34% | 0.00% | ❌ |
| Documentation - User-Facing | 13.33% | 10.00% | ⚠️ |
| Supporting Files | 13.33% | 10.00% | ⚠️ |
| Nice to Have | 20.00% | 5.00% | ⚠️ |

**Legend:** ✅ Complete | ⚠️ Partial | ❌ Missing
```

### Quick Wins

```markdown
## Quick Wins

Items that are easy to fix and would improve the score:

1. **Generate Terraform Module** - Add 4.44%
   - Run rules 013-015
   - Creates complete Terraform implementation
   - 15-20 minutes

2. **Generate Research Documentation** - Add 13.34%
   - Run rule 020
   - Creates comprehensive docs/README.md
   - 10-15 minutes

**Total Quick Win Potential: +17.78%**
```

### Critical Gaps

```markdown
## Critical Gaps

Blocking issues that prevent production readiness:

1. **Missing Terraform Module** - 4.44% missing
   - **Why it matters:** Users need choice between Pulumi and Terraform
   - **What to do:** Run `@update-openmcf-component MongodbAtlas --scenario fill-gaps`
   - **Forge rules:** 013-015

2. **Missing Research Documentation** - 13.34% missing
   - **Why it matters:** Platform engineers need context for design decisions
   - **What to do:** Run rule 020 to generate docs/README.md
   - **Expected outcome:** Comprehensive landscape analysis (300-1000 lines)
```

### Detailed Findings (Per Category)

```markdown
## Detailed Findings

### 1. Cloud Resource Registry (4.44%)

✅ **Passed:**
- Enum entry exists: MongodbAtlas = 51
- Enum value in correct range (50-199 for SaaS)
- Unique id_prefix: "mdbatl"
- Complete metadata (provider, version, id_prefix)

**Score:** 4.44% / 4.44% ✅

---

### 2. Folder Structure (4.44%)

✅ **Passed:**
- Correct provider hierarchy: apis/org/openmcf/provider/atlas/
- Lowercase folder naming: mongodbatlas (matches enum)
- v1/ subfolder exists

**Score:** 4.44% / 4.44% ✅

---

### 3. Protobuf API Definitions (17.76%)

✅ **Passed:**
- api.proto exists (1.2 KB) ✅
- spec.proto exists (2.5 KB) ✅
- stack_input.proto exists (800 bytes) ✅
- stack_outputs.proto exists (600 bytes) ✅
- All .pb.go stubs exist (4 files) ✅
- spec_test.go exists (1.8 KB) ✅

**Score:** 17.76% / 17.76% ✅

---

### 4. IaC Modules - Pulumi (13.32%)

✅ **Passed:**
- Module files exist:
  - module/main.go (3.2 KB) ✅
  - module/locals.go (1.8 KB) ✅
  - module/outputs.go (1.1 KB) ✅
- Entrypoint files exist:
  - main.go (450 bytes) ✅
  - Pulumi.yaml (220 bytes) ✅
  - Makefile (800 bytes) ✅

**Score:** 13.32% / 13.32% ✅

---

### 5. IaC Modules - Terraform (4.44%)

❌ **Failed:**
- iac/tf/ directory does not exist
- No Terraform implementation found

**Score:** 0.00% / 4.44% ❌

**Fix:** Run `@update-openmcf-component MongodbAtlas --scenario fill-gaps`

---

### 6. Documentation - Research (13.34%)

❌ **Failed:**
- docs/README.md does not exist
- No research documentation found

**Score:** 0.00% / 13.34% ❌

**Fix:** Run rule 020 to generate comprehensive research document

---

... (continues for all categories)
```

### Prioritized Recommendations

```markdown
## Prioritized Recommendations

### High Priority (Do First)

1. **Create Terraform Module**
   - **File:** `iac/tf/` (multiple files)
   - **Why:** Critical for feature parity between IaC tools
   - **How:** `@update-openmcf-component MongodbAtlas --scenario fill-gaps`
   - **Impact:** +4.44% (65% → 69.44%)

2. **Create Research Documentation**
   - **File:** `v1/docs/README.md`
   - **Why:** Essential for understanding design decisions
   - **How:** Run forge rule 020
   - **Impact:** +13.34% (69.44% → 82.78%)

### Medium Priority (Do Next)

3. **Expand Examples**
   - **File:** `examples.md`
   - **Why:** Only 1 example, need 3-5 for completeness
   - **How:** Add more use cases to examples.md
   - **Impact:** +3.33% (87.78% → 91.11%)

### Low Priority (Polish)

5. **Add Terraform README**
   - **File:** `iac/tf/README.md`
   - **Why:** Usage documentation for Terraform users
   - **How:** Generate with rule 015
   - **Impact:** +3.33% (91.11% → 94.44%)
```

### Comparison

```markdown
## Comparison to Complete Components

**Most Similar Complete Component:** GcpCertManagerCert (98% complete)

**What it has that MongodbAtlas lacks:**
- Complete Terraform module (variables.tf, main.tf, outputs.tf, etc.)
- Comprehensive research documentation (850 lines)
- Multiple examples (5 use cases)
- Pulumi architecture overview
- Complete IaC documentation

**Path to Reference:** `apis/org/openmcf/provider/gcp/gcpcertmanagercert/v1/`

**Recommendation:** Review GcpCertManagerCert as a template for completeness.
```

### Next Steps

```markdown
## Next Steps

1. Address critical gaps (Terraform + research docs)
2. Run update to fill gaps:
   ```
   @update-openmcf-component MongodbAtlas --scenario fill-gaps
   ```
3. Re-run audit to verify improvements:
   ```
   @audit-openmcf-component MongodbAtlas
   ```
4. Expected result: 95-100% complete

**Estimated time to 100%:** 30-45 minutes
```

## Report Storage

Audit reports are saved with timestamps:

```
apis/org/openmcf/provider/atlas/mongodbatlas/v1/docs/audit/
├── 2025-11-10-091500.md  # First audit (60%)
├── 2025-11-11-143000.md  # After adding Terraform (75%)
└── 2025-11-13-143022.md  # After adding docs (98%)
```

**Benefits:**
- **Historical tracking** - See improvement over time
- **Comparison** - Compare audits to measure progress
- **Documentation** - Record of component evolution
- **Quality gates** - Validate before releases

## Usage Examples

### Example 1: Check New Component

```bash
# After forge
@forge-openmcf-component NewComponent --provider gcp

# Verify completeness
@audit-openmcf-component NewComponent

# Expected: 95-100% complete
# If lower, identifies what's missing
```

### Example 2: Find Gaps to Fill

```bash
# Audit existing component
@audit-openmcf-component MongodbAtlas
# Result: 65% complete (missing Terraform, docs)

# Fill gaps
@update-openmcf-component MongodbAtlas --scenario fill-gaps

# Verify improvement
@audit-openmcf-component MongodbAtlas
# Result: 98% complete
```

### Example 3: Quality Gate

```bash
# Before committing
@audit-openmcf-component ModifiedComponent

# If score decreased:
# - Investigate what was lost
# - Fix before committing
# - Re-audit

# If score same or improved:
# - Safe to commit
```

### Example 4: Batch Audit

```bash
# Audit all components
@audit-all-components --output-summary

# Output:
# 45 components audited
# Average score: 82%
# 12 components need attention (<80%)
# 33 components production-ready (≥80%)
```

## Interpreting Results

### 100% Complete

**What it means:**
- All required files present
- All documentation complete
- Both IaC modules implemented
- Tests passing
- Build succeeds

**Action:** None needed, production-ready!

### 80-99% Complete

**What it means:**
- Core functionality present
- Minor items missing (polish, extra docs, etc.)
- Functionally complete

**Action:** Optional improvements for perfection

### 60-79% Complete

**What it means:**
- Major implementation gaps
- Missing critical pieces (IaC module, docs)
- Not production-ready

**Action:** Run update to fill gaps (30-60 minutes)

### 40-59% Complete

**What it means:**
- Skeleton exists
- Significant work needed
- Major components missing

**Action:** Consider re-running forge or extensive updates

### <40% Complete

**What it means:**
- Barely started or abandoned
- Most items missing

**Action:** Consider starting over with forge

## Integration with Other Rules

### Audit → Update → Audit

```bash
# 1. Initial audit
@audit-openmcf-component MongodbAtlas
# Result: 65%

# 2. Fill gaps
@update-openmcf-component MongodbAtlas --scenario fill-gaps

# 3. Verify improvement
@audit-openmcf-component MongodbAtlas
# Result: 98%
```

### Forge → Audit

```bash
# 1. Create component
@forge-openmcf-component NewComponent --provider aws

# 2. Validate
@audit-openmcf-component NewComponent
# Result: Should be 95-100%
```

### Audit → Delete

```bash
# 1. Check if worth keeping
@audit-openmcf-component OldComponent
# Result: 35% (very incomplete)

# 2. Decision: Not worth fixing
@delete-openmcf-component OldComponent --backup
```

## Best Practices

### When to Run Audit

- ✅ **After forge** - Validate creation
- ✅ **After update** - Confirm improvements
- ✅ **Before committing** - Quality gate
- ✅ **Weekly/monthly** - Regular health checks
- ✅ **Before release** - Final validation
- ✅ **When onboarding** - Understand component state

### Interpreting Scores

- **100%** = Perfect, no action needed
- **95-99%** = Excellent, minor polish possible
- **80-94%** = Good, some improvements recommended
- **60-79%** = Fair, significant work needed
- **<60%** = Poor, major work or reconsider

### Using Reports

- Read quick wins first (easy improvements)
- Address critical gaps before medium priority
- Compare to similar complete components
- Track improvement with historical reports
- Share reports in PRs for transparency

## Tips

### Getting to 100%

1. Start with critical gaps (40% weight)
2. Add important items (40% weight)
3. Polish with nice-to-haves (20% weight)
4. Re-audit after each phase

### Understanding Categories

- **Critical (40%)** = Must-have for functionality
- **Important (40%)** = Must-have for quality
- **Nice-to-have (20%)** = Polish and extras

### Quick Score Improvements

- **Missing Terraform?** +4.44% (run rules 013-015)
- **Missing research docs?** +13.34% (run rule 020)
- **Missing examples?** +6.66% (enhance examples.md)

## Troubleshooting

### Audit Shows 0% but Component Exists

**Check:**
1. Component path correct?
2. Files named correctly?
3. Minimum file sizes met?
4. Enum entry exists?

### Score Lower Than Expected

**Check:**
1. File sizes (some must be substantial)
2. All required files present
3. Tests actually exist and pass
4. Documentation is comprehensive

### Audit Fails to Run

**Check:**
1. Component name spelled correctly
2. Component registered in cloud_resource_kind.proto
3. Folder structure matches conventions

## Success Metrics

Good audit outcomes:

- ✅ Clear completion percentage
- ✅ Specific gaps identified
- ✅ Actionable recommendations
- ✅ Historical report saved
- ✅ Path to 100% clear

## Related Commands

- `@forge-openmcf-component` - Create new component
- `@update-openmcf-component` - Fill gaps, enhance component
- `@complete-openmcf-component` - Auto-improve to 95%+ (audit + update + audit)
- `@fix-openmcf-component` - Targeted fixes with cascading updates
- `@delete-openmcf-component` - Remove component

## Reference

- **Ideal State Definition:** `architecture/deployment-component.md`
- **Audit Rule:** `_rules/deployment-component/audit/audit-openmcf-component.mdc`

---

**Ready to audit?** Run `@audit-openmcf-component <ComponentName>` to generate a comprehensive report!
