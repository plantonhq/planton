package module

import (
	openstackprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	openstackcontainerclustertemplatev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackcontainerclustertemplate/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OpenStackProviderConfig           *openstackprovider.OpenStackProviderConfig
	OpenStackContainerClusterTemplate *openstackcontainerclustertemplatev1.OpenStackContainerClusterTemplate
	Image                             string
	Keypair                           string
	ExternalNetwork                   string
	FixedNetwork                      string
	FixedSubnet                       string
}

func initializeLocals(_ *pulumi.Context, stackInput *openstackcontainerclustertemplatev1.OpenStackContainerClusterTemplateStackInput) *Locals {
	locals := &Locals{
		OpenStackContainerClusterTemplate: stackInput.Target,
		OpenStackProviderConfig:           stackInput.ProviderConfig,
	}

	spec := stackInput.Target.Spec

	locals.Image = spec.Image.GetValue()

	if spec.Keypair != nil {
		locals.Keypair = spec.Keypair.GetValue()
	}
	if spec.ExternalNetwork != nil {
		locals.ExternalNetwork = spec.ExternalNetwork.GetValue()
	}
	if spec.FixedNetwork != nil {
		locals.FixedNetwork = spec.FixedNetwork.GetValue()
	}
	if spec.FixedSubnet != nil {
		locals.FixedSubnet = spec.FixedSubnet.GetValue()
	}

	return locals
}
