# Complete: One-Command Quality Improvement

## Overview

**Complete** is the convenience operation that combines audit and update into a single command: assess a kind's completeness and automatically fill all gaps to reach the target score (default 95%).

**Philosophy:** Maximum automation, minimum friction.

## Why Complete Exists

While audit and update are powerful when used separately, most of the time you want a simple answer to: **"Make this kind production-ready."**

**Complete gives you that one-command workflow.**

## The Problem Complete Solves

### Manual Workflow (3 Commands)

```bash
# 1. Check status
@audit-catalog-kind CloudflareD1Database
# Output: 65% complete, missing Terraform, docs, examples

# 2. Read report, decide what to do

# 3. Fill gaps
@update-catalog-kind CloudflareD1Database --scenario fill-gaps
# Wait 15-20 minutes

# 4. Verify
@audit-catalog-kind CloudflareD1Database
# Output: 98% complete

# Time: 20+ minutes + manual steps
```

### Complete Workflow (1 Command)

```bash
@complete-catalog-kind CloudflareD1Database

# Automatically:
# - Audits (65% complete)
# - Fills gaps (Terraform, docs, examples)
# - Re-audits (98% complete)
# - Reports improvement (+33%)

# Time: 18 minutes, zero manual steps
```

**Time Savings:** Eliminates manual coordination overhead

## When to Use Complete

### ✅ Use Complete When

- **Quick improvement needed** - Want kind production-ready ASAP
- **Batch processing** - Improving multiple kinds
- **Trust automation** - Confident in automatic gap-filling
- **Quality gates** - Ensuring standards before release
- **Onboarding legacy** - Systematically completing old kinds
- **Time-constrained** - Want results without manual intervention

### ❌ Use Manual Audit/Update When

- **Selective improvements** - Only want to fill specific gaps
- **Custom updates** - Need update scenarios other than fill-gaps
- **Review before changes** - Want to see what will change
- **Learning mode** - Understanding the system step-by-step
- **Complex situations** - Edge cases requiring manual intervention

## How It Works

### The Three-Step Process

```
Step 1: Audit
   ↓
   Assess current state
   Calculate score
   Identify all gaps
   ↓
Step 2: Fill Gaps (if score < target)
   ↓
   Run update --fill-gaps
   Execute forge rules for missing items
   Validate after each creation
   ↓
Step 3: Verify
   ↓
   Re-audit to measure improvement
   Compare before/after
   Generate summary report
```

### What Gets Filled

Based on audit results, complete will create:

**Critical Items (if missing):**
- Terraform module (variables.tf, main.tf, outputs.tf, etc.)
- Pulumi module files (if somehow missing)
- Proto files (api.proto, spec.proto, input.proto, outputs.proto)
- Generated stubs (.pb.go files)
- Unit tests (spec_test.go)

**Important Items (if missing):**
- Kind guide (GUIDE.md at the kind root)
- User-facing docs (README.md and catalog.md at the kind root)
- Presets (presets/ with .md sidecars, multiple use cases)
- IaC documentation (Pulumi/Terraform READMEs)

**Nice-to-Have Items (if missing and target=100%):**
- Additional examples
- Extra supporting files

### What Complete Won't Do

Complete is specifically for **filling gaps**, not other updates:

❌ Won't modify existing proto schema (use update --proto-changed)
❌ Won't refresh already-existing docs (use update --refresh-docs)
❌ Won't change IaC implementation (use update --update-iac)
❌ Won't fix specific issues (use update --explain)

**Complete only fills missing items, doesn't modify existing ones.**

## Usage

### Basic Usage

```bash
@complete-catalog-kind <KindName>
```

**Behavior:**
1. Audits kind
2. If score <95%, fills gaps
3. Re-audits to verify
4. Reports results

### With Dry-Run

```bash
@complete-catalog-kind CloudflareD1Database --dry-run
```

**Shows:**
- Current audit score
- All gaps that would be filled
- Estimated time and file count
- Expected final score
- **No files modified**

### With Custom Target

```bash
@complete-catalog-kind KubernetesPostgres --target-score 90
```

**Behavior:**
- Stops when 90% reached (instead of default 95%)
- May skip some nice-to-have items
- Faster completion
- Good for "good enough" scenarios

### With Skip Validation

```bash
@complete-catalog-kind QuickKind --skip-validation
```

**Warning:** Faster but riskier (skips final build/test validation)

## Examples

### Example 1: Incomplete Kind

**Scenario:** Legacy kind at 60% completion

```bash
@complete-catalog-kind CloudflareD1Database
```

**Output:**
```
🎯 Complete: CloudflareD1Database

Step 1/3: Initial Audit
  Current: 65%
  Missing: Terraform, kind guide
  
Step 2/3: Filling Gaps (18 minutes)
  ✅ Created Terraform module (7 files)
  ✅ Wrote GUIDE.md (kind guide)
  ✅ Generated Pulumi module README
  ✅ Enhanced examples (+3 examples)
  
Step 3/3: Final Verification
  Final: 98%
  
✅ Success! (+33% improvement)

Before: 65% (Partially Complete)
After: 98% (Functionally Complete)

Duration: 18 minutes
Files Created: 12
```

### Example 2: Already Complete

**Scenario:** Kind already at high score

```bash
@complete-catalog-kind GcpCertManagerCert
```

**Output:**
```
🎯 Complete: GcpCertManagerCert

Step 1/3: Initial Audit
  Current: 98%
  
✅ Kind already complete! (≥95% target)

No gaps to fill. Production-ready! 🎉

Duration: 30 seconds (audit only)
```

### Example 3: Preview Mode

**Scenario:** Want to see what would happen

```bash
@complete-catalog-kind OldKind --dry-run
```

**Output:**
```
🔍 Dry-Run: Complete OldKind

Current State:
  Score: 45%
  Status: Skeleton Exists
  
Gaps Identified (14 items):
  ❌ Missing Terraform module
  ❌ Missing kind guide
  ❌ Missing Pulumi module README
  ❌ Missing examples
  ❌ Incomplete proto definitions
  ... (9 more)

Planned Actions:
  - Run 15 forge flow rules
  - Create ~25 files
  - Generate ~3000 lines of code/docs
  - Estimated duration: 35-45 minutes

Expected Result:
  Score: 95% (+50%)
  
Decision: Kind is very incomplete.
Consider: Is it worth completing or should you delete it?

No changes made (dry-run mode)
```

### Example 4: Target Score 100%

**Scenario:** Want absolute perfection

```bash
@complete-catalog-kind MyKind --target-score 100
```

**Output:**
```
🎯 Complete: MyKind (Target: 100%)

Current: 92%
Missing: Overview docs, additional examples

Filling all gaps including nice-to-haves...

Final: 100% ✅

Duration: 12 minutes (extra time for polish items)
```

### Example 5: Batch Processing

**Scenario:** Complete 10 kinds

```bash
# Complete all SaaS platform kinds
for kind in CloudflareD1Database Auth0Client CloudflareDnsZone; do
  echo "=== Completing $kind ==="
  @complete-catalog-kind $kind
  echo ""
done
```

**Output:**
```
=== Completing CloudflareD1Database ===
✅ 65% → 98% (+33%) in 18 min

=== Completing Auth0Client ===
✅ Already complete (97%)

=== Completing CloudflareDnsZone ===
✅ 70% → 95% (+25%) in 22 min

Summary: 3 kinds processed, 2 improved, 1 already complete
```

## Progress Indicators

### Visual Progress Bar

```
Step 2/3: Filling Gaps

Terraform Module [████████░░] 80% (6/7 files)
Documentation   [██░░░░░░░░] 20% (1/5 sections)
Validation      [░░░░░░░░░░] 0% (pending)

Overall: [███░░░░░░░] 35% complete
```

### Percentage Tracking

```
Initial: 65%
  ↓ +4.44% (Terraform created)
Current: 69.44%
  ↓ +13.34% (GUIDE.md written)
Current: 82.78%
  ↓ +5% (Overview created)
Current: 87.78%
  ↓ Validation...
Final: 98%
```

## Integration Patterns

### With CI/CD

```yaml
# .github/workflows/kind-quality.yml
name: Kind Quality

on: [pull_request]

jobs:
  complete-kinds:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Complete modified kinds
        run: |
          # Detect which kinds changed
          KINDS=$(detect-modified-kinds.sh)
          
          # Complete each
          for kind in $KINDS; do
            @complete-catalog-kind $kind --target-score 80
          done
          
          # Fail if any <80%
          verify-scores.sh --min 80
```

### With Git Workflows

```bash
# Feature branch workflow
git checkout -b feature/improve-kinds

# Complete multiple kinds
@complete-catalog-kind Kind1
@complete-catalog-kind Kind2
@complete-catalog-kind Kind3

# Commit improvements
git add -A
git commit -m "improve: complete Kind1, Kind2, Kind3 to 95%+"

# Create PR
gh pr create --title "Improve kind quality" --body "..."
```

### With Forge

```bash
# Create new kind
@forge-catalog-kind NewKind --provider aws

# Forge might create at 90-95% if some optional items skipped
# Complete to 100%
@complete-catalog-kind NewKind --target-score 100

# Result: Perfect kind
```

## Comparison Table

| Aspect | Manual (Audit + Update) | Complete (Automated) |
|--------|------------------------|----------------------|
| Commands | 3 (audit, update, audit) | 1 (complete) |
| Time | 20+ min + manual steps | 18 min, zero manual |
| Control | Full control over updates | Automatic gap-filling |
| Use Case | Selective improvements | All gaps filled |
| Dry-run | Separate for each command | Single dry-run |
| Reports | 1 audit report | 2 audit reports (before/after) |
| Best For | Custom scenarios | Quick completion |

## Best Practices

### Before Complete

- [ ] Commit current changes (clean git state)
- [ ] Understand kind purpose (read existing docs)
- [ ] Check disk space (doc generation can be large)
- [ ] Allocate time (15-30 minutes typical)
- [ ] Consider dry-run first (preview changes)

### During Complete

- [ ] Monitor progress messages
- [ ] Watch for error messages
- [ ] Don't interrupt (let it complete)
- [ ] Trust the automation (it validates)

### After Complete

- [ ] Review before/after audit reports
- [ ] Check generated files quality
- [ ] Test deployment (use the e2e manifest)
- [ ] Commit with meaningful message
- [ ] Share improvement metrics

## Tips

### Quick Quality Improvement

```bash
# One command to production-ready
@complete-catalog-kind AnyKind

# Typical results:
# 40-60% → 95%+ (30-40 min)
# 60-80% → 95%+ (15-25 min)
# 80-94% → 95%+ (5-15 min)
# 95%+ → Already complete (30 sec)
```

### Batch Improvement Workflow

```bash
# Audit all kinds to find low scores
# (manually or with script)

# Complete only those <80%
@complete-catalog-kind LowScoreKind1
@complete-catalog-kind LowScoreKind2

# Result: All kinds now ≥95%
```

### Target Score Strategy

- **95% (default)** - Functionally complete, production-ready
- **90%** - Good enough for most uses, faster
- **100%** - Absolute perfection, takes longer

Choose based on:
- Time available
- Kind importance
- Release timeline
- Team standards

## Troubleshooting

### Complete Runs But Score Doesn't Improve

**Check:**
1. Are gaps fillable? (Some might need manual work)
2. Did update fail silently?
3. Are validation rules preventing completion?

**Debug:**
```bash
# Run manual update to see detailed errors
@update-catalog-kind KindName --scenario fill-gaps
```

### Complete Takes Too Long

**Typical Times:**
- Small gaps (1-3 items): 5-10 min
- Medium gaps (4-6 items): 15-20 min
- Large gaps (7+ items): 25-35 min

**If exceeds:**
- Check if stuck on specific rule
- Cancel and investigate
- Run manual update to debug

### Build/Test Failures After Complete

**Complete validates, but if it fails:**
```
❌ Validation failed after gap-filling

Build: ✅ Passed
Tests: ❌ Failed (2 tests)

Recommendation:
  1. Check test output
  2. Fix manually
  3. Re-run complete (it's idempotent)
```

## Success Metrics

Good complete outcomes:

- ✅ Score improved by 20-50%
- ✅ Reached target score (≥95%)
- ✅ All critical gaps filled
- ✅ Build and tests pass
- ✅ Two audit reports (historical tracking)
- ✅ Ready to commit

## Use Cases

### Use Case 1: Legacy Kind Onboarding

**Problem:** 50 old kinds at 60-70% completion

**Solution:**
```bash
# Systematically complete all
for kind in $(list-legacy-kinds); do
  @complete-catalog-kind $kind
done

# Result: All at 95%+ in 10-20 hours total
# Manual would take 400-800 hours
```

### Use Case 2: Quality Gate

**Problem:** Want all kinds at 95%+ before v1.0 release

**Solution:**
```bash
# Week before release, complete all
@complete-catalog-kind Kind1
@complete-catalog-kind Kind2
# ... etc

# All kinds now meet release standards
```

### Use Case 3: Forge Follow-Up

**Problem:** Forge created kind at 92% (some failures)

**Solution:**
```bash
# After forge
@complete-catalog-kind NewKind

# Fills remaining 8%
# Guaranteed 95%+ result
```

### Use Case 4: Rapid Response

**Problem:** Customer needs kind ASAP, it's only 70% complete

**Solution:**
```bash
# Quick complete
@complete-catalog-kind UrgentKind

# 15-20 minutes later: 95%+ complete
# Ship to customer
```

## Comparison to Alternatives

### Complete vs Manual Audit + Update

**Manual Workflow:**
- More control
- More steps
- More time (manual coordination)
- Better for selective improvements

**Complete:**
- Less control (automatic)
- One step
- Less time (automated)
- Better for full completion

**Choose based on:** Need for control vs need for speed

### Complete vs Update --fill-gaps

**They're almost the same!**

Difference:
- Complete: Audits before AND after (shows improvement)
- Update: Only shows progress during update

**Complete = Audit + Update --fill-gaps + Audit**

### Complete vs Forge

**Forge:** Creates from scratch (0% → 95%)
**Complete:** Improves existing (X% → 95%)

**Don't use complete to create new kinds** - that's forge's job.

## Advanced Usage

### Custom Target Scores

```bash
# Production kinds (strict)
@complete-catalog-kind ProdKind --target-score 100

# Development kinds (lenient)
@complete-catalog-kind DevKind --target-score 80

# Experimental (minimal)
@complete-catalog-kind ExperimentalKind --target-score 60
```

### Scripted Batch Processing

```bash
#!/bin/bash
# complete-all-saas.sh

KINDS=(CloudflareD1Database Auth0Client CloudflareDnsZone)

for kind in "${KINDS[@]}"; do
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "Completing: $kind"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  
  @complete-catalog-kind $kind
  
  if [ $? -eq 0 ]; then
    echo "✅ $kind completed successfully"
  else
    echo "❌ $kind failed"
  fi
  
  echo ""
done

echo "Batch completion finished!"
```

### With Reporting

```bash
# Complete and capture metrics
BEFORE=$(@audit-catalog-kind CloudflareD1Database --score-only)

@complete-catalog-kind CloudflareD1Database

AFTER=$(@audit-catalog-kind CloudflareD1Database --score-only)

echo "Improvement: $BEFORE% → $AFTER% (+$(($AFTER - $BEFORE))%)"
```

## Troubleshooting

### "Kind already complete"

**Message:** `✅ Kind is already at 98% (target is 95%)`

**Action:** None needed! Kind is production-ready.

**If you want 100%:**
```bash
@complete-catalog-kind KindName --target-score 100
```

### "Partial completion"

**Message:** `⚠️ Reached 88% (target was 95%)`

**Reason:** Some gaps couldn't be filled automatically

**Action:**
```bash
# See audit report for remaining gaps
# Use manual update for specific fixes
@update-catalog-kind KindName --scenario <specific>
```

### "Build failed after completion"

**Message:** `❌ Build validation failed after gap-filling`

**Action:**
```bash
# Check build errors
go build ./catalog/<provider>/<kind>/...

# Fix manually
# Re-run complete (it's idempotent)
@complete-catalog-kind KindName
```

### "Takes too long"

**Expected times:**
- 5-10 min: Small gaps (1-3 items)
- 15-20 min: Medium gaps (4-6 items)
- 25-35 min: Large gaps (7+ items)

**If exceeds 45 minutes:**
- Cancel (Ctrl+C)
- Check what's taking long
- Run manual update to debug

## Success Criteria

After complete finishes:

✅ Score ≥ target (default 95%)
✅ All critical gaps filled
✅ spec_test.go exists with validation tests
✅ Kind tests execute successfully
✅ All tests pass (`go test ./catalog/<provider>/<kind>/...`)
✅ Build validation passed (`go build ./catalog/<provider>/<kind>/...`)
✅ Full test suite passed (`go test -v ./catalog/<provider>/<kind>/...`)
✅ Two audit reports (before/after)
✅ Summary shows improvement
✅ Ready to commit

**Critical:** Test execution is now part of completeness. Kinds with failing tests are considered incomplete even if all files are present.

## Workflow Patterns

### Pattern 1: Quality Sprint

```bash
# Monday: Audit all kinds
# Identify 15 kinds <80%

# Tuesday-Wednesday: Complete all
for kind in LowScoreKinds; do
  @complete-catalog-kind $kind
done

# Thursday: Review and test
# Friday: Commit and release

# Result: All kinds ≥95%
```

### Pattern 2: Pre-Release

```bash
# 1 week before release
# Complete all kinds for release

@complete-catalog-kind Kind1
@complete-catalog-kind Kind2
# ... etc

# All kinds now meet quality standards
```

### Pattern 3: Continuous Improvement

```bash
# Weekly: Complete lowest-scoring kind

# Week 1
@complete-catalog-kind LowestKind1

# Week 2
@complete-catalog-kind LowestKind2

# Over time, all kinds reach 95%+
```

## Related Commands

- `@forge-catalog-kind` - Create new kind
- `@audit-catalog-kind` - Check status only
- `@update-catalog-kind` - Selective improvements
- `@fix-catalog-kind` - Targeted fixes with cascading updates
- `@delete-catalog-kind` - Remove kind

## Reference

- **Ideal State:** `architecture/catalog-kind.md`
- **Complete Rule:** `_rules/catalog-kind/complete/complete-catalog-kind.mdc`
- **Master README:** `_rules/catalog-kind/README.md`

## FAQ

**Q: What's the difference between complete and update --fill-gaps?**

A: Complete = Audit + Update --fill-gaps + Audit. It adds the before/after audit reports and summary.

**Q: Can I complete a kind that doesn't exist?**

A: No. Use `@forge-catalog-kind` to create new kinds.

**Q: What if I only want to fill some gaps, not all?**

A: Use `@update-catalog-kind` with specific scenario instead.

**Q: Is complete safe?**

A: Yes! It uses the same update --fill-gaps logic with all safety features (validation, retry, etc.). Use `--dry-run` to preview.

**Q: Can I undo complete?**

A: Yes, via git (revert commit) or by deleting generated files. Complete creates new files, doesn't modify existing ones (usually).

**Q: How long does complete take?**

A: 5-35 minutes depending on gaps. Dry-run shows estimate.

---

**Ready to complete?** Run `@complete-catalog-kind <KindName>` for one-command quality improvement!

