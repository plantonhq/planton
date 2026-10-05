# Audit: Assess Kind Completeness

## Overview

**Audit** is the diagnostic rule that evaluates kinds against the ideal state defined in `architecture/catalog-kind.md`. It generates comprehensive, actionable reports showing exactly what's complete, what's missing, and how to achieve 100% completion.

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

- **After forge** - Verify kind was created completely
- **After update** - Confirm improvements were made
- **Before update** - Identify what needs fixing
- **Regular checks** - Periodic quality reviews
- **Quality gates** - Before committing or releasing
- **Understanding state** - New to a kind
- **Comparing kinds** - See which are most complete

### Audit Use Cases

**Quality Assurance:**
```bash
# After creating kind
@forge-catalog-kind NewKind --provider aws
@audit-catalog-kind NewKind  # Expect 95-100%
```

**Gap Identification:**
```bash
# Check existing kind
@audit-catalog-kind CloudflareD1Database  # Shows 65% complete
# Report lists missing items
@update-catalog-kind CloudflareD1Database --scenario fill-gaps
```

**Progress Tracking:**
```bash
# Monthly audit of all kinds
for kind in $(list-all-kinds); do
  @audit-catalog-kind $kind
done
# Track improvement over time
```

**Pre-Commit Validation:**
```bash
# Before committing changes
@audit-catalog-kind ModifiedKind
# Ensure score didn't decrease
```

## What Audit Checks

Audit evaluates **9 categories** against the ideal state:

### 1. Catalog Kind Registry (Critical - 4.44%)

- [ ] Enum entry exists in `catalog_kind.proto`
- [ ] Enum value in correct provider range
- [ ] Unique `id_prefix`
- [ ] Complete metadata (provider, version, id_prefix)
- [ ] Kubernetes metadata (if applicable)

### 2. Anatomy Conformance (Critical - 4.44%)

Folder structure and file-set presence/absence are machine-enforced by the `pkg/anatomy`
conformance gate (CI lane `lint.kind-anatomy.yaml`); the audit runs the gate instead
of re-checking file inventories by hand:

- [ ] `go test ./pkg/anatomy/...` reports no violation for this kind
- [ ] Any `pkg/anatomy/baseline.yaml` entries for this kind are reported as burn-down gaps

### 3. Protobuf API Definitions (Critical - 22.20%)

**Proto Files (13.32%):** (presence is gate-owned; judge substance)
- [ ] `api.proto` is substantial (>500 bytes) with correct apiVersion/kind constants
- [ ] `spec.proto` is substantial (>500 bytes) with documented, validated fields
- [ ] `input.proto` carries the `{Kind}IacInput` message (>300 bytes)
- [ ] `outputs.proto` carries the `{Kind}Outputs` message (>300 bytes)

**Generated Stubs (3.33%):**
- [ ] All four `.pb.go` stubs are regenerated and current

**Test File Presence (2.77%):**
- [ ] `spec_test.go` exists (>500 bytes)
- [ ] File contains actual test functions
- [ ] File imports testing framework

**Test Execution (2.78%):**
- [ ] Tests compile without errors
- [ ] Tests execute when running: `go test ./catalog/<provider>/<kind>/...`
- [ ] All tests pass (no failures)
- [ ] Tests validate all buf.validate rules are correct

**Critical:** Kinds with failing tests are considered incomplete. Test execution is mandatory for production-readiness.

### 4. IaC Modules - Pulumi (Critical - 13.32%)

**Module Files:**
- [ ] `iac/pulumi/module/main.go` exists
- [ ] `iac/pulumi/module/locals.go` exists
- [ ] `iac/pulumi/module/outputs.go` exists

**Entrypoint Files:**
- [ ] `iac/pulumi/main.go` exists (root `package main`)
- [ ] `iac/pulumi/Pulumi.yaml` exists
- [ ] NO `iac/pulumi/Makefile` (nothing consumes it; the anatomy gate forbids it)

### 5. IaC Modules - Terraform (Critical - 4.44%)

- [ ] `iac/tf/variables.tf` exists (>1KB)
- [ ] `iac/tf/provider.tf` exists
- [ ] `iac/tf/locals.tf` exists
- [ ] `iac/tf/main.tf` exists (>1KB)
- [ ] `iac/tf/outputs.tf` exists

### 6. Documentation - Authored Judgment (Important - 13.34%)

Research documents are never committed -- durable judgment lives in the kind-root
`GUIDE.md` (optional), backed by `README.md` and `catalog.md`:

- [ ] `GUIDE.md` exists where the kind carries non-obvious judgment
- [ ] States the pinned provider version, parity accounting, and recorded exclusions with reasons
- [ ] Teaches judgment, not feature lists; every claim grounded in source

### 7. Documentation - User-Facing (Important - 13.33%)

- [ ] `README.md` at the kind root is substantial (>2KB)
- [ ] `catalog.md` at the kind root follows the catalog page standard

### 8. Supporting Files (Important - 13.33%)

**Pulumi:**
- [ ] `iac/pulumi/README.md` documents the module usefully

**Terraform:**
- [ ] `iac/tf/README.md` documents the module usefully

**Helpers:**
- [ ] `e2e/manifest.yaml` exists at the kind root

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

**Kind A:**
- Critical items: 38% / 40% (missing 1 test)
- Important items: 35% / 40% (missing research doc)
- Nice-to-have: 15% / 20% (missing some polish)
- **Total: 88% - Functionally Complete**

**Kind B:**
- Critical items: 40% / 40% (all present)
- Important items: 40% / 40% (all present)
- Nice-to-have: 20% / 20% (all present)
- **Total: 100% - Fully Complete**

## Audit Report Structure

### Report Header

```markdown
# Audit Report: CloudflareD1Database

**Audit Date:** 2025-11-13 14:30:22
**kind:** CloudflareD1Database
**Provider:** cloudflare
**Kind Path:** `catalog/cloudflare/cloudflared1database/`
**Enum Value:** 7005
**ID Prefix:** cfd1db
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
| Catalog Kind Registry | 4.44% | 4.44% | ✅ |
| Anatomy Conformance | 4.44% | 4.44% | ✅ |
| Protobuf API Definitions | 17.76% | 17.76% | ✅ |
| IaC Modules - Pulumi | 13.32% | 13.32% | ✅ |
| IaC Modules - Terraform | 4.44% | 0.00% | ❌ |
| Documentation - Authored Judgment | 13.34% | 0.00% | ❌ |
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
   - Run forge rules 012-013
   - Creates complete Terraform implementation
   - 15-20 minutes

2. **Write Kind Guide** - Add 13.34%
   - Run forge rule 019 (or `_rules/docs/write-catalog-kind-guide.mdc`)
   - Creates the kind-root GUIDE.md
   - 10-15 minutes

**Total Quick Win Potential: +17.78%**
```

### Critical Gaps

```markdown
## Critical Gaps

Blocking issues that prevent production readiness:

1. **Missing Terraform Module** - 4.44% missing
   - **Why it matters:** Users need choice between Pulumi and Terraform
   - **What to do:** Run `@update-catalog-kind CloudflareD1Database --scenario fill-gaps`
   - **Forge rules:** 012-013

2. **Missing Kind Guide** - 13.34% missing
   - **Why it matters:** Agents and engineers composing architectures need the kind's authored judgment
   - **What to do:** Run forge rule 019 to write the kind-root GUIDE.md
   - **Expected outcome:** Grounded judgment: when to use it, conventions, gotchas, pairings
```

### Detailed Findings (Per Category)

```markdown
## Detailed Findings

### 1. Catalog Kind Registry (4.44%)

✅ **Passed:**
- Enum entry exists: CloudflareD1Database = 7005
- Enum value in correct range (7000-7999 for Cloudflare)
- Unique id_prefix: "cfd1db"
- Complete metadata (provider, version, id_prefix)

**Score:** 4.44% / 4.44% ✅

---

### 2. Anatomy Conformance (4.44%)

✅ **Passed:**
- `go test ./pkg/anatomy/...` reports no violation for catalog/cloudflare/cloudflared1database/
- No baseline.yaml entries for this kind

**Score:** 4.44% / 4.44% ✅

---

### 3. Protobuf API Definitions (17.76%)

✅ **Passed:**
- api.proto substantial (1.2 KB) ✅
- spec.proto substantial (2.5 KB) ✅
- input.proto carries CloudflareD1DatabaseIacInput (800 bytes) ✅
- outputs.proto carries CloudflareD1DatabaseOutputs (600 bytes) ✅
- All .pb.go stubs current (4 files) ✅
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
  - No Makefile (gate-enforced absence) ✅

**Score:** 13.32% / 13.32% ✅

---

### 5. IaC Modules - Terraform (4.44%)

❌ **Failed:**
- iac/tf/ directory does not exist
- No Terraform implementation found

**Score:** 0.00% / 4.44% ❌

**Fix:** Run `@update-catalog-kind CloudflareD1Database --scenario fill-gaps`

---

### 6. Documentation - Authored Judgment (13.34%)

❌ **Failed:**
- GUIDE.md does not exist despite non-obvious operational judgment (state-ownership flags)

**Score:** 0.00% / 13.34% ❌

**Fix:** Run forge rule 019 to write the kind-root GUIDE.md

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
   - **How:** `@update-catalog-kind CloudflareD1Database --scenario fill-gaps`
   - **Impact:** +4.44% (65% → 69.44%)

2. **Write Kind Guide**
   - **File:** `GUIDE.md` (kind root)
   - **Why:** Essential for understanding design decisions
   - **How:** Run forge rule 019
   - **Impact:** +13.34% (69.44% → 82.78%)

### Medium Priority (Do Next)

3. **Expand Presets**
   - **File:** `presets/` (kind root)
   - **Why:** Only 1 preset, distinct use cases warrant 2-3
   - **How:** `@create-planton-preset` for each missing pattern
   - **Impact:** +3.33% (87.78% → 91.11%)

### Low Priority (Polish)

5. **Improve Terraform README**
   - **File:** `iac/tf/README.md`
   - **Why:** Usage documentation for Terraform users
   - **How:** Generate with forge rule 013
   - **Impact:** +3.33% (91.11% → 94.44%)
```

### Comparison

```markdown
## Comparison to Complete Kinds

**Most Similar Complete Kind:** GcpCertManagerCert (98% complete)

**What it has that CloudflareD1Database lacks:**
- Complete Terraform module (variables.tf, main.tf, outputs.tf, etc.)
- A grounded kind GUIDE.md
- Multiple presets (distinct use cases)
- Complete IaC documentation

**Path to Reference:** `catalog/gcp/gcpcertmanagercert/`

**Recommendation:** Review GcpCertManagerCert as a template for completeness.
```

### Next Steps

```markdown
## Next Steps

1. Address critical gaps (Terraform + guide)
2. Run update to fill gaps:
   ```
   @update-catalog-kind CloudflareD1Database --scenario fill-gaps
   ```
3. Re-run audit to verify improvements:
   ```
   @audit-catalog-kind CloudflareD1Database
   ```
4. Expected result: 95-100% complete

**Estimated time to 100%:** 30-45 minutes
```

## Report Delivery

The report is delivered in the session -- kind directories carry no audit artifacts
(the `pkg/anatomy` gate keeps the kind file set closed). To track progress across
sessions, keep a ledger in a shared issue or a location outside the catalog.

**Benefits:**
- **Historical tracking** - Ledger rows show improvement over time
- **Comparison** - Compare audits to measure progress
- **Quality gates** - Validate before releases

## Usage Examples

### Example 1: Check New Kind

```bash
# After forge
@forge-catalog-kind NewKind --provider gcp

# Verify completeness
@audit-catalog-kind NewKind

# Expected: 95-100% complete
# If lower, identifies what's missing
```

### Example 2: Find Gaps to Fill

```bash
# Audit existing kind
@audit-catalog-kind CloudflareD1Database
# Result: 65% complete (missing Terraform, docs)

# Fill gaps
@update-catalog-kind CloudflareD1Database --scenario fill-gaps

# Verify improvement
@audit-catalog-kind CloudflareD1Database
# Result: 98% complete
```

### Example 3: Quality Gate

```bash
# Before committing
@audit-catalog-kind ModifiedKind

# If score decreased:
# - Investigate what was lost
# - Fix before committing
# - Re-audit

# If score same or improved:
# - Safe to commit
```

### Example 4: Batch Audit

```bash
# Audit all kinds
@audit-all-kinds --output-summary

# Output:
# 45 kinds audited
# Average score: 82%
# 12 kinds need attention (<80%)
# 33 kinds production-ready (≥80%)
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
- Major kinds missing

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
@audit-catalog-kind CloudflareD1Database
# Result: 65%

# 2. Fill gaps
@update-catalog-kind CloudflareD1Database --scenario fill-gaps

# 3. Verify improvement
@audit-catalog-kind CloudflareD1Database
# Result: 98%
```

### Forge → Audit

```bash
# 1. Create kind
@forge-catalog-kind NewKind --provider aws

# 2. Validate
@audit-catalog-kind NewKind
# Result: Should be 95-100%
```

### Audit → Delete

```bash
# 1. Check if worth keeping
@audit-catalog-kind OldKind
# Result: 35% (very incomplete)

# 2. Decision: Not worth fixing
@delete-catalog-kind OldKind --backup
```

## Best Practices

### When to Run Audit

- ✅ **After forge** - Validate creation
- ✅ **After update** - Confirm improvements
- ✅ **Before committing** - Quality gate
- ✅ **Weekly/monthly** - Regular health checks
- ✅ **Before release** - Final validation
- ✅ **When onboarding** - Understand kind state

### Interpreting Scores

- **100%** = Perfect, no action needed
- **95-99%** = Excellent, minor polish possible
- **80-94%** = Good, some improvements recommended
- **60-79%** = Fair, significant work needed
- **<60%** = Poor, major work or reconsider

### Using Reports

- Read quick wins first (easy improvements)
- Address critical gaps before medium priority
- Compare to similar complete kinds
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

- **Missing Terraform?** +4.44% (run forge rules 012-013)
- **Missing guide?** +13.34% (run forge rule 019)
- **Thin catalog page?** +6.66% (run forge rule 020, or upgrade the page per `_rules/docs/write-catalog-kind-catalog-md.mdc`)

## Troubleshooting

### Audit Shows 0% but Kind Exists

**Check:**
1. Kind path correct?
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
1. Kind name spelled correctly
2. Kind registered in catalog_kind.proto
3. Folder structure matches conventions

## Success Metrics

Good audit outcomes:

- ✅ Clear completion percentage
- ✅ Specific gaps identified
- ✅ Actionable recommendations
- ✅ Progress recorded in the ledger (outside the catalog)
- ✅ Path to 100% clear

## Related Commands

- `@forge-catalog-kind` - Create new kind
- `@update-catalog-kind` - Fill gaps, enhance kind
- `@complete-catalog-kind` - Auto-improve to 95%+ (audit + update + audit)
- `@fix-catalog-kind` - Targeted fixes with cascading updates
- `@delete-catalog-kind` - Remove kind

## Reference

- **Ideal State Definition:** `architecture/catalog-kind.md`
- **Audit Rule:** `_rules/catalog-kind/audit/audit-catalog-kind.mdc`

---

**Ready to audit?** Run `@audit-catalog-kind <KindName>` to generate a comprehensive report!
