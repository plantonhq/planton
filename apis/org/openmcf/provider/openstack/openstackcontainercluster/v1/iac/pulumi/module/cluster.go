package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/containerinfra"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func cluster(ctx *pulumi.Context, locals *Locals, openstackProvider *openstack.Provider) error {
	spec := locals.OpenStackContainerCluster.Spec
	clusterName := locals.OpenStackContainerCluster.Metadata.Name

	clusterArgs := &containerinfra.ClusterArgs{
		Name:              pulumi.StringPtr(clusterName),
		ClusterTemplateId: pulumi.String(locals.ClusterTemplate),
	}

	if spec.MasterCount != nil {
		clusterArgs.MasterCount = pulumi.IntPtr(int(spec.GetMasterCount()))
	}
	if spec.NodeCount != nil {
		clusterArgs.NodeCount = pulumi.IntPtr(int(spec.GetNodeCount()))
	}
	if locals.Keypair != "" {
		clusterArgs.Keypair = pulumi.StringPtr(locals.Keypair)
	}
	if spec.Flavor != "" {
		clusterArgs.Flavor = pulumi.StringPtr(spec.Flavor)
	}
	if spec.MasterFlavor != "" {
		clusterArgs.MasterFlavor = pulumi.StringPtr(spec.MasterFlavor)
	}
	if spec.DockerVolumeSize != nil {
		clusterArgs.DockerVolumeSize = pulumi.IntPtr(int(spec.GetDockerVolumeSize()))
	}
	if len(spec.Labels) > 0 {
		labels := pulumi.StringMap{}
		for k, v := range spec.Labels {
			labels[k] = pulumi.String(v)
		}
		clusterArgs.Labels = labels
	}
	if spec.CreateTimeout != nil {
		clusterArgs.CreateTimeout = pulumi.IntPtr(int(spec.GetCreateTimeout()))
	}
	if spec.FloatingIpEnabled != nil {
		clusterArgs.FloatingIpEnabled = pulumi.BoolPtr(spec.GetFloatingIpEnabled())
	}
	if spec.Region != "" {
		clusterArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdCluster, err := containerinfra.NewCluster(
		ctx,
		strings.ToLower(clusterName),
		clusterArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create container cluster")
	}

	// Export standard outputs
	ctx.Export(OpClusterId, createdCluster.ID())
	ctx.Export(OpName, createdCluster.Name)
	ctx.Export(OpApiAddress, createdCluster.ApiAddress)
	ctx.Export(OpCoeVersion, createdCluster.CoeVersion)
	ctx.Export(OpMasterAddresses, createdCluster.MasterAddresses)
	ctx.Export(OpNodeAddresses, createdCluster.NodeAddresses)
	ctx.Export(OpRegion, createdCluster.Region)

	// Export kubeconfig outputs (SENSITIVE)
	ctx.Export(OpKubeconfigRaw, pulumi.ToSecret(
		createdCluster.Kubeconfig.MapIndex(pulumi.String("raw_config")),
	))
	ctx.Export(OpKubeconfigHost, createdCluster.Kubeconfig.MapIndex(pulumi.String("host")))
	ctx.Export(OpKubeconfigClusterCaCert, pulumi.ToSecret(
		createdCluster.Kubeconfig.MapIndex(pulumi.String("cluster_ca_certificate")),
	))
	ctx.Export(OpKubeconfigClientCert, pulumi.ToSecret(
		createdCluster.Kubeconfig.MapIndex(pulumi.String("client_certificate")),
	))
	ctx.Export(OpKubeconfigClientKey, pulumi.ToSecret(
		createdCluster.Kubeconfig.MapIndex(pulumi.String("client_key")),
	))

	return nil
}
