package module

import (
	"strconv"

	kubernetesingressnginxv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesingressnginx/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesingressnginxv1alpha1.KubernetesIngressNginxSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace the controller installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name: multiple
	// controller instances per cluster (public + internal split) are a
	// first-class upstream pattern, so each manifest gets its own release.
	// The chart fullname is pinned to this too, which also isolates
	// leader election per instance (electionID defaults to
	// "<fullname>-leader" in the chart).
	ReleaseName string

	// IngressClass this instance owns (spec default: "nginx").
	IngressClassName string

	// spec.controller of the IngressClass. Empty spec value derives:
	// the chart default "k8s.io/ingress-nginx" for class "nginx",
	// otherwise "k8s.io/<class-name>" so additional controllers isolate
	// automatically without the user inventing a vocabulary.
	IngressClassControllerValue string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Deterministic chart-derived object names ("<fullname>-controller"
	// and "-internal" sibling) — what verification and downstream
	// composition key off.
	ControllerServiceName string
	InternalServiceName   string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesingressnginxv1alpha1.KubernetesIngressNginxIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesIngressNginx.String(),
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

	className := spec.GetIngressClass().GetName()
	if className == "" {
		className = "nginx"
	}

	controllerValue := spec.GetIngressClass().GetControllerValue()
	if controllerValue == "" {
		if className == "nginx" {
			controllerValue = "k8s.io/ingress-nginx"
		} else {
			controllerValue = "k8s.io/" + className
		}
	}

	internalServiceName := ""
	if spec.GetService().GetInternal().GetEnabled() {
		internalServiceName = target.Metadata.Name + "-controller-internal"
	}

	return &Locals{
		Spec:                        spec,
		Labels:                      labels,
		Namespace:                   spec.Namespace.GetValue(),
		ReleaseName:                 target.Metadata.Name,
		IngressClassName:            className,
		IngressClassControllerValue: controllerValue,
		ChartVersion:                chartVersion,
		ControllerServiceName:       target.Metadata.Name + "-controller",
		InternalServiceName:         internalServiceName,
	}
}
