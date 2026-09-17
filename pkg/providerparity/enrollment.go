//go:build !codegen
// +build !codegen

// Enrollment: which catalog providers are under provider-parity accounting,
// and against which GA schema each one is measured. Enrollment is FILE
// PRESENCE: a provider is enrolled exactly when its generated parity page
// (catalog/<provider>/terraform-parity.md) is committed, and the
// provider-to-GA-schema pairing is read from the page's embedded generation
// parameters -- the same machine-read line the public-report drift gate
// already trusts. The pairing stays explicit and reviewed (a page is only
// ever produced by --write-report with explicit --provider/--ga-schema
// flags), but it lives in ONE committed artifact instead of a code table
// that would drift from the pages.
//
// Onboarding a provider is therefore one reviewed change: generate its page
// with explicit flags (generation never consults the enrollment set, so no
// chicken-and-egg) and commit it alongside the schema artifact and the
// dispositions ledger; the CI gate, every baseline write, and the drift
// gate all follow from the committed page.
//
// The enrollment set is also what makes the shared baseline SAFE to touch:
// baseline.yaml is one file for every provider, so any gate or write must
// consume ALL enrolled providers' findings. EnrolledAccountings +
// MergeFindings are the single input path for the CI gate test, its
// regeneration mode, and the CLI's --check and --write-baseline -- one
// source of truth, so no caller can gate the shared baseline against (or
// truncate it to) a single provider's findings.

package providerparity

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// Enrollment binds one catalog provider to the GA schema its parity is
// declared against.
type Enrollment struct {
	Provider cloudresourcekind.CloudResourceProvider
	// GASchema names the committed schema artifact (schemas/<name>-*.json.gz)
	// that is this provider's parity baseline.
	GASchema string
}

// DiscoverEnrollments reads the enrolled provider set from the committed
// parity pages under <repoRoot>/catalog/*/terraform-parity.md. Order is
// stable (lexical by page path). A page whose parameters are unreadable or
// whose embedded provider does not match its directory is a hard error --
// a malformed page must fail loudly, never silently narrow the baseline's
// gate and write scope.
func DiscoverEnrollments(repoRoot string) ([]Enrollment, error) {
	pages, err := filepath.Glob(filepath.Join(repoRoot, catalogRoot, "*", PublicReportFileName))
	if err != nil {
		return nil, errors.Wrap(err, "globbing committed parity pages")
	}
	sort.Strings(pages)
	enrollments := make([]Enrollment, 0, len(pages))
	for _, page := range pages {
		raw, err := os.ReadFile(page)
		if err != nil {
			return nil, errors.Wrapf(err, "reading %s", page)
		}
		providerName, gaSchema, err := ParseReportParams(string(raw))
		if err != nil {
			return nil, errors.Wrapf(err, "%s", page)
		}
		provider, err := providerFromName(providerName)
		if err != nil {
			return nil, errors.Wrapf(err, "%s", page)
		}
		if dir := filepath.Base(filepath.Dir(page)); dir != crkreflect.ProviderDirName(provider) {
			return nil, errors.Errorf("%s: embedded provider %q does not match its directory %q", page, providerName, dir)
		}
		enrollments = append(enrollments, Enrollment{Provider: provider, GASchema: gaSchema})
	}
	return enrollments, nil
}

// EnrolledAccountings runs the full accounting for every enrolled provider.
// It is the input side of every baseline gate and write: gating or writing
// from any subset would report the other providers' baseline entries as
// stale (the gate) or silently drop them (the write). Zero enrollments is
// an error, not an empty result -- a vacuously passing gate is the failure
// mode this package exists to prevent.
func EnrolledAccountings(repoRoot string, schemas map[string]*Schema) ([]Accounting, error) {
	enrollments, err := DiscoverEnrollments(repoRoot)
	if err != nil {
		return nil, err
	}
	if len(enrollments) == 0 {
		return nil, errors.New("no committed parity pages found -- the gate would vacuously pass (run from the repository root)")
	}
	accountings := make([]Accounting, 0, len(enrollments))
	for _, e := range enrollments {
		acc, err := BuildAccounting(repoRoot, e.Provider, schemas, e.GASchema, "", "")
		if err != nil {
			return nil, errors.Wrapf(err, "accounting %s catalog against GA schema %q", e.Provider, e.GASchema)
		}
		accountings = append(accountings, acc)
	}
	return accountings, nil
}

// MergeFindings flattens the enrolled accountings' findings into the one
// slice WriteBaseline and Gate consume. Trivial on purpose: the value is
// that every baseline write and gate check goes through the same union.
func MergeFindings(accountings []Accounting) []Finding {
	var all []Finding
	for _, acc := range accountings {
		all = append(all, acc.Findings...)
	}
	return all
}
