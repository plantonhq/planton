package module

import (
	"strconv"

	kuberneteskarpenterv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskarpenter/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kuberneteskarpenterv1alpha1.KubernetesKarpenterSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the charts' own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace both releases install into (resolved literal from the
	// spec's value-or-ref; "kube-system" is upstream's recommended home).
	Namespace string

	// Chart version resolved to the pinned default when unset. ONE version
	// drives BOTH releases — the karpenter and karpenter-crd charts version
	// together with the controller.
	ChartVersion string

	// CRD lifecycle with the spec defaults applied (install=true,
	// keep_on_uninstall=true) — resolved once here so the release wiring in
	// main.go and the outputs agree on whether the CRD release exists.
	CrdsInstall bool
	CrdsKeep    bool
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kuberneteskarpenterv1alpha1.KubernetesKarpenterIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesKarpenter.String(),
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

	crdsInstall := true
	if spec.GetCrds() != nil && spec.GetCrds().Install != nil {
		crdsInstall = spec.GetCrds().GetInstall()
	}
	crdsKeep := true
	if spec.GetCrds() != nil && spec.GetCrds().KeepOnUninstall != nil {
		crdsKeep = spec.GetCrds().GetKeepOnUninstall()
	}

	return &Locals{
		Spec:         spec,
		Labels:       labels,
		Namespace:    spec.Namespace.GetValue(),
		ChartVersion: chartVersion,
		CrdsInstall:  crdsInstall,
		CrdsKeep:     crdsKeep,
	}
}
