package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// nodePool provisions the Kubernetes node‑pool and exports its IDs.
func nodePool(
	ctx *pulumi.Context,
	locals *Locals,
	civoProvider *civo.Provider,
) (*civo.KubernetesNodePool, error) {

	// --- translate proto fields directly into Pulumi args --------------------

	// Labels
	labels := pulumi.StringMap{}
	for k, v := range locals.CivoLabels {
		labels[k] = pulumi.String(v)
	}

	// Tags
	tags := pulumi.StringArray{}
	for _, t := range locals.CivoKubernetesNodePool.Spec.Tags {
		tags = append(tags, pulumi.String(t))
	}

	// Node‑pool args: keep field names close to Terraform’s civo_kubernetes_node_pool.
	nodePoolArgs := &civo.KubernetesNodePoolArgs{
		ClusterId: pulumi.String(locals.CivoKubernetesNodePool.Spec.Cluster.GetValue()),
		Size:      pulumi.String(locals.CivoKubernetesNodePool.Spec.Size),
		Labels:    labels,
	}

	// --- create the resource -------------------------------------------------
	createdNodePool, err := civo.NewKubernetesNodePool(
		ctx,
		"node_pool",
		nodePoolArgs,
		pulumi.Provider(civoProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create civo kubernetes node pool")
	}

	// --- export stack outputs -----------------------------------------------
	ctx.Export(OpNodePoolId, createdNodePool.ID())
	ctx.Export(OpNodeIds, createdNodePool.InstanceNames)

	return createdNodePool, nil
}
