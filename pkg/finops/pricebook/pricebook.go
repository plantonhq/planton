// Package pricebook loads and validates the per-provider price books --
// the catalog/_pricing/pricebook/<provider>.yaml documents pinning every
// unit price the catalog's cost estimates cite (each with source URL and
// retrieval date). Estimate models reference entries by slug and never
// restate a price, so a price exists in exactly one place; the estimate
// generator joins the two and additionally rejects book entries no model
// references. Enrollment is the file's presence: every discovered book is
// held to this package's conformance gate, with no allowlist to keep in
// sync.
package pricebook

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"

	pricebookv1 "github.com/plantonhq/planton/finops/pricebook/v1"
	"github.com/plantonhq/planton/pkg/protobufyaml"
)

// Dir is the price books' home, relative to the repo root. Prices live
// centrally rather than beside kinds because one SKU's price serves
// many kinds and churns on the provider's cadence -- one tree to
// refresh, no touch on the kinds.
const Dir = "catalog/_pricing/pricebook"

// GlobalRegion is the region value of entries the provider prices
// identically everywhere (management fees, globally priced network SKUs).
const GlobalRegion = "global"

// Path is a provider's price book location. The filename is the
// provider's identity (the provider directory name under catalog/).
func Path(repoRoot, provider string) string {
	return filepath.Join(repoRoot, Dir, provider+".yaml")
}

// Discover returns the providers that ship a price book, sorted.
func Discover(repoRoot string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(repoRoot, Dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	var providers []string
	for _, m := range matches {
		providers = append(providers, strings.TrimSuffix(filepath.Base(m), ".yaml"))
	}
	sort.Strings(providers)
	return providers, nil
}

// Load reads and strictly parses a provider's price book.
func Load(repoRoot, provider string) (*pricebookv1.PriceBook, error) {
	book := &pricebookv1.PriceBook{}
	if err := protobufyaml.Load(Path(repoRoot, provider), book); err != nil {
		return nil, errors.Wrapf(err, "loading price book for %s", provider)
	}
	return book, nil
}

// Entries returns a provider's price book entries keyed by slug name, for
// resolving estimate model references. A missing book is not an error --
// providers without priced estimates simply have no book -- so callers get
// an empty map and their own "entry not found" failures stay precise.
func Entries(repoRoot, provider string) (map[string]*pricebookv1.PriceBookEntry, error) {
	providers, err := Discover(repoRoot)
	if err != nil {
		return nil, err
	}
	entries := map[string]*pricebookv1.PriceBookEntry{}
	for _, p := range providers {
		if p != provider {
			continue
		}
		book, err := Load(repoRoot, provider)
		if err != nil {
			return nil, err
		}
		for _, entry := range book.GetSpec().GetEntries() {
			entries[entry.GetName()] = entry
		}
	}
	return entries, nil
}
