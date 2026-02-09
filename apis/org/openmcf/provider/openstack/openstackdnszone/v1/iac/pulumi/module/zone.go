package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/dns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dnsZone provisions the OpenStack Designate DNS zone, creates inline
// record sets, and exports outputs.
func dnsZone(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackDnsZone.Spec

	zoneArgs := &dns.ZoneArgs{
		Name: pulumi.String(spec.DomainName),
	}

	// Set email if provided.
	if spec.Email != "" {
		zoneArgs.Email = pulumi.StringPtr(spec.Email)
	}

	// Set description if provided.
	if spec.Description != "" {
		zoneArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Set TTL if explicitly provided.
	if spec.Ttl != nil {
		zoneArgs.Ttl = pulumi.IntPtr(int(spec.GetTtl()))
	}

	// Set zone type if provided.
	if spec.Type != "" {
		zoneArgs.Type = pulumi.StringPtr(spec.Type)
	}

	// Set masters for SECONDARY zones.
	if len(spec.Masters) > 0 {
		masters := make(pulumi.StringArray, len(spec.Masters))
		for i, m := range spec.Masters {
			masters[i] = pulumi.String(m)
		}
		zoneArgs.Masters = masters
	}

	// Set region override if provided.
	if spec.Region != "" {
		zoneArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdZone, err := dns.NewZone(
		ctx,
		strings.ToLower(locals.OpenStackDnsZone.Metadata.Name),
		zoneArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create dns zone")
	}

	// Create inline records. Each record is a separate RecordSet resource,
	// keyed by recordType + recordName for stable naming.
	for _, record := range spec.Records {
		recordTypeName := record.RecordType.String()
		resourceKey := fmt.Sprintf("%s-%s", strings.ToLower(recordTypeName), record.RecordName)

		recordValues := make(pulumi.StringArray, len(record.Values))
		for i, v := range record.Values {
			recordValues[i] = pulumi.String(v)
		}

		recordArgs := &dns.RecordSetArgs{
			ZoneId:  createdZone.ID(),
			Name:    pulumi.String(record.RecordName),
			Type:    pulumi.String(recordTypeName),
			Records: recordValues,
			Ttl:     pulumi.Int(int(record.GetTtl())),
		}

		// Set region on the record to match the zone.
		if spec.Region != "" {
			recordArgs.Region = pulumi.StringPtr(spec.Region)
		}

		_, err := dns.NewRecordSet(
			ctx,
			resourceKey,
			recordArgs,
			pulumi.Provider(openstackProvider),
			pulumi.DependsOn([]pulumi.Resource{createdZone}),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create dns record %s", resourceKey)
		}
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpZoneId, createdZone.ID())
	ctx.Export(OpZoneName, createdZone.Name)
	ctx.Export(OpRegion, createdZone.Region)

	return nil
}
