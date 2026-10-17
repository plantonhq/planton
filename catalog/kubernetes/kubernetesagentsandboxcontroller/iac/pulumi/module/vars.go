package module

// vars holds the upstream contract the module installs: the pinned release
// and the fixed names upstream's own manifests bake in. Every value has an
// exact twin in the Terraform module's locals.tf, and a test holds the
// release pin equal across the two engines.
var vars = struct {
	// Release is the pinned kubernetes-sigs/agent-sandbox release tag. The
	// catalog's Sandbox, SandboxTemplate, SandboxWarmPool and SandboxClaim
	// kinds mirror this release's resource definitions field for field, so
	// moving it is a deliberate catalog update (the spec has no version
	// field). Always an exact tag: it names the release assets, the router's
	// deployment files and both images.
	Release string

	// BundleName is the upstream project's own name, the ownership label the
	// kept-CRD primitive stamps on the definitions.
	BundleName string

	// Namespace is fixed: upstream's manifests bake it in, and the default
	// network policy every SandboxTemplate gets admits only a router from it.
	Namespace string

	// UpstreamRegistry is where upstream publishes both images; the spec's
	// image_registry replaces exactly this part of an image.
	UpstreamRegistry string

	ControllerDeploymentName string
	ControllerContainerName  string
	ControllerServiceName    string

	// ConfigMapName and ConfigMapKey are the runtime configuration the
	// extensions read once at startup from ConfigMapMountPath (upstream's
	// release manifest mounts the ConfigMap as optional).
	ConfigMapName      string
	ConfigMapKey       string
	ConfigMapMountPath string

	RouterDeploymentName string
	RouterContainerName  string
	RouterServiceName    string
	RouterImageRepo      string

	// MetricsTLSMountPath is where the controller's metrics certificate
	// Secret is mounted (--metrics-cert-dir reads tls.crt and tls.key there).
	MetricsTLSMountPath string

	// Router mount paths for the Secrets its settings name as files.
	RouterTLSMountPath          string
	RouterClientCAMountPath     string
	RouterScopedTokenV1Mount    string
	RouterScopedTokenV2Mount    string
	RouterTerminationGraceExtra int64

	// ConfigChecksumAnnotation carries a digest of the runtime configuration
	// on the controller's pod template, so a change restarts the controller
	// (upstream reads the ConfigMap only at startup).
	ConfigChecksumAnnotation string
}{
	Release:                     "v1.0.7",
	BundleName:                  "agent-sandbox",
	Namespace:                   "agent-sandbox-system",
	UpstreamRegistry:            "registry.k8s.io",
	ControllerDeploymentName:    "agent-sandbox-controller",
	ControllerContainerName:     "agent-sandbox-controller",
	ControllerServiceName:       "agent-sandbox-controller",
	ConfigMapName:               "agent-sandbox-config",
	ConfigMapKey:                "allowed-label-domains",
	ConfigMapMountPath:          "/etc/sandbox-config",
	RouterDeploymentName:        "sandbox-router",
	RouterContainerName:         "sandbox-router",
	RouterServiceName:           "sandbox-router-svc",
	RouterImageRepo:             "registry.k8s.io/agent-sandbox/sandbox-router-go",
	MetricsTLSMountPath:         "/etc/agent-sandbox/metrics-tls",
	RouterTLSMountPath:          "/etc/sandbox-router/tls",
	RouterClientCAMountPath:     "/etc/sandbox-router/client-ca",
	RouterScopedTokenV1Mount:    "/etc/sandbox-router/scoped-token-v1",
	RouterScopedTokenV2Mount:    "/etc/sandbox-router/scoped-token-v2",
	RouterTerminationGraceExtra: 15,
	ConfigChecksumAnnotation:    "planton.ai/config-sha256",
}

// releaseAssetBaseURL serves the release's assets; the asset is chosen by
// whether the extensions are on (sandbox.yaml carries the core only,
// sandbox-with-extensions.yaml the core and the extensions in one render).
const releaseAssetBaseURL = "https://github.com/kubernetes-sigs/agent-sandbox/releases/download/"

// routerFilesBaseURL serves the router's deployment files at a tag. Upstream
// publishes no release asset for the router; these files are its install.
const routerFilesBaseURL = "https://raw.githubusercontent.com/kubernetes-sigs/agent-sandbox/"

// releaseAsset names the release asset for the extensions setting.
func releaseAsset(extensions bool) string {
	if extensions {
		return "sandbox-with-extensions.yaml"
	}
	return "sandbox.yaml"
}

// ReleaseAssetURL is the pinned release asset the module applies.
func ReleaseAssetURL(extensions bool) string {
	return releaseAssetBaseURL + vars.Release + "/" + releaseAsset(extensions)
}

// CRDBundleURLTemplate is the same asset as a kept-CRD bundle URL, with the
// version left for the primitive to fill in.
func CRDBundleURLTemplate(extensions bool) string {
	return releaseAssetBaseURL + "{{version}}/" + releaseAsset(extensions)
}

// RouterFileURL is one of the router's deployment files at the pinned tag.
func RouterFileURL(file string) string {
	return routerFilesBaseURL + vars.Release + "/sandbox-router/deploy/" + file
}

// Router deployment files, by role. Which are applied depends on the spec:
// rbac.yaml only with the pod-IP cache, rbac-tokenreview.yaml only with
// token-review authorization, pdb.yaml only with a disruption budget,
// networkpolicy.yaml only with the policy on.
const (
	routerFileServiceAccount = "serviceaccount.yaml"
	routerFileRBAC           = "rbac.yaml"
	routerFileTokenReview    = "rbac-tokenreview.yaml"
	routerFileDeployment     = "deployment.yaml"
	routerFileService        = "service.yaml"
	routerFilePDB            = "pdb.yaml"
	routerFileNetworkPolicy  = "networkpolicy.yaml"
)

// routerFiles is every router file the module may apply, in apply order.
var routerFiles = []string{
	routerFileServiceAccount,
	routerFileRBAC,
	routerFileTokenReview,
	routerFileDeployment,
	routerFileService,
	routerFilePDB,
	routerFileNetworkPolicy,
}
