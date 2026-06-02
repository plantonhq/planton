package module

// Istio CRD source pin.
//
// IstioRelease MUST stay in sync with `istio_release` in
// pkg/kubernetes/kubernetestypes/Makefile. That Makefile pin is the single source of
// truth for the Istio version this OpenMCF release targets: it drives the
// crd2pulumi-generated typed SDK that the Istio components are built against, and this
// constant drives the CRDs installed on the cluster. Keeping them equal guarantees the
// installed CRD schema matches the typed custom resources (no silent field pruning).
const (
	// IstioRelease is the istio/istio git ref the base CRDs are fetched from.
	IstioRelease = "release-1.26"
)

// GetCrdManifestURL returns the upstream istio/base CRDs-only bundle URL.
// This bundle contains only CustomResourceDefinitions (no istiod, no controller).
func GetCrdManifestURL() string {
	return "https://raw.githubusercontent.com/istio/istio/" + IstioRelease +
		"/manifests/charts/base/files/crd-all.gen.yaml"
}
