package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	hetznerclouddnszonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud/hetznerclouddnszone/v1"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func zone(
	ctx *pulumi.Context,
	locals *Locals,
	hcloudProvider *hcloud.Provider,
) error {
	spec := locals.HetznerCloudDnsZone.Spec

	zoneArgs := &hcloud.ZoneArgs{
		Name:             pulumi.StringPtr(spec.DomainName),
		Mode:             pulumi.String(spec.Mode.String()),
		Labels:           pulumi.ToStringMap(locals.Labels),
		DeleteProtection: pulumi.BoolPtr(spec.DeleteProtection),
	}

	if spec.Ttl != nil {
		zoneArgs.Ttl = pulumi.IntPtr(int(*spec.Ttl))
	}

	if len(spec.PrimaryNameservers) > 0 {
		nsArgs := make(hcloud.ZonePrimaryNameserverArray, 0, len(spec.PrimaryNameservers))
		for _, ns := range spec.PrimaryNameservers {
			nsArg := hcloud.ZonePrimaryNameserverArgs{
				Address: pulumi.String(ns.Address),
			}
			if ns.Port != nil {
				nsArg.Port = pulumi.IntPtr(int(*ns.Port))
			}
			if ns.TsigAlgorithm != "" {
				nsArg.TsigAlgorithm = pulumi.StringPtr(ns.TsigAlgorithm)
			}
			if ns.TsigKey != "" {
				nsArg.TsigKey = pulumi.StringPtr(ns.TsigKey)
			}
			nsArgs = append(nsArgs, nsArg)
		}
		zoneArgs.PrimaryNameservers = nsArgs
	}

	createdZone, err := hcloud.NewZone(
		ctx,
		"zone",
		zoneArgs,
		pulumi.Provider(hcloudProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create hetzner cloud dns zone")
	}

	zoneIdStr := createdZone.ID().ToStringOutput()

	if err := createRecordSets(ctx, spec.RecordSets, zoneIdStr, hcloudProvider); err != nil {
		return errors.Wrap(err, "failed to create dns record sets")
	}

	ctx.Export(OpZoneId, createdZone.ID())
	ctx.Export(OpNameservers, createdZone.AuthoritativeNameservers.Assigneds())

	return nil
}

// createRecordSets creates an hcloud_zone_rrset for each record set entry.
// Record sets are keyed by {name}-{type} per CG02.
func createRecordSets(
	ctx *pulumi.Context,
	recordSets []*hetznerclouddnszonev1.HetznerCloudDnsZoneSpec_RecordSet,
	zoneId pulumi.StringOutput,
	hcloudProvider *hcloud.Provider,
) error {
	for _, rs := range recordSets {
		records := make(hcloud.ZoneRrsetRecordArray, 0, len(rs.Records))
		for _, rec := range rs.Records {
			recArg := hcloud.ZoneRrsetRecordArgs{
				Value: pulumi.String(rec.Value.GetValue()),
			}
			if rec.Comment != "" {
				recArg.Comment = pulumi.StringPtr(rec.Comment)
			}
			records = append(records, recArg)
		}

		rrsetArgs := &hcloud.ZoneRrsetArgs{
			Zone:    zoneId,
			Name:    pulumi.StringPtr(rs.Name),
			Type:    pulumi.String(rs.Type),
			Records: records,
		}

		if rs.Ttl != nil {
			rrsetArgs.Ttl = pulumi.IntPtr(int(*rs.Ttl))
		}

		resourceName := fmt.Sprintf("rrset-%s-%s", sanitizeDnsName(rs.Name), strings.ToLower(rs.Type))
		if _, err := hcloud.NewZoneRrset(
			ctx,
			resourceName,
			rrsetArgs,
			pulumi.Provider(hcloudProvider),
		); err != nil {
			return errors.Wrapf(err, "failed to create rrset %s/%s", rs.Name, rs.Type)
		}
	}

	return nil
}

// sanitizeDnsName converts a DNS record name into a Pulumi-safe resource
// name component by replacing special characters.
func sanitizeDnsName(name string) string {
	switch name {
	case "@":
		return "at"
	case "*":
		return "wildcard"
	default:
		r := strings.NewReplacer(".", "-", "/", "-", ":", "-")
		return r.Replace(name)
	}
}
