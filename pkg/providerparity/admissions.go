//go:build !codegen
// +build !codegen

// The admission list: recorded judgment that lets a secondary-channel
// provider resource into a catalog whose parity baseline is the canonical
// (GA) provider. Some providers ship capability in a second channel --
// Google's `google-beta` provider carries pre-GA resources and fields; a
// whole product family can live there for years (Firebase's core
// resources). The doctrine admits such capability surgically, never
// wholesale: per resource, per kind, with the reason and the path by which
// its promotion to GA is tracked, so the day the canonical provider ships
// it the admission becomes a staleness finding and the module moves back to
// the baseline provider.
//
// One admissions file per secondary channel (admissions/<provider>.yaml),
// keyed to the GA schema it supplements, loaded from the repo tree like the
// schema artifacts and the dispositions ledger. A missing directory or file
// is an empty list, not an error: with nothing admitted, every
// secondary-channel resource a module consumes is a finding.
//
// The accounting reads the list in three places: a resource served only by
// a non-GA schema is accounted against that schema exactly when this list
// admits it for the consuming kind (otherwise the consumption is a finding);
// a module's explicit `provider = <channel>` attachment (module_census.go)
// must agree with the list in both directions; and an admitted resource the
// GA schema serves at the pin is a stale admission. The public report
// renders the list so the catalog's readers see every admission and its
// reason.

package providerparity

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// DefaultAdmissionsDir is repo-root-relative, beside the schema artifacts and
// the dispositions ledger.
const DefaultAdmissionsDir = "pkg/providerparity/admissions"

// AdmissionEntry is one recorded admission: one resource, for one kind.
type AdmissionEntry struct {
	// Provider is the secondary channel's schema-artifact identity (the
	// admissions file's provider), e.g. "google-beta". Filled by the loader
	// so every entry carries the channel it belongs to.
	Provider string `yaml:"-"`
	// Resource is the provider resource type, e.g. "google_firebase_project".
	Resource string `yaml:"resource"`
	// Kind is the registry kind whose module attaches the resource through
	// the secondary channel. An admission is per kind: two kinds consuming
	// one beta resource carry two entries, each with its own reason.
	Kind string `yaml:"kind"`
	// Reason is mandatory: why the capability is worth a secondary-channel
	// dependency (the canonical provider does not serve it at the pin).
	Reason string `yaml:"reason"`
	// PromotionTracking names where the resource's promotion to the
	// canonical provider is watched (a provider source path, an upstream
	// issue). Mandatory: an admission without a way to know when it can end
	// is a permanent exception with extra steps.
	PromotionTracking string `yaml:"promotionTracking"`
}

type admissionsDoc struct {
	// Provider names the secondary channel's schema artifact, e.g.
	// "google-beta" -- kept in the content (the schema-artifact identity
	// convention) so a list can never be applied against the wrong channel.
	Provider string `yaml:"provider"`
	// GASchema names the parity baseline this channel supplements, e.g.
	// "google". The accounting for that baseline loads exactly the lists
	// that name it.
	GASchema  string           `yaml:"gaSchema"`
	Resources []AdmissionEntry `yaml:"resources"`
}

// LoadAdmissions reads every admissions file under dir that supplements the
// named GA schema and returns their entries, sorted by resource then kind. A
// missing directory is an empty list; a present file is validated strictly
// -- it is authored judgment, so a malformed entry stops the run rather
// than accounting against a half-read record. Files that supplement a
// different GA schema are skipped (one directory serves every provider).
func LoadAdmissions(dir, gaSchema string) ([]AdmissionEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "reading admissions dir %s", dir)
	}
	var out []AdmissionEntry
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		doc, err := readAdmissionsFile(path)
		if err != nil {
			return nil, err
		}
		if doc.GASchema != gaSchema {
			continue
		}
		for _, e := range doc.Resources {
			e.Provider = doc.Provider
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Resource != out[j].Resource {
			return out[i].Resource < out[j].Resource
		}
		return out[i].Kind < out[j].Kind
	})
	return out, nil
}

// readAdmissionsFile parses and validates one admissions file.
func readAdmissionsFile(path string) (*admissionsDoc, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "opening admissions file %s", path)
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	var doc admissionsDoc
	if err := dec.Decode(&doc); err != nil {
		return nil, errors.Wrapf(err, "admissions file %s is not valid (strict) YAML", path)
	}
	if doc.Provider == "" {
		return nil, errors.Errorf("admissions file %s names no provider (the secondary channel's schema identity)", path)
	}
	if doc.GASchema == "" {
		return nil, errors.Errorf("admissions file %s names no gaSchema (the parity baseline it supplements)", path)
	}
	if doc.Provider == doc.GASchema {
		return nil, errors.Errorf("admissions file %s admits %q into itself -- the admission list is for a secondary channel, never the baseline", path, doc.Provider)
	}
	seen := map[string]bool{}
	for _, e := range doc.Resources {
		if e.Resource == "" {
			return nil, errors.Errorf("admissions file %s: entry without a resource", path)
		}
		if e.Kind == "" {
			return nil, errors.Errorf("admissions file %s: %s carries no kind -- an admission is per consuming kind", path, e.Resource)
		}
		key := e.Resource + "/" + e.Kind
		if seen[key] {
			return nil, errors.Errorf("admissions file %s: %s is admitted twice for %s -- exactly one admission per resource per kind", path, e.Resource, e.Kind)
		}
		seen[key] = true
		if e.Reason == "" {
			return nil, errors.Errorf("admissions file %s: %s (%s) carries no reason", path, e.Resource, e.Kind)
		}
		if e.PromotionTracking == "" {
			return nil, errors.Errorf("admissions file %s: %s (%s) carries no promotionTracking -- an admission must say where its end is watched", path, e.Resource, e.Kind)
		}
	}
	return &doc, nil
}
