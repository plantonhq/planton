package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// cluster provisions the Kubernetes cluster itself and exports its outputs.
func cluster(
	ctx *pulumi.Context,
	locals *Locals,
	civoProvider *civo.Provider,
) (*civo.KubernetesCluster, error) {

	// 1. Collect tags from the spec.
	var tags pulumi.StringArray
	for _, t := range locals.CivoKubernetesCluster.Spec.Tags {
		tags = append(tags, pulumi.String(t))
	}

	// 3. Build the cluster arguments.
	clusterArgs := &civo.KubernetesClusterArgs{
		Applications:      nil,
		ClusterType:       nil,
		Cni:               nil,
		FirewallId:        nil,
		KubernetesVersion: pulumi.String(locals.CivoKubernetesCluster.Spec.KubernetesVersion),
		Name:              pulumi.String(locals.CivoKubernetesCluster.Spec.ClusterName),
		NetworkId:         pulumi.String(locals.CivoKubernetesCluster.Spec.Network.GetValue()),
		NumTargetNodes:    nil,
		Pools:             nil,
		Region:            pulumi.String(locals.CivoKubernetesCluster.Spec.Region.String()),
		Tags:              nil,
		TargetNodesSize:   nil,
		WriteKubeconfig:   nil,
	}

	// 4. Create the cluster.
	createdCluster, err := civo.NewKubernetesCluster(
		ctx,
		"cluster",
		clusterArgs,
		pulumi.Provider(civoProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Civo Kubernetes cluster")
	}

	// 5. Export required stack outputs.
	ctx.Export(OpClusterId, createdCluster.ID())
	ctx.Export(OpApiServerEndpoint, createdCluster.ApiEndpoint)
	ctx.Export(OpKubeconfig, createdCluster.Kubeconfig)

	return createdCluster, nil
}
