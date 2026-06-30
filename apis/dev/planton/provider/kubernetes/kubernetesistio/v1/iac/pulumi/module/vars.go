package module

var vars = struct {
	// names & repo
	SystemNamespace  string
	GatewayNamespace string
	HelmRepo         string
	BaseChart        string
	IstiodChart      string
	GatewayChart     string

	// DefaultStableVersion is the fallback Istio version used only when spec.version
	// is unset (spec.version is normally defaulted at manifest-load time). Keep this
	// in sync with the Istio version this Planton release targets (the crd2pulumi SDK
	// pin in pkg/kubernetes/kubernetestypes/Makefile and KubernetesIstioBaseCrds).
	DefaultStableVersion string
}{
	// namespaces
	SystemNamespace:  "istio-system",
	GatewayNamespace: "istio-ingress",

	// upstream repo & charts
	HelmRepo:     "https://istio-release.storage.googleapis.com/charts",
	BaseChart:    "base",
	IstiodChart:  "istiod",
	GatewayChart: "gateway",

	// version pin – bump together with the SDK/CRD pin when the release target moves
	DefaultStableVersion: "1.26.8",
}
