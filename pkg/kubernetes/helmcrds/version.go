package helmcrds

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Helm's ownership annotations on a release resource. A CRD carrying them was
// installed by `helm install` as a templated resource of that release.
const (
	helmReleaseNameAnnotation      = "meta.helm.sh/release-name"
	helmReleaseNamespaceAnnotation = "meta.helm.sh/release-namespace"
)

// ExistingCRD is what a cluster already carries under the name of a CRD the
// module is about to apply, read by name before anything is written. Both
// engines read the same fields: the stamps (whose source, at which version)
// and the ownership marks that say who else may have put it there.
type ExistingCRD struct {
	Name string
	// SourceLabel and Version are this module's stamps (LabelSource,
	// AnnotationSourceVersion) when the CRD was derived by a Planton module;
	// both empty when something else applied it.
	SourceLabel string
	Version     string
	// HelmReleaseName and HelmReleaseNamespace are set when Helm owns the CRD
	// as a release resource (meta.helm.sh/release-* annotations).
	HelmReleaseName      string
	HelmReleaseNamespace string
	// Managers are the field managers in metadata.managedFields, the API
	// server's own record of who has written the object (kubectl,
	// pulumi-kubernetes, an operator's name); the fallback when no other mark
	// says who owns it.
	Managers []string
}

// ExistingFromObject reads an ExistingCRD from an object's metadata as the
// API server returns it: labels, annotations, and the managedFields managers.
// Both engines feed it what they read, so the ownership sentence is composed
// in one place.
func ExistingFromObject(name string, labels, annotations map[string]string, managers []string) ExistingCRD {
	return ExistingCRD{
		Name:                 name,
		SourceLabel:          labels[LabelSource],
		Version:              annotations[AnnotationSourceVersion],
		HelmReleaseName:      annotations[helmReleaseNameAnnotation],
		HelmReleaseNamespace: annotations[helmReleaseNamespaceAnnotation],
		Managers:             managers,
	}
}

// ownedBy reports whether this source stamped the CRD.
func (c ExistingCRD) ownedBy(src Source) bool {
	return c.SourceLabel == labelSourceValue(src)
}

// OwnerDescription is the sentence the refusal names the other owner with,
// composed from the most telling mark present: Helm's release annotations, a
// Planton stamp from a different source, the field managers, or the honest
// absence of any mark.
func (c ExistingCRD) OwnerDescription() string {
	switch {
	case c.HelmReleaseName != "":
		return fmt.Sprintf("owned by Helm release %s in namespace %s", c.HelmReleaseName, c.HelmReleaseNamespace)
	case c.SourceLabel != "":
		return fmt.Sprintf("derived from %s by another Planton module", c.SourceLabel)
	case len(c.Managers) > 0:
		return fmt.Sprintf("last written by field manager(s) %s", strings.Join(c.Managers, ", "))
	}
	return "it carries no ownership marks"
}

// CheckOwnership refuses when any CRD about to be applied already exists
// without this source's stamp. A CRD this source stamped (a kept CRD from an
// earlier install, an upgrade, a re-adoption) passes; its version is then the
// never-downgrade check's business. Callers pass only CRDs that exist.
func CheckOwnership(existing []ExistingCRD, src Source) error {
	for _, crd := range existing {
		if !crd.ownedBy(src) {
			return OwnedElsewhereFailure(crd, src)
		}
	}
	return nil
}

// CheckNoDowngrade refuses when any existing CRD carries a strictly higher
// source version than src.Version, the version about to be applied. Helm's
// own semver (Masterminds/semver) orders the versions, so prerelease and
// build metadata behave exactly as Helm's repository index would order them,
// and an upstream release's leading "v" reads as it does in a tag. A CRD
// without a version stamp is not this source's (CheckOwnership answers for
// it) and is skipped here.
func CheckNoDowngrade(existing []ExistingCRD, src Source) error {
	requested, err := semver.NewVersion(src.Version)
	if err != nil {
		return UnparseableVersionFailure(src, "requested", src.Version)
	}
	for _, crd := range existing {
		if crd.Version == "" {
			continue
		}
		current, err := semver.NewVersion(crd.Version)
		if err != nil {
			return UnparseableVersionFailure(src, "cluster's "+crd.Name, crd.Version)
		}
		if requested.LessThan(current) {
			return DowngradeFailure(src, crd.Name, crd.Version)
		}
	}
	return nil
}
