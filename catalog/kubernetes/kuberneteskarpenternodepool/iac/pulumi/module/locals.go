package module

import (
	"strconv"

	kuberneteskarpenternodepoolv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskarpenternodepool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the resolved inputs the module operates on: the full target
// resource plus the scalar identifiers used for the resource name, labels,
// and outputs. NodePool is cluster-scoped, so there is no namespace.
type Locals struct {
	KubernetesKarpenterNodePool *kuberneteskarpenternodepoolv1alpha1.KubernetesKarpenterNodePool
	NodePoolName                string
	Labels                      map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *kuberneteskarpenternodepoolv1alpha1.KubernetesKarpenterNodePoolIacInput) *Locals {
	target := iacInput.Target
	metadata := target.Metadata

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesKarpenterNodePool.String(),
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
		KubernetesKarpenterNodePool: target,
		NodePoolName:                metadata.Name,
		Labels:                      labels,
	}
}
