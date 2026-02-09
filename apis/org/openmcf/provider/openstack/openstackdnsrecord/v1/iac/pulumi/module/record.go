package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/dns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func dnsRecord(ctx *pulumi.Context, locals *Locals, openstackProvider *openstack.Provider) error {
	spec := locals.OpenStackDnsRecord.Spec
	recordName := locals.OpenStackDnsRecord.Metadata.Name

	recordValues := make(pulumi.StringArray, len(spec.Values))
	for i, v := range spec.Values {
		recordValues[i] = pulumi.String(v)
	}

	recordArgs := &dns.RecordSetArgs{
		ZoneId:  pulumi.String(locals.ZoneId),
		Name:    pulumi.String(spec.RecordName),
		Type:    pulumi.String(spec.Type.String()),
		Records: recordValues,
	}

	if spec.Ttl != nil {
		recordArgs.Ttl = pulumi.IntPtr(int(spec.GetTtl()))
	}
	if spec.Description != "" {
		recordArgs.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.Region != "" {
		recordArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdRecord, err := dns.NewRecordSet(
		ctx,
		strings.ToLower(recordName),
		recordArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create dns record")
	}

	ctx.Export(OpRecordsetId, createdRecord.ID())
	ctx.Export(OpFqdn, createdRecord.Name)
	ctx.Export(OpRecordType, createdRecord.Type)
	ctx.Export(OpZoneId, createdRecord.ZoneId)
	ctx.Export(OpRegion, createdRecord.Region)

	return nil
}
