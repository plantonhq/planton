package catalogkindreflect

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// KindDeprecations returns the kind's declared version deprecations — the
// schema versions announced as on their way out, each with its optional
// authored note. Most kinds return an empty slice; a deprecated version keeps
// serving and converting exactly as before (deprecation is advice, never a
// gate), so callers use this purely to ANNOUNCE: discovery responses,
// dashboards, CLI listings.
//
// The contract each entry obeys (a deprecation names an existing, non-served
// version with an authored conversion path to the served version) is enforced
// where each fact lives: grammar/duplicates/not-served by the registry test
// beside this accessor, existence and conversion-path by the catalog bundle's
// conformance gate.
func KindDeprecations(kind catalogkind.CatalogKind) ([]*catalogkind.CatalogKindVersionDeprecation, error) {
	kindMeta, err := KindMeta(kind)
	if err != nil {
		return nil, errors.Wrapf(err, "no kind metadata for %s", kind)
	}
	return kindMeta.Deprecations, nil
}
