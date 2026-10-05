package catalogpage

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// repoRoot resolves the repository root from this file's location so the
// gate works from any test working directory (including the Bazel sandbox,
// where the catalog source tree is absent -- callers skip there).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve caller location")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// TestCatalogPageGate is the CI guardrail: the live walk over every catalog
// page must not introduce a violation outside the baseline or leave a stale
// baseline entry. On failure, either bring the page to the standard (the
// detail says how) or -- for a page whose upgrade is deliberately routed to
// its provider's sweep batch -- regenerate the baseline with
// PLANTON_REGEN_CATALOG_PAGE_BASELINE=1 and justify the growth in review.
func TestCatalogPageGate(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.catalog-page lane")
	}

	violations, err := Check(root)
	if err != nil {
		t.Fatalf("catalog page walk: %v", err)
	}

	_, thisFile, _, _ := runtime.Caller(0)
	baselinePath := filepath.Join(filepath.Dir(thisFile), "baseline.yaml")
	if os.Getenv("PLANTON_REGEN_CATALOG_PAGE_BASELINE") == "1" {
		if err := WriteBaseline(baselinePath, violations); err != nil {
			t.Fatalf("write baseline: %v", err)
		}
		t.Logf("baseline regenerated -- review the diff before committing")
		return
	}

	baseline, err := LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("load baseline: %v", err)
	}
	res := Gate(violations, baseline)
	for _, v := range res.NewViolations {
		t.Errorf("catalog page below the bar: %s -- %s", v.ID(), v.Detail)
	}
	for _, id := range res.StaleEntries {
		t.Errorf("stale baseline entry (no longer a violation): %s -- remove it from baseline.yaml", id)
	}
}

// conformingPage is a synthetic page that satisfies every structural rule:
// the hermetic green against which each red below is one deliberate defect.
// It carries no complete manifest (manifest validity needs the real kind
// registry; the invalid side is proven separately below).
const conformingPage = `# Acme Widget

Deploys an Acme widget with sensible defaults. More detail follows here.

## What Gets Created

- **Widget** - the widget.

## Before You Deploy

### Planton Setup

Connect things.

### Acme Account

Have an account.

## Deploy

### Console

Open the deployment store.

### CLI

Create a manifest and apply it.

### InfraChart

Wire the reference.

## Key Configuration

These are the most important decisions.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Acme Zone | spec.zoneId | status.outputs.zone_id |

### What This Kind Provides

| Output | Description | Common Downstream Use |
|---|---|---|
| widget_id | The widget id. | Downstream wiring. |

## Common Patterns

Use it well.

## Works With

- [**Acme Zone**](/infra-catalog/acme-zone) - composes.
`

// TestCheckPage_HermeticFixture proves every rule fires against synthetic
// pages that each break exactly one thing -- the gate's own deliberate red.
// A gate that cannot fail teaches false confidence.
func TestCheckPage_HermeticFixture(t *testing.T) {
	rulesOf := func(page string) map[string]bool {
		got := map[string]bool{}
		for _, v := range CheckPage("catalog/x/y/catalog.md", []byte(page)) {
			got[v.Rule] = true
		}
		return got
	}

	if got := rulesOf(conformingPage); len(got) != 0 {
		t.Errorf("the conforming page must pass every rule, got %v", got)
	}

	// The H1 is not the first heading.
	if got := rulesOf(strings.Replace(conformingPage, "# Acme Widget\n", "## Stray\n\nProse here.\n\n# Acme Widget\n", 1)); !got[RuleMissingH1] {
		t.Errorf("expected %s, got %v", RuleMissingH1, got)
	}

	// Structure arrives before any intro prose.
	if got := rulesOf(strings.Replace(conformingPage, "\nDeploys an Acme widget with sensible defaults. More detail follows here.\n", "", 1)); !got[RuleMissingIntro] {
		t.Errorf("expected %s, got %v", RuleMissingIntro, got)
	}

	// A hard-wrapped intro: the first physical line is a mid-sentence
	// fragment -- exactly what the bundle would ship as the description.
	if got := rulesOf(strings.Replace(conformingPage,
		"Deploys an Acme widget with sensible defaults. More detail follows here.",
		"Deploys an Acme widget with sensible\ndefaults and more detail.", 1)); !got[RuleIntroFragment] {
		t.Errorf("expected %s, got %v", RuleIntroFragment, got)
	}

	// The retired old-standard skeleton is nonstandard structure, and the
	// finer checks stay quiet behind that one verdict.
	oldA := "# Acme Widget\n\nDeploys an Acme widget with sensible defaults.\n\n" +
		"## What Gets Created\n\n## Prerequisites\n\n## Quick Start\n\n## Configuration Reference\n\n## Outputs\n\n## Related Kinds\n"
	if got := rulesOf(oldA); !got[RuleNonstandardStructure] || got[RuleMissingAnchor] || got[RuleInfraChartArm] {
		t.Errorf("expected only %s from the old skeleton, got %v", RuleNonstandardStructure, got)
	}

	// Right H2 skeleton, one fixed H3 anchor missing.
	if got := rulesOf(strings.Replace(conformingPage, "### CLI\n\nCreate a manifest and apply it.\n\n", "", 1)); !got[RuleMissingAnchor] {
		t.Errorf("expected %s, got %v", RuleMissingAnchor, got)
	}

	// The InfraChart law, both directions.
	if got := rulesOf(strings.Replace(conformingPage, "### InfraChart\n\nWire the reference.\n\n", "", 1)); !got[RuleInfraChartArm] {
		t.Errorf("expected %s when Consumes has rows and the arm is absent, got %v", RuleInfraChartArm, got)
	}
	noRows := strings.Replace(conformingPage,
		"| Dependency | Field | ValueFromRef Path |\n|---|---|---|\n| Acme Zone | spec.zoneId | status.outputs.zone_id |",
		"This kind has no foreign key dependencies.", 1)
	if got := rulesOf(noRows); !got[RuleInfraChartArm] {
		t.Errorf("expected %s when Consumes is empty and the arm is present, got %v", RuleInfraChartArm, got)
	}

	// A complete manifest that cannot validate (unknown kind) is a lie the
	// page teaches; fenced yaml comments must not parse as headings.
	withManifest := strings.Replace(conformingPage, "Create a manifest and apply it.\n",
		"Create a manifest and apply it:\n\n```yaml\n# a comment, not a heading\napiVersion: acme.planton.dev/v1alpha1\nkind: AcmeNotARealKind\nmetadata:\n  name: w\nspec: {}\n```\n", 1)
	if got := rulesOf(withManifest); !got[RuleInvalidManifest] || got[RuleMissingH1] {
		t.Errorf("expected %s (and no heading misparse), got %v", RuleInvalidManifest, got)
	}
}

// TestGate mirrors the anatomy gate semantics: new drift fails, baselined
// drift passes, a fixed entry left in the baseline fails as stale.
func TestGate(t *testing.T) {
	v := Violation{Path: "catalog/aws/x/catalog.md", Rule: RuleNonstandardStructure}
	if res := Gate([]Violation{v}, map[string]bool{}); res.OK() || len(res.NewViolations) != 1 {
		t.Errorf("expected new drift to be detected, got %+v", res)
	}
	if res := Gate([]Violation{v}, map[string]bool{v.ID(): true}); !res.OK() {
		t.Errorf("expected baselined drift to pass, got %+v", res)
	}
	if res := Gate(nil, map[string]bool{"catalog/aws/gone/catalog.md:missing-h1": true}); res.OK() || len(res.StaleEntries) != 1 {
		t.Errorf("expected a stale entry to be detected, got %+v", res)
	}
	// One rule firing twice on one page collapses to one baseline id.
	if res := Gate([]Violation{v, v}, map[string]bool{v.ID(): true}); !res.OK() {
		t.Errorf("expected duplicate-rule collapse, got %+v", res)
	}
}
