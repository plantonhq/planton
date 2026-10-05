package module

import (
	"fmt"
	"strconv"

	kuberneteskyvernov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskyverno/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kuberneteskyvernov1alpha1.KubernetesKyvernoSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace the engine installs into (resolved literal from the
	// spec's value-or-ref). The engine's webhooks exclude this namespace
	// by default so a misbehaving policy cannot lock Kyverno out.
	Namespace string

	// ReleaseName is metadata.name. fullnameOverride pins the chart
	// fullname to it, so every chart-derived name below follows from it.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// AdmissionServiceName is the webhook Service the runtime-registered
	// webhook configurations point at: "<fullname>-svc" (chart-truth:
	// the admission-controller serviceName helper). Exported.
	AdmissionServiceName string

	// ConfigMapName is the engine's runtime ConfigMap (resource filters,
	// webhook selectors): the fullname when config.create is true
	// (chart-truth: the config configMapName helper). Exported — the
	// object to inspect when a resource is unexpectedly skipped.
	ConfigMapName string

	// WebhookGCConfigMapName is the module-owned destroy sentinel
	// ("<name>-webhook-gc") that runs label-selected webhook cleanup
	// after the helm release is gone. Distinct from ConfigMapName.
	WebhookGCConfigMapName string
}

// initializeLocals extracts and transforms spec fields into module-local
// values. The fullname budget is enforced FAIL-LOUD here (the Terraform
// twin uses a precondition): the chart derives child names from the
// fullname and silently truncates past 63 chars, breaking its own
// name-based wiring.
func initializeLocals(_ *pulumi.Context, iacInput *kuberneteskyvernov1alpha1.KubernetesKyvernoIacInput) (*Locals, error) {
	target := iacInput.Target
	spec := target.Spec

	if len(target.Metadata.Name) > vars.FullnameMaxLen {
		return nil, fmt.Errorf(
			"metadata.name %q is %d characters; the kyverno chart derives the webhook Service, "+
				"the runtime ConfigMap and the pre-delete hook Job (longest suffix: -hook-pre-delete) "+
				"from it and truncates past the Kubernetes 63-character limit — use a name of at "+
				"most %d characters",
			target.Metadata.Name, len(target.Metadata.Name), vars.FullnameMaxLen)
	}

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesKyverno.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	chartVersion := spec.GetChartVersion()
	if chartVersion == "" {
		chartVersion = vars.DefaultChartVersion
	}

	return &Locals{
		Spec:                   spec,
		Labels:                 labels,
		Namespace:              spec.Namespace.GetValue(),
		ReleaseName:            target.Metadata.Name,
		ChartVersion:           chartVersion,
		AdmissionServiceName:   target.Metadata.Name + "-svc",
		ConfigMapName:          target.Metadata.Name,
		WebhookGCConfigMapName: target.Metadata.Name + "-webhook-gc",
	}, nil
}
