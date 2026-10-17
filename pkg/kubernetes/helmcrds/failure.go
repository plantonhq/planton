package helmcrds

import (
	"fmt"
	"strings"

	"github.com/plantonhq/planton/pkg/failure"
)

// Failure is the one shape every CRD-lifecycle error takes: what the platform
// observed (with the value), what it most likely means (one root cause), and
// the exact next step. It is the repository-wide failure shape, so a CRD
// refusal reads like every other explained refusal the engines and the CLI
// produce.
type Failure = failure.Failure

// Raw substrings Helm produces for the failures the primitive anticipates. The
// provider's helm_template data source and the SDK are the same code, so one
// vocabulary classifies both. The words of each explanation live in
// pkg/failure, where the layer that runs the Terraform engine reads the same
// raw texts out of the provider's output; the Source here only supplies the
// values Helm's text does not always carry.
const (
	helmVersionNotFound  = "not found in"
	helmRepoUnreachable  = "is not a valid chart repository or cannot be reached"
	helmOCINotFound      = "FetchReference"
	helmStaleRepoCache   = "no cached repo found"
	helmChartNotInRepo   = "chart \"" // "chart \"x\" version \"y\" not found in ..."
	helmRepoIndexMissing = "index.yaml"
)

// classifyLocateError turns Helm's chart-location error into a Failure.
func classifyLocateError(src Source, err error) error {
	raw := err.Error()
	switch {
	case strings.Contains(raw, helmStaleRepoCache):
		return failure.HelmStaleRepositoryCache(src.Chart+" "+src.Version, raw)
	case strings.HasPrefix(src.Repository, "oci://") && strings.Contains(raw, helmOCINotFound):
		return failure.HelmOCIVersionNotPublished(src.Repository, src.Chart, src.Version, raw)
	case strings.Contains(raw, helmVersionNotFound) && strings.Contains(raw, helmChartNotInRepo):
		return failure.HelmVersionNotPublished(src.Repository, src.Chart, src.Version, raw)
	case strings.Contains(raw, helmRepoUnreachable) || strings.Contains(raw, helmRepoIndexMissing):
		return failure.HelmRepositoryUnreachable(src.Repository, raw)
	}
	chart := src.Chart + " " + src.Version
	return &Failure{
		Observed: fmt.Sprintf("Helm could not locate chart %s in %s (%s)", chart, src.Repository, raw),
		Meaning:  "the chart coordinates the module carries (repository, chart name, version) do not resolve from here",
		NextStep: "check spec.chart_version against the repository index and the module's chart identity, then re-run",
	}
}

func renderFailure(src Source, err error) error {
	return &Failure{
		Observed: fmt.Sprintf("rendering chart %s %s from %s failed (%s)", src.Chart, src.Version, src.Repository, err.Error()),
		Meaning:  "the chart's templates rejected the values the release would install with; the release install would fail identically",
		NextStep: "fix the value the error names (spec fields or helm_values), or pin a chart_version whose templates accept it",
	}
}

func zeroCRDsFailure(src Source) error {
	if src.IsBundle() {
		return &Failure{
			Observed: fmt.Sprintf("the CRD bundle at %s contained no CustomResourceDefinition documents", src.ResolvedBundleURL()),
			Meaning:  "upstream changed what that URL serves, or the version has no CRD bundle",
			NextStep: "open the URL, confirm it is the CRD bundle for this version, and correct the module's bundle URL template if upstream moved it",
		}
	}
	return &Failure{
		Observed: fmt.Sprintf("rendering chart %s %s with its CRD switch on (%s) produced no CustomResourceDefinition documents", src.Chart, src.Version, strings.TrimSpace(src.CRDOverride)),
		Meaning:  "the chart's CRD switch is not the value the module sets at this version (upstream renamed it), or the chart stopped shipping CRDs",
		NextStep: fmt.Sprintf("read the chart's values at this version (`helm show values %s --repo %s --version %s`) and update the module's CRD override to the switch the chart uses", src.Chart, src.Repository, src.Version),
	}
}

func bundleFetchFailure(src Source, status int, err error) error {
	url := src.ResolvedBundleURL()
	if err != nil {
		return &Failure{
			Observed: fmt.Sprintf("the CRD bundle at %s could not be fetched (%s)", url, err.Error()),
			Meaning:  "the bundle host does not resolve or egress to it is blocked from where this plan runs",
			NextStep: fmt.Sprintf("confirm DNS and egress to the bundle host (`curl -I %s`), then re-run", url),
		}
	}
	if status == 404 {
		if src.VersionField == "" {
			return &Failure{
				Observed: fmt.Sprintf("the CRD bundle URL %s answered HTTP 404", url),
				Meaning:  fmt.Sprintf("upstream does not serve %s, the release this module pins, at that URL: the release was withdrawn, or upstream changed its download layout", src.VersionPhrase(src.Version)),
				NextStep: fmt.Sprintf("confirm what upstream publishes for %s (`curl -I %s`), then correct the module's bundle URL template or its pin", src.VersionPhrase(src.Version), url),
			}
		}
		return &Failure{
			Observed: fmt.Sprintf("the CRD bundle URL %s answered HTTP 404", url),
			Meaning:  "no bundle is published at this version, or upstream changed the URL pattern",
			NextStep: fmt.Sprintf("confirm version %s exists at the upstream download page; set %s to a published version, or correct the module's bundle URL template", src.Version, src.VersionField),
		}
	}
	return &Failure{
		Observed: fmt.Sprintf("the CRD bundle URL %s answered HTTP %d", url, status),
		Meaning:  "the upstream host is refusing or failing the request",
		NextStep: fmt.Sprintf("retry once the host answers 200 (`curl -I %s`); if it persists, the upstream download location has changed", url),
	}
}

// HelmManagedCRDsFailure is the refusal when a chart templates CRDs as ordinary
// release resources without Helm's keep annotation and the kind's policy has
// not accepted that. Helm would install and upgrade them with the release and
// delete them with it, taking every custom resource built on them; nothing the
// module applies can protect a resource Helm owns. The remedies are ordered
// from most to least protective.
func HelmManagedCRDsFailure(src Source, names []string) error {
	return &Failure{
		Observed: fmt.Sprintf("chart %s %s templates %d CustomResourceDefinition(s) as release resources without the %s: keep annotation (%s)", src.Chart, src.Version, len(names), HelmKeepAnnotation, strings.Join(names, ", ")),
		Meaning:  "Helm owns these CRDs: it installs and upgrades them with the release and deletes them when the release is uninstalled, along with every custom resource built on them; keep_on_uninstall cannot reach resources Helm owns",
		NextStep: "use the catalog's typed kind for this chart if one exists; otherwise turn on the chart's own keep switch in its values (cert-manager-style `crds.keep: true`) so the chart protects them itself; or set spec.crds.allow_helm_managed to true to accept that these CRDs live and die with the release",
	}
}

// DowngradeFailure is the refusal when the cluster already carries a CRD at a
// higher source version than src.Version, the version this deploy would
// apply. Lowering a schema can strip fields from existing custom resources on
// their next write, so the primitive stops before touching anything.
func DowngradeFailure(src Source, crdName, existingVersion string) error {
	asks := "the manifest asks for"
	if src.VersionField == "" {
		asks = "this module pins"
	}
	return &Failure{
		Observed: fmt.Sprintf("the cluster carries CRD %s derived from %s; %s %s", crdName, src.VersionPhrase(existingVersion), asks, src.VersionPhrase(src.Version)),
		Meaning:  "applying an older CRD schema over a newer one can strip fields from existing custom resources the next time anything writes them",
		NextStep: fmt.Sprintf("%s; if you accept the loss and want the older schema, delete the CRD first (`kubectl delete crd %s`) and re-run", src.versionRemedy(existingVersion+" or higher"), crdName),
	}
}

// OwnedElsewhereFailure is the refusal when a CRD the module is about to apply
// already exists on the cluster without this source's stamp: a hand-run
// `helm install`, a colleague's `kubectl apply`, another tool, or a Planton
// module deriving the same CRD name from a different chart. Server-side apply
// would take it over silently, and because the never-downgrade check can only
// order versions it stamped, a newer schema could be lowered without a word.
// So the module stops before writing, names the owner it found, and offers the
// two honest ways forward: leave the definitions with their owner, or hand
// them over explicitly once they are known to match the pinned version.
//
// owner is the sentence Existing.OwnerDescription composes; helmOwned says
// whether Helm owns the CRD as a release resource, because handing over a
// CRD Helm still owns is a trap (Helm deletes it with that release later), so
// that remedy first tells the user to free it.
//
// The first remedy differs by branch. A chart's release installs the
// operator either way, so leaving the definitions with their owner is
// harmless. A bundle kind installs an upstream's own workloads, and a foreign
// definition most often means another installation of that upstream is
// already running (a managed add-on, a `kubectl apply` of the same release),
// which must never get a second copy beside it, so the remedy says when
// leaving the definitions with their owner is right.
func OwnedElsewhereFailure(crd ExistingCRD, src Source) error {
	handOver := fmt.Sprintf("hand them to this module once you know they match %s: `kubectl label crd %s %s=%s` and `kubectl annotate crd %s %s=%s %s=%s`, then re-run",
		src.VersionPhrase(src.Version), crd.Name, LabelSource, labelSourceValue(src), crd.Name, AnnotationSourceChart, src.SourceDescription(), AnnotationSourceVersion, src.Version)
	if crd.HelmReleaseName != "" {
		handOver = fmt.Sprintf("or, to move ownership here, first uninstall Helm release %s (or mark the CRD helm.sh/resource-policy: keep in that release) so Helm will not delete it later, then %s",
			crd.HelmReleaseName, handOver)
	} else {
		handOver = "or " + handOver
	}
	leaveWithOwner := "set spec.crds.install to false to leave the definitions with their current owner (the release still installs and uses them)"
	if src.IsBundle() {
		leaveWithOwner = fmt.Sprintf("set spec.crds.install to false only when another tool manages these definitions and no other installation of %s runs on this cluster (this kind still installs %s itself)", src.Name(), src.Name())
	}
	return &Failure{
		Observed: fmt.Sprintf("CRD %s already exists on the cluster and was not applied by this module (%s)", crd.Name, crd.OwnerDescription()),
		Meaning:  fmt.Sprintf("two owners would write one schema, and the module cannot tell whether the existing definition is newer or older than %s, so it stops before changing it", src.VersionPhrase(src.Version)),
		NextStep: fmt.Sprintf("%s; %s", leaveWithOwner, handOver),
	}
}

// CRDApplyDeniedFailure is the refusal when the identity the deploy runs as
// may not write CustomResourceDefinitions. The Pulumi twin learns this from a
// SelfSubjectAccessReview before anything registers; the Terraform twin learns
// it from the API server at apply, and the layer that runs the engine explains
// the same words (pkg/failure.KubernetesForbidden). reason is what the review
// answered, kept as the observation's raw text.
func CRDApplyDeniedFailure(user, verb, reason string) error {
	return failure.KubernetesForbidden(user, verb, failure.CustomResourceDefinitionsResource, "apiextensions.k8s.io", "at the cluster scope", reason)
}

// UnparseableVersionFailure covers a stamp or a requested version that is not
// semver. Refusing is safer than guessing an order.
func UnparseableVersionFailure(src Source, which, value string) error {
	next := "use the exact version string the chart repository index publishes (for example 0.120.0)"
	switch {
	case src.IsBundle() && src.VersionField != "":
		next = fmt.Sprintf("set %s to the release tag exactly as upstream publishes it (for example v1.0.7)", src.VersionField)
	case src.IsBundle():
		next = "the module's pinned release must be upstream's tag exactly as published (for example v1.0.7); correct the module's pin"
	}
	return &Failure{
		Observed: fmt.Sprintf("the %s %s %q is not a semantic version", which, src.versionNoun(), value),
		Meaning:  "the never-downgrade check cannot order versions it cannot parse, so it refuses rather than guess",
		NextStep: next,
	}
}
