package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/containerinfra"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func clusterTemplate(ctx *pulumi.Context, locals *Locals, openstackProvider *openstack.Provider) error {
	spec := locals.OpenStackContainerClusterTemplate.Spec
	templateName := locals.OpenStackContainerClusterTemplate.Metadata.Name

	templateArgs := &containerinfra.ClusterTemplateArgs{
		Name:  pulumi.String(templateName),
		Coe:   pulumi.String(spec.Coe),
		Image: pulumi.String(locals.Image),
	}

	if locals.Keypair != "" {
		templateArgs.KeypairId = pulumi.StringPtr(locals.Keypair)
	}
	if locals.ExternalNetwork != "" {
		templateArgs.ExternalNetworkId = pulumi.StringPtr(locals.ExternalNetwork)
	}
	if locals.FixedNetwork != "" {
		templateArgs.FixedNetwork = pulumi.StringPtr(locals.FixedNetwork)
	}
	if locals.FixedSubnet != "" {
		templateArgs.FixedSubnet = pulumi.StringPtr(locals.FixedSubnet)
	}
	if spec.NetworkDriver != "" {
		templateArgs.NetworkDriver = pulumi.StringPtr(spec.NetworkDriver)
	}
	if spec.VolumeDriver != "" {
		templateArgs.VolumeDriver = pulumi.StringPtr(spec.VolumeDriver)
	}
	if spec.DnsNameserver != "" {
		templateArgs.DnsNameserver = pulumi.StringPtr(spec.DnsNameserver)
	}
	if spec.DockerVolumeSize != nil {
		templateArgs.DockerVolumeSize = pulumi.IntPtr(int(spec.GetDockerVolumeSize()))
	}
	if spec.Flavor != "" {
		templateArgs.Flavor = pulumi.StringPtr(spec.Flavor)
	}
	if spec.MasterFlavor != "" {
		templateArgs.MasterFlavor = pulumi.StringPtr(spec.MasterFlavor)
	}
	if spec.FloatingIpEnabled != nil {
		templateArgs.FloatingIpEnabled = pulumi.BoolPtr(spec.GetFloatingIpEnabled())
	}
	if spec.MasterLbEnabled != nil {
		templateArgs.MasterLbEnabled = pulumi.BoolPtr(spec.GetMasterLbEnabled())
	}
	if spec.TlsDisabled != nil {
		templateArgs.TlsDisabled = pulumi.BoolPtr(spec.GetTlsDisabled())
	}
	if len(spec.Labels) > 0 {
		labels := pulumi.StringMap{}
		for k, v := range spec.Labels {
			labels[k] = pulumi.String(v)
		}
		templateArgs.Labels = labels
	}
	if spec.Region != "" {
		templateArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdTemplate, err := containerinfra.NewClusterTemplate(
		ctx,
		strings.ToLower(templateName),
		templateArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create container cluster template")
	}

	ctx.Export(OpTemplateId, createdTemplate.ID())
	ctx.Export(OpName, createdTemplate.Name)
	ctx.Export(OpCoe, createdTemplate.Coe)
	ctx.Export(OpRegion, createdTemplate.Region)

	return nil
}
