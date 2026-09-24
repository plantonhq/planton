package module

var vars = struct {
	// OperatorRelease is the pinned tektoncd/operator release tag.
	//
	// MUST stay in sync with the Terraform module's operator_release
	// local: the TektonConfig surface the KubernetesTekton kind renders
	// is designed against this release's operator API. Always an exact
	// release TAG, never a branch — tag pinning keeps installs
	// reproducible. Moving it moves the image table too (images.go): a
	// test and both engines refuse a table read from another release.
	OperatorRelease string

	// Namespace is the fixed installation namespace. The release
	// manifest bakes `tekton-operator` into its own cross-references
	// (RBAC subjects, the webhook Service) — it is not configurable, and
	// exactly one install per cluster is the upstream contract.
	Namespace string

	// OperatorDeploymentName / WebhookDeploymentName are the release
	// manifest's fixed Deployment names the typed overrides patch.
	OperatorDeploymentName string
	WebhookDeploymentName  string

	// LifecycleContainerName is the operator Deployment's container that
	// reconciles the components and reads the IMAGE_* variables naming
	// their images; image_registry appends its entries there.
	LifecycleContainerName string

	// ConfigDefaultsConfigMapName is the manifest ConfigMap whose
	// AUTOINSTALL_COMPONENTS key the module ALWAYS patches to "false":
	// the operator must never race the KubernetesTekton declaration for
	// ownership of the cluster's TektonConfig (the operator would
	// auto-create one named `config` with profile `all` — the exact
	// object the declaration kind renders).
	ConfigDefaultsConfigMapName string

	// TektonConfigCrdName is the CRD the KubernetesTekton kind renders
	// against — deleted with this resource (cascade warning on the
	// spec).
	TektonConfigCrdName string
}{
	OperatorRelease:             "v0.80.0",
	Namespace:                   "tekton-operator",
	OperatorDeploymentName:      "tekton-operator",
	WebhookDeploymentName:       "tekton-operator-webhook",
	LifecycleContainerName:      "tekton-operator-lifecycle",
	ConfigDefaultsConfigMapName: "tekton-config-defaults",
	TektonConfigCrdName:         "tektonconfigs.operator.tekton.dev",
}

// ManifestURL is the released single-file manifest for the pinned tag —
// the operator's OFFICIAL distribution (the in-repo Helm chart is
// unpublished, version "devel"). The GitHub release asset is immutable
// per tag; the old storage.googleapis.com release host is dead.
func ManifestURL() string {
	return "https://github.com/tektoncd/operator/releases/download/" +
		vars.OperatorRelease + "/release.yaml"
}
