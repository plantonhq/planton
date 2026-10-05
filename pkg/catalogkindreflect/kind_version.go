package catalogkindreflect

import (
	"regexp"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// versionGrammar is the maturity-channel version grammar: v1alpha1 → v1beta1
// → v1. The version name declares the kind's compatibility channel, so a
// malformed version would silently misdeclare a kind's guarantees. Enum-value
// options are compile-time data protovalidate never evaluates, which makes
// the registry test built on this grammar the enforcement point.
var versionGrammar = regexp.MustCompile(`^v\d+((alpha|beta)\d+)?$`)

// KindVersion returns the kind's declared API version — the version half of
// the manifest apiVersion (e.g. "aws.planton.dev/v1alpha1") and the name of
// the version directory holding the kind's versioned contract
// (catalog/<provider>/<kind>/<version>/ — protos, stubs, spec test,
// reference page; the living kind sits at the kind root above it).
func KindVersion(kind catalogkind.CatalogKind) (string, error) {
	kindMeta, err := KindMeta(kind)
	if err != nil {
		return "", errors.Wrapf(err, "no kind metadata for %s", kind)
	}
	if kindMeta.Version == "" {
		return "", errors.Errorf(
			"kind %s declares no version in its registry metadata — every kind must declare one", kind)
	}
	return kindMeta.Version, nil
}

// KindVersionDir resolves a kind name to the version segment of
// its kind directory. Names arrive in directory and manifest shapes
// alike ("AwsVpc", "awsvpc", "aws-vpc"), so resolution goes through the
// tolerant KindFromString rather than the strict manifest-semantics index —
// safe because registry tests already guarantee canonical uniqueness.
//
// Path builders derive the version segment from the registry through this
// helper instead of assuming a literal, so a kind's directory always
// follows its declared version. An unknown name fails here, plainly, instead
// of composing a path that does not exist and failing later as a confusing
// file-not-found.
func KindVersionDir(kindName string) (string, error) {
	kind := KindFromString(kindName)
	if kind == catalogkind.CatalogKind_unspecified {
		return "", errors.Errorf(
			"cannot resolve %q to a catalog kind, so its kind directory cannot be located — check the kind name against the catalog", kindName)
	}
	return KindVersion(kind)
}
