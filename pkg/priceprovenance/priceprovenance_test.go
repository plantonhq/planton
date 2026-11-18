package priceprovenance

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// repoRoot locates the repository root from this source file's own location,
// so the scan works identically under `go test` (package-dir cwd) and Bazel.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this source file")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
}

// TestPriceProvenanceGate is the CI guardrail. The live scan over the
// catalog's markdown must not carry a dollar token outside the
// baseline's two lists, and neither list may hold a stale entry. On failure:
// rewrite the price to driver-teaching copy (prices live only in
// catalog/_pricing/), or -- ONLY for a user-chosen dollar-typed config
// example -- add an `allowed` entry with its reason. Regenerate the
// burn-down list with PLANTON_REGEN_PRICE_BASELINE=1 go test ./pkg/priceprovenance/
func TestPriceProvenanceGate(t *testing.T) {
	root := repoRoot(t)
	findings, err := Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	baselinePath := filepath.Join(root, "pkg", "priceprovenance", "baseline.yaml")
	if os.Getenv("PLANTON_REGEN_PRICE_BASELINE") == "1" {
		if err := WriteBaseline(baselinePath, findings); err != nil {
			t.Fatalf("write baseline: %v", err)
		}
		t.Logf("baseline regenerated at %s", baselinePath)
		return
	}

	baseline, err := LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("load baseline: %v", err)
	}
	res := Gate(findings, baseline)
	for _, id := range res.NewTokens {
		t.Errorf("hand-typed dollar token outside the baseline: %s -- rewrite to driver-teaching copy (prices live only in catalog/_pricing/), or add an allowed entry with its reason if this is a user-chosen dollar-typed config example", id)
	}
	for _, id := range res.StaleGaps {
		t.Errorf("stale gaps entry (no longer present): %s -- remove it from baseline.yaml", id)
	}
	for _, id := range res.StaleAllowed {
		t.Errorf("stale allowed entry (no longer present): %s -- remove it from baseline.yaml", id)
	}
}

func TestPricePattern(t *testing.T) {
	cases := []struct {
		name string
		line string
		want []string
	}{
		{"hourly rate", "Each SSD node costs ~$0.65/hour (~$468/month).", []string{"$0.65", "$468"}},
		{"thousands separator", "AWS S3 Equivalent: $9,042/month", []string{"$9,042"}},
		{"regex backreference matches too -- allowed-list territory", `replace: "/v2/$1"`, []string{"$1"}},
		{"bare threshold", "alert above a $100 absolute impact", []string{"$100"}},
		{"no dollars", "billed hourly whether idle or not", nil},
		{"dollar sign without digits", "the $VARIABLE shell form", nil},
	}
	for _, tc := range cases {
		got := pricePattern.FindAllString(tc.line, -1)
		if len(got) != len(tc.want) {
			t.Errorf("%s: tokens = %v, want %v", tc.name, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("%s: token[%d] = %q, want %q", tc.name, i, got[i], tc.want[i])
			}
		}
	}
}

func TestGateVerdicts(t *testing.T) {
	finding := Finding{Path: "catalog/aws/x/README.md", Token: "$3"}
	baselineWith := func(allowed, gaps []string) Baseline {
		b := Baseline{Allowed: map[string]bool{}, Gaps: map[string]bool{}}
		for _, id := range allowed {
			b.Allowed[id] = true
		}
		for _, id := range gaps {
			b.Gaps[id] = true
		}
		return b
	}

	if res := Gate([]Finding{finding}, baselineWith(nil, nil)); res.OK() || len(res.NewTokens) != 1 {
		t.Errorf("expected a new token to be detected, got %+v", res)
	}
	if res := Gate([]Finding{finding}, baselineWith(nil, []string{finding.ID()})); !res.OK() {
		t.Errorf("expected a baselined gap to pass, got %+v", res)
	}
	if res := Gate([]Finding{finding}, baselineWith([]string{finding.ID()}, nil)); !res.OK() {
		t.Errorf("expected an allowed token to pass, got %+v", res)
	}
	if res := Gate(nil, baselineWith(nil, []string{"gone.md:$1"})); res.OK() || len(res.StaleGaps) != 1 {
		t.Errorf("expected a stale gap to be detected, got %+v", res)
	}
	if res := Gate(nil, baselineWith([]string{"gone.md:$2"}, nil)); res.OK() || len(res.StaleAllowed) != 1 {
		t.Errorf("expected a stale allowed entry to be detected, got %+v", res)
	}
}

// TestScanExclusions proves the two structural exclusions: catalog/_pricing/
// (the verified-data home where dollars BELONG) and non-markdown files are
// never scanned.
func TestScanExclusions(t *testing.T) {
	root := repoRoot(t)
	findings, err := Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	for _, f := range findings {
		if filepath.ToSlash(f.Path) == "" {
			continue
		}
		if match, _ := filepath.Match("catalog/_pricing/*", f.Path); match || len(f.Path) > 17 && f.Path[:17] == "catalog/_pricing/" {
			t.Errorf("finding inside the excluded verified-data home: %s", f.Path)
		}
		if filepath.Ext(f.Path) != ".md" {
			t.Errorf("finding in a non-markdown file: %s", f.Path)
		}
	}
}
