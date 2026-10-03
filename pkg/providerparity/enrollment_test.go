//go:build !codegen
// +build !codegen

package providerparity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/plantonhq/planton/shared/catalogkind"
)

// writeParityPage writes a minimal committed parity page carrying the
// embedded generation parameters DiscoverEnrollments machine-reads.
func writeParityPage(t *testing.T, repoRoot, providerDir, providerName, gaSchema string) {
	t.Helper()
	dir := filepath.Join(repoRoot, catalogRoot, providerDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	page := "# Terraform parity\n\n" +
		"     parameters: provider=" + providerName + " ga-schema=" + gaSchema + "\n"
	if err := os.WriteFile(filepath.Join(dir, PublicReportFileName), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Enrollment is file presence: committing a provider's parity page enrolls
// it, and the provider-to-GA-schema pairing is read from the page itself.
func TestDiscoverEnrollmentsReadsCommittedPages(t *testing.T) {
	root := t.TempDir()
	writeParityPage(t, root, "gcp", "gcp", "google")
	writeParityPage(t, root, "aws", "aws", "aws")

	enrollments, err := DiscoverEnrollments(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []Enrollment{
		{Provider: catalogkind.CatalogProvider_aws, GASchema: "aws"},
		{Provider: catalogkind.CatalogProvider_gcp, GASchema: "google"},
	}
	if len(enrollments) != len(want) {
		t.Fatalf("discovered %d enrollments, want %d", len(enrollments), len(want))
	}
	for i, w := range want {
		if enrollments[i] != w {
			t.Errorf("enrollment[%d] = %+v, want %+v", i, enrollments[i], w)
		}
	}
}

// A page whose embedded provider disagrees with its directory is a hard
// error -- silently trusting either side would let one committed page
// misdirect the shared baseline's gate and write scope.
func TestDiscoverEnrollmentsRejectsDirectoryMismatch(t *testing.T) {
	root := t.TempDir()
	writeParityPage(t, root, "gcp", "aws", "aws")

	if _, err := DiscoverEnrollments(root); err == nil {
		t.Fatal("expected an error for a page whose provider does not match its directory")
	}
}

// A committed page without the machine-readable parameters line must fail
// loudly, never be skipped.
func TestDiscoverEnrollmentsRejectsUnreadableParams(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, catalogRoot, "gcp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, PublicReportFileName), []byte("# hand-authored\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := DiscoverEnrollments(root); err == nil {
		t.Fatal("expected an error for a page without embedded generation parameters")
	}
}

// Zero enrollments is an error, not an empty accounting -- a vacuously
// passing gate is the failure mode this package exists to prevent.
func TestEnrolledAccountingsRejectsEmptyEnrollment(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, catalogRoot), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := EnrolledAccountings(root, map[string]*Schema{}); err == nil {
		t.Fatal("expected an error when no parity pages are committed")
	}
}

// A baseline write consumes the MERGED findings of every enrollment; losing
// any provider's findings in the merge would truncate the shared baseline
// exactly the way the single-provider write path used to.
func TestMergeFindingsPreservesEveryProvider(t *testing.T) {
	accountings := []Accounting{
		{CloudProvider: "gcp", Findings: []Finding{
			{BaselineKey: "kind:GcpGcsBucket", Detail: "a"},
			{BaselineKey: "resource:google_widget", Detail: "b"},
		}},
		{CloudProvider: "aws", Findings: []Finding{
			{BaselineKey: "kind:AwsS3Bucket", Detail: "c"},
		}},
	}
	merged := MergeFindings(accountings)
	if len(merged) != 3 {
		t.Fatalf("merged %d findings, want 3", len(merged))
	}
	keys := map[string]bool{}
	for _, f := range merged {
		keys[f.BaselineKey] = true
	}
	for _, want := range []string{"kind:GcpGcsBucket", "resource:google_widget", "kind:AwsS3Bucket"} {
		if !keys[want] {
			t.Errorf("merged findings lost %s", want)
		}
	}
}

// The shared baseline carries every enrolled provider's entries, so gating
// a SINGLE provider's findings against it misreports the other providers'
// entries as stale -- the exact defect the CLI's --check carried before it
// was routed through EnrolledAccountings + MergeFindings. This pins the
// semantics that make the merged path mandatory for every gate caller.
func TestGateRequiresMergedFindings(t *testing.T) {
	gcp := Accounting{CloudProvider: "gcp", Findings: []Finding{
		{BaselineKey: "kind:GcpGcsBucket", Detail: "open depth gap"},
	}}
	aws := Accounting{CloudProvider: "aws", Findings: []Finding{
		{BaselineKey: "kind:AwsS3Bucket", Detail: "open depth gap"},
	}}
	baseline := map[string]bool{"kind:AwsS3Bucket": true, "kind:GcpGcsBucket": true}

	if res := Gate(aws.Findings, baseline); len(res.StaleEntries) == 0 {
		t.Fatal("single-provider findings should misreport the other provider's baseline entries as stale")
	}
	if res := Gate(MergeFindings([]Accounting{gcp, aws}), baseline); !res.OK() {
		t.Fatalf("merged findings should gate cleanly, got new=%v stale=%v", res.NewFindings, res.StaleEntries)
	}
}
