package module

import (
	"fmt"
	"strconv"

	kubernetesrabbitmqoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesrabbitmqoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	KubernetesRabbitMqOperator *kubernetesrabbitmqoperatorv1alpha1.KubernetesRabbitMqOperator
	Spec                       *kubernetesrabbitmqoperatorv1alpha1.KubernetesRabbitMqOperatorSpec

	// ResourceName keys the applied manifest bundle in the Pulumi state.
	ResourceName string

	// Labels tie the install back to the Planton resource. The manifest's
	// own documents keep their upstream labels untouched (faithful
	// apply); these are used on the module's own state identity only.
	Labels map[string]string

	// MetricsEndpoint is the in-cluster Prometheus metrics endpoint of
	// the operator (release-manifest Service + port).
	MetricsEndpoint string
}

func initializeLocals(_ *pulumi.Context, iacInput *kubernetesrabbitmqoperatorv1alpha1.KubernetesRabbitMqOperatorIacInput) *Locals {
	target := iacInput.Target

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesRabbitMqOperator.String(),
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

	return &Locals{
		KubernetesRabbitMqOperator: target,
		Spec:                       target.Spec,
		ResourceName:               target.Metadata.Name,
		Labels:                     labels,
		MetricsEndpoint: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d/metrics",
			vars.MetricsServiceName, vars.Namespace, vars.MetricsPort),
	}
}
