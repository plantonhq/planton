package module

import (
	"strings"

	"github.com/pkg/errors"
	ocidnszonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocidnszone/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/dns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var zoneTypeMap = map[ocidnszonev1.OciDnsZoneSpec_ZoneType]string{
	ocidnszonev1.OciDnsZoneSpec_primary:   "PRIMARY",
	ocidnszonev1.OciDnsZoneSpec_secondary: "SECONDARY",
}

func dnsZone(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciDnsZone.Spec

	args := &dns.ZoneArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		Name:          pulumi.String(locals.ZoneName),
		ZoneType:      pulumi.String(zoneTypeMap[spec.ZoneType]),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.Scope == ocidnszonev1.OciDnsZoneSpec_scope_private {
		args.Scope = pulumi.String("PRIVATE")
	} else if spec.Scope == ocidnszonev1.OciDnsZoneSpec_global {
		args.Scope = pulumi.String("GLOBAL")
	}

	if spec.ViewId != nil {
		args.ViewId = pulumi.String(spec.ViewId.GetValue())
	}

	if spec.IsDnssecEnabled != nil {
		if *spec.IsDnssecEnabled {
			args.DnssecState = pulumi.String("ENABLED")
		} else {
			args.DnssecState = pulumi.String("DISABLED")
		}
	}

	if len(spec.ExternalMasters) > 0 {
		args.ExternalMasters = buildExternalMasters(spec.ExternalMasters)
	}

	if len(spec.ExternalDownstreams) > 0 {
		args.ExternalDownstreams = buildExternalDownstreams(spec.ExternalDownstreams)
	}

	zone, err := dns.NewZone(ctx, locals.ZoneName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create dns zone")
	}

	ctx.Export(OpZoneId, zone.ID())

	nameservers := zone.Nameservers.ApplyT(func(nsList []dns.ZoneNameserver) string {
		var hostnames []string
		for _, ns := range nsList {
			if ns.Hostname != nil && *ns.Hostname != "" {
				hostnames = append(hostnames, *ns.Hostname)
			}
		}
		return strings.Join(hostnames, ",")
	}).(pulumi.StringOutput)

	ctx.Export(OpNameservers, nameservers)

	return nil
}

func buildExternalMasters(servers []*ocidnszonev1.OciDnsZoneSpec_ExternalServer) dns.ZoneExternalMasterArray {
	var result dns.ZoneExternalMasterArray

	for _, s := range servers {
		master := &dns.ZoneExternalMasterArgs{
			Address: pulumi.String(s.Address),
		}

		if s.Port != nil {
			master.Port = pulumi.Int(int(*s.Port))
		}

		if s.TsigKeyId != "" {
			master.TsigKeyId = pulumi.String(s.TsigKeyId)
		}

		result = append(result, master)
	}

	return result
}

func buildExternalDownstreams(servers []*ocidnszonev1.OciDnsZoneSpec_ExternalServer) dns.ZoneExternalDownstreamArray {
	var result dns.ZoneExternalDownstreamArray

	for _, s := range servers {
		downstream := &dns.ZoneExternalDownstreamArgs{
			Address: pulumi.String(s.Address),
		}

		if s.Port != nil {
			downstream.Port = pulumi.Int(int(*s.Port))
		}

		if s.TsigKeyId != "" {
			downstream.TsigKeyId = pulumi.String(s.TsigKeyId)
		}

		result = append(result, downstream)
	}

	return result
}
