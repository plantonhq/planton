package helmcrds

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
)

// Stamp keys written onto every derived CRD. They are the one contract both
// engines and every verifier read; the Terraform template quotes them as
// literals and a test in the generators package holds the literals equal to
// these constants.
const (
	// AnnotationSourceChart records "<repository>/<chart>" the CRD was derived
	// from (or the bundle URL template for bundle-branch kinds).
	AnnotationSourceChart = "planton.ai/crd-source-chart"
	// AnnotationSourceVersion records the version the CRD was derived at (the
	// chart version, or the upstream release for a bundle). The
	// never-downgrade check compares against it.
	AnnotationSourceVersion = "planton.ai/crd-source-version"
	// LabelSource carries the source's name (Source.Name) so a module can
	// select the CRDs it owns with one label selector (label values cannot
	// carry "/" or ":", which is why the full source lives in the annotation).
	LabelSource = "planton.ai/crd-source"
)

// HelmKeepAnnotation is Helm's own instruction to leave a resource behind on
// uninstall. A chart that templates its CRDs and marks them with it owns their
// lifecycle correctly (the cert-manager shape); a chart that templates them
// without it lets an uninstall cascade-delete every custom resource.
const HelmKeepAnnotation = "helm.sh/resource-policy"

// Source names where a kind's CRD set comes from at a pinned version. Exactly
// one of the two branches is used: the render branch when Chart is set, the
// bundle branch when BundleURL is set.
type Source struct {
	// Repository is the chart repository: an https:// index URL or an
	// oci://<registry>/<path> reference. Render branch.
	Repository string
	// Chart is the chart name inside Repository. Render branch.
	Chart string
	// Version is the exact chart version to render, or the upstream release
	// a bundle is fetched at. Both branches use it for the stamp; the bundle
	// branch also substitutes it into BundleURL.
	Version string
	// VersionField is the spec field a user sets to choose Version (for
	// example "spec.chart_version"), so a refusal names the setting that
	// fixes it. Empty when the module pins the version itself, as a kind
	// built on one upstream release does: the remedy is then a Planton
	// release that pins another version, never a setting.
	VersionField string
	// Username and Password authenticate to a private repository or registry,
	// exactly as the release install does. Empty for public charts.
	Username string
	Password string
	// Values are the release's OWN values documents, in helm -f order (later
	// documents override earlier ones), exactly as the release resource
	// receives them. The render must see what the install will see.
	Values []string
	// CRDOverride is one more YAML document merged LAST, carrying only the
	// chart's CRD switch turned on (for example "crds:\n  create: true\n").
	// The release itself installs with the switch off and skip_crds set, so
	// this override never reaches the cluster through Helm.
	//
	// The override also decides ownership of what the render produces. With
	// one, the module has turned the chart's switch on for the render and off
	// for the release, so EVERY CRD the render produces is module-owned.
	// Without one, the render is the release's own render: CRDs on Helm's
	// crds/ surface are module-owned (skip_crds keeps Helm off them), while
	// CRDs the chart templates belong to Helm and are reported as such.
	CRDOverride string
	// APIVersions feeds .Capabilities.APIVersions for charts that gate
	// templates on them (a chart that emits cert-manager annotations only when
	// cert-manager.io/v1 is served). Omitting a version the chart checks makes
	// the render silently drop or reshape a CRD.
	APIVersions []string
	// KubeVersion feeds .Capabilities.KubeVersion; empty uses Helm's default.
	KubeVersion string

	// BundleURL is a manifest upstream publishes its CRDs in: a CRD bundle
	// beside a chart that ships none, or the install manifest of an upstream
	// that publishes no chart at all (every non-CRD document in it is
	// ignored here; the module applies those itself). It may contain
	// "{{version}}", replaced with Version. Bundle branch. Use the URL form
	// that survives the next upstream release (Apache's dist tree holds only
	// the current release; archive.apache.org holds every release).
	BundleURL string
	// BundleName names the upstream project the bundle belongs to (for
	// example "agent-sandbox"): the bundle branch's twin of Chart, carried
	// by LabelSource as the mark that the CRDs are this source's. The URL
	// cannot serve, because unrelated projects publish from the same host.
	// Bundle branch; required there, and a valid label value.
	BundleName string
}

// IsBundle reports whether the source is the bundle branch.
func (s Source) IsBundle() bool { return s.BundleURL != "" }

// Name is what LabelSource carries and what a refusal calls the source: the
// chart name on the render branch, the bundle's own name on the bundle branch.
func (s Source) Name() string {
	if s.IsBundle() {
		return s.BundleName
	}
	return s.Chart
}

// versionNoun is how a reader recognizes one of the source's versions: a
// chart version on the render branch, an upstream release on the bundle
// branch ("agent-sandbox release").
func (s Source) versionNoun() string {
	if s.IsBundle() {
		return s.BundleName + " release"
	}
	return "chart version"
}

// VersionPhrase is versionNoun with a version, for example "chart version
// 0.120.0" or "agent-sandbox release v1.0.7": how every text, the deploy log
// included, names a version of this source.
func (s Source) VersionPhrase(version string) string {
	return s.versionNoun() + " " + version
}

// versionRemedy is the next step that moves the source to target (a version,
// or a version with a qualifier such as "0.120.0 or higher"): the setting
// when a user chooses the version, a Planton release when the module pins it.
func (s Source) versionRemedy(target string) string {
	if s.VersionField != "" {
		return fmt.Sprintf("set %s to %s", s.VersionField, target)
	}
	return fmt.Sprintf("deploy with a Planton release whose module pins %s (this module pins %s)", s.VersionPhrase(target), s.Version)
}

// validate refuses a source a module author assembled wrongly. These are
// defects in the module, not conditions a user meets, so they are plain
// errors rather than Failures.
func (s Source) validate() error {
	switch {
	case s.IsBundle() && s.Chart != "":
		return errors.New("a CRD source sets Chart or BundleURL, never both")
	case s.IsBundle() && s.BundleName == "":
		return errors.New("a bundle CRD source needs a BundleName")
	case s.IsBundle() && len(validation.IsValidLabelValue(s.BundleName)) > 0:
		return errors.Errorf("a bundle CRD source's BundleName must be a label value, got %q", s.BundleName)
	case !s.IsBundle() && s.Chart == "":
		return errors.New("a CRD source needs a Chart or a BundleURL")
	}
	return nil
}

// ResolvedBundleURL substitutes the version into the bundle URL template.
func (s Source) ResolvedBundleURL() string {
	return strings.ReplaceAll(s.BundleURL, "{{version}}", s.Version)
}

// SourceDescription is what AnnotationSourceChart records: the chart's
// coordinates for the render branch, the URL template for the bundle branch.
func (s Source) SourceDescription() string {
	if s.IsBundle() {
		return s.BundleURL
	}
	return strings.TrimSuffix(s.Repository, "/") + "/" + s.Chart
}

// Policy says what a kind accepts from the derivation. Typed kinds know their
// chart carries CRDs and pin its switches, so they expect CRDs and never meet
// Helm-managed ones; the generic Helm kind installs an arbitrary chart, so a
// chart without CRDs is ordinary and Helm-managed CRDs are a user's informed
// choice.
type Policy struct {
	// ExpectCRDs makes a derivation that yields no module-owned CRD a failure
	// (the chart's switch was renamed, or the chart stopped shipping CRDs).
	// False for a kind whose chart may legitimately carry none.
	ExpectCRDs bool
	// AllowHelmManaged accepts CRDs the chart templates as ordinary release
	// resources without Helm's keep annotation. Helm installs and upgrades
	// them with the release and deletes them with it, along with every custom
	// resource built on them. False refuses such a chart with the CRD names
	// and the remedies.
	AllowHelmManaged bool
}

// Derived is what one derivation yields.
type Derived struct {
	// Owned are the CRDs the module applies as kept, stamped resources ahead
	// of the release, sorted by name.
	Owned []CRD
	// HelmManaged names the CRDs the chart templates as release resources
	// without Helm's keep annotation (only possible without a CRDOverride).
	// Empty when the policy refused them, since Derive returns the failure
	// instead; populated when the policy allowed them, so a caller can say
	// what Helm owns.
	HelmManaged []string
}

// CRD is one derived CustomResourceDefinition, stamped and ready to apply.
type CRD struct {
	// Name is metadata.name (for example "opentelemetrycollectors.opentelemetry.io"),
	// the key every engine uses for the resource address so state addresses
	// stay stable across reorderings.
	Name string
	// YAML is the single-document manifest with the stamps applied.
	YAML string
}

// Derive produces the stamped CRD set for a source under a policy: render or
// fetch, split on document separators, keep only CustomResourceDefinitions,
// decide ownership, apply the policy, stamp each owned CRD. The release name
// and namespace are the identity the render runs under, so release-derived
// values inside a CRD (a conversion-webhook service, an inject-ca-from
// annotation) come out pointing at the real release.
func Derive(ctx context.Context, src Source, policy Policy, releaseName, namespace string) (Derived, error) {
	if err := src.validate(); err != nil {
		return Derived{}, err
	}
	var candidates []crdDocument
	var err error
	if src.IsBundle() {
		var documents []string
		documents, err = fetchBundle(ctx, src)
		if err != nil {
			return Derived{}, err
		}
		candidates, err = filterCRDs(documents, nil)
	} else {
		var rendered renderResult
		rendered, err = render(src, releaseName, namespace)
		if err != nil {
			return Derived{}, err
		}
		candidates, err = filterCRDs(rendered.documents, rendered.directoryNames)
	}
	if err != nil {
		return Derived{}, err
	}

	var result Derived
	owned := make([]CRD, 0, len(candidates))
	for _, candidate := range candidates {
		switch {
		case src.CRDOverride != "" || candidate.onDirectorySurface || src.IsBundle():
			owned = append(owned, candidate.CRD)
		case candidate.keptByChart:
			// The chart owns this CRD's lifecycle itself (templated with
			// Helm's keep annotation); Helm keeps it on uninstall and
			// upgrades it with the release. Nothing for the module to do.
		default:
			result.HelmManaged = append(result.HelmManaged, candidate.Name)
		}
	}
	if len(result.HelmManaged) > 0 && !policy.AllowHelmManaged {
		return Derived{}, HelmManagedCRDsFailure(src, result.HelmManaged)
	}
	if len(owned) == 0 && policy.ExpectCRDs {
		return Derived{}, zeroCRDsFailure(src)
	}

	result.Owned = make([]CRD, 0, len(owned))
	for _, crd := range owned {
		stamped, err := stamp(crd, src)
		if err != nil {
			return Derived{}, err
		}
		result.Owned = append(result.Owned, stamped)
	}
	return result, nil
}
