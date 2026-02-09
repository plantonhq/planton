package module

import (
	"github.com/pkg/errors"
	openstackdnsrecordv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackdnsrecord/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/openstack/pulumiopenstackprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *openstackdnsrecordv1.OpenStackDnsRecordStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	openstackProvider, err := pulumiopenstackprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup openstack provider")
	}
	if err := dnsRecord(ctx, locals, openstackProvider); err != nil {
		return errors.Wrap(err, "failed to create openstack dns record")
	}
	return nil
}
