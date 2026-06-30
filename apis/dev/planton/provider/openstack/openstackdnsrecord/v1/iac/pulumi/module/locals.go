package module

import (
	openstackprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openstack"
	openstackdnsrecordv1 "github.com/plantonhq/planton/apis/dev/planton/provider/openstack/openstackdnsrecord/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OpenStackProviderConfig *openstackprovider.OpenStackProviderConfig
	OpenStackDnsRecord      *openstackdnsrecordv1.OpenStackDnsRecord
	ZoneId                  string
}

func initializeLocals(_ *pulumi.Context, stackInput *openstackdnsrecordv1.OpenStackDnsRecordStackInput) *Locals {
	return &Locals{
		OpenStackDnsRecord:      stackInput.Target,
		OpenStackProviderConfig: stackInput.ProviderConfig,
		ZoneId:                  stackInput.Target.Spec.ZoneId.GetValue(),
	}
}
