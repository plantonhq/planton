package module

import (
	"fmt"
	"strconv"

	kubernetestektonv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestekton/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	KubernetesTekton *kubernetestektonv1alpha1.KubernetesTekton
	Spec             *kubernetestektonv1alpha1.KubernetesTektonSpec

	// ResourceName keys the rendered CR in the Pulumi state (the CR's
	// own name is the operator-required fixed `config` — see vars).
	ResourceName string

	// Labels tie the CR back to the Planton resource.
	Labels map[string]string

	// Profile is the resolved profile (`all` when the spec leaves it
	// empty — the operator's own default).
	Profile string

	// TargetNamespace is the resolved component namespace.
	TargetNamespace string

	// Dashboard handles — empty unless the profile installs the
	// dashboard (`all`).
	DashboardService      string
	DashboardKubeEndpoint string
	PortForwardCommand    string
}

func initializeLocals(_ *pulumi.Context, iacInput *kubernetestektonv1alpha1.KubernetesTektonIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesTekton.String(),
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

	profile := spec.GetProfile()
	if profile == "" {
		profile = "all"
	}

	targetNamespace := spec.GetTargetNamespace()
	if targetNamespace == "" {
		targetNamespace = vars.DefaultTargetNamespace
	}

	locals := &Locals{
		KubernetesTekton: target,
		Spec:             spec,
		ResourceName:     target.Metadata.Name,
		Labels:           labels,
		Profile:          profile,
		TargetNamespace:  targetNamespace,
	}

	// The dashboard installs on profile `all` only.
	if profile == "all" {
		locals.DashboardService = vars.DashboardServiceName
		locals.DashboardKubeEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
			vars.DashboardServiceName, targetNamespace, vars.DashboardPort)
		locals.PortForwardCommand = fmt.Sprintf(
			"kubectl port-forward -n %s service/%s %d:%d",
			targetNamespace, vars.DashboardServiceName, vars.DashboardPort, vars.DashboardPort)
	}

	return locals
}
