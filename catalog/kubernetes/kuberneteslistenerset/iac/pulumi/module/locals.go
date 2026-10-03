package module

import (
	kuberneteslistenersetv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteslistenerset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"strconv"
)

// Locals holds the resolved inputs the module operates on: the full target
// resource plus the scalar identifiers used for the resource name, namespace,
// labels, and outputs.
type Locals struct {
	KubernetesListenerSet *kuberneteslistenersetv1alpha1.KubernetesListenerSet
	ListenerSetName       string
	Namespace             string
	GatewayName           string
	Labels                map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *kuberneteslistenersetv1alpha1.KubernetesListenerSetIacInput) *Locals {
	target := iacInput.Target
	metadata := target.Metadata
	spec := target.Spec

	// namespace and the parent Gateway name are StringValueOrRef foreign
	// keys. The platform middleware resolves valueFrom references to literal
	// strings before the IaC module runs, so GetValue() returns the resolved
	// values.
	namespace := spec.GetNamespace().GetValue()
	gatewayName := spec.GetParentRef().GetName().GetValue()

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesListenerSet.String(),
	}
	if metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = metadata.Id
	}
	if metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = metadata.Org
	}
	if metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = metadata.Env
	}

	return &Locals{
		KubernetesListenerSet: target,
		ListenerSetName:       metadata.Name,
		Namespace:             namespace,
		GatewayName:           gatewayName,
		Labels:                labels,
	}
}
