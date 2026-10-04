package module

// Istio CRD source pin.
//
// IstioRelease MUST stay in sync with `istio_release` in
// pkg/kubernetes/kubernetestypes/Makefile. That Makefile pin is the single source of
// truth for the Istio version this Planton release targets: it drives the
// crd2pulumi-generated typed SDK that the Istio kinds are built against, and this
// constant drives the CRDs installed on the cluster. Keeping them equal guarantees the
// installed CRD schema matches the typed custom resources (no silent field pruning).
const (
	// IstioRelease is the istio/istio git ref the base CRDs are fetched from.
	// Always an exact release TAG (e.g. "1.30.3"), never a release BRANCH: a
	// branch ref moves as patches land, so the same deployed resource would
	// install different CRD schemas at different times — tag pinning keeps
	// installs reproducible and exactly matched to the generated SDK.
	IstioRelease = "1.30.3"
)

// GetCrdManifestURL returns the upstream istio/base CRDs-only bundle URL.
// This bundle contains only CustomResourceDefinitions (no istiod, no controller).
func GetCrdManifestURL() string {
	return "https://raw.githubusercontent.com/istio/istio/" + IstioRelease +
		"/manifests/charts/base/files/crd-all.gen.yaml"
}
