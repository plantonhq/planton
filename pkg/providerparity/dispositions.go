//go:build !codegen
// +build !codegen

// The dispositions ledger: recorded breadth judgment over the GA provider
// surface. Three disposition classes are COMPUTED and never appear here --
// modeled (the module census proves consumption), and iam-covered and
// iam-uncovered (the *_iam_* pattern, split by whether a module consumes a
// sibling form) -- plus schema-flagged deprecations. The ledger records
// only what requires judgment: composed (covered by an existing kind's
// fields), model-planned (covered by a planned kind not built yet -- its
// own or one it composes into), deferred (with the reason), and doc-level
// deprecations the schema flag misses.
//
// One ledger file per GA schema (dispositions/<schema>.yaml), loaded from
// the repo tree like the schema artifacts. A missing file is an empty
// ledger, not an error: every un-dispositioned resource is a Finding, and
// the burn-down baseline carries the not-yet-judged remainder loudly.

package providerparity

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// DefaultDispositionsDir is repo-root-relative, beside the schema artifacts.
const DefaultDispositionsDir = "pkg/providerparity/dispositions"

// LedgerEntry is one recorded resource disposition.
type LedgerEntry struct {
	Resource    string `yaml:"resource"`
	Disposition string `yaml:"disposition"`
	// Reason is mandatory: an unexplained disposition is an omission with
	// extra steps.
	Reason string `yaml:"reason"`
}

type ledgerDoc struct {
	// Provider names the schema the ledger judges, e.g. "google" -- kept in
	// the content (the schema-artifact identity convention) so a ledger can
	// never be applied against the wrong provider.
	Provider  string        `yaml:"provider"`
	Resources []LedgerEntry `yaml:"resources"`
}

// LoadLedger reads one GA schema's dispositions ledger. A missing file is
// an empty ledger (the baseline carries the un-judged remainder); a present
// file is validated strictly -- it is authored judgment, so a malformed
// entry stops the run rather than accounting against a half-read record.
func LoadLedger(path, gaSchema string) ([]LedgerEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "opening dispositions ledger %s", path)
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	var doc ledgerDoc
	if err := dec.Decode(&doc); err != nil {
		return nil, errors.Wrapf(err, "dispositions ledger %s is not valid (strict) YAML", path)
	}
	if doc.Provider != gaSchema {
		return nil, errors.Errorf("dispositions ledger %s judges provider %q, not %q", path, doc.Provider, gaSchema)
	}
	seen := map[string]bool{}
	for _, e := range doc.Resources {
		if e.Resource == "" {
			return nil, errors.Errorf("dispositions ledger %s: entry without a resource", path)
		}
		if seen[e.Resource] {
			return nil, errors.Errorf("dispositions ledger %s: %s is dispositioned twice -- exactly one disposition per resource", path, e.Resource)
		}
		seen[e.Resource] = true
		if !ledgerDispositions[e.Disposition] {
			return nil, errors.Errorf(
				"dispositions ledger %s: %s carries disposition %q -- the ledger records composed/model-planned/deferred/excluded-deprecated (modeled, iam-covered, and iam-uncovered are computed)",
				path, e.Resource, e.Disposition)
		}
		if e.Reason == "" {
			return nil, errors.Errorf("dispositions ledger %s: %s carries no reason", path, e.Resource)
		}
	}
	return doc.Resources, nil
}
