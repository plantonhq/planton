package module

import (
	openstackprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	openstackcontainerclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackcontainercluster/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OpenStackProviderConfig   *openstackprovider.OpenStackProviderConfig
	OpenStackContainerCluster *openstackcontainerclusterv1.OpenStackContainerCluster
	ClusterTemplate           string
	Keypair                   string
}

func initializeLocals(_ *pulumi.Context, stackInput *openstackcontainerclusterv1.OpenStackContainerClusterStackInput) *Locals {
	locals := &Locals{
		OpenStackContainerCluster: stackInput.Target,
		OpenStackProviderConfig:   stackInput.ProviderConfig,
	}

	spec := stackInput.Target.Spec

	// Required FK: cluster template
	locals.ClusterTemplate = spec.ClusterTemplate.GetValue()

	// Optional FK: keypair
	if spec.Keypair != nil {
		locals.Keypair = spec.Keypair.GetValue()
	}

	return locals
}
