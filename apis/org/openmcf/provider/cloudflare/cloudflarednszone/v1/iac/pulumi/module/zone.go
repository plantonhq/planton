package module

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	cloudflarednszonev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflarednszone/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// zone provisions the Cloudflare zone (plus its folded DNS settings, DNSSEC, and
// inline records) and exports outputs.
func zone(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.Zone, error) {
	spec := locals.CloudflareDnsZone.Spec
	resourceName := strings.ToLower(locals.CloudflareDnsZone.Metadata.Name)

	// Zone type defaults to "full" (guard the proto's unspecified zero value).
	zoneType := "full"
	if spec.Type != cloudflarednszonev1.CloudflareDnsZoneSpec_zone_type_unspecified {
		zoneType = spec.Type.String()
	}

	zoneArgs := &cloudflare.ZoneArgs{
		Account: cloudflare.ZoneAccountArgs{
			Id: pulumi.String(spec.AccountId),
		},
		Name:   pulumi.String(spec.ZoneName),
		Paused: pulumi.BoolPtr(spec.Paused),
		Type:   pulumi.String(zoneType),
	}
	if len(spec.VanityNameServers) > 0 {
		vns := make(pulumi.StringArray, 0, len(spec.VanityNameServers))
		for _, ns := range spec.VanityNameServers {
			vns = append(vns, pulumi.String(ns))
		}
		zoneArgs.VanityNameServers = vns
	}

	createdZone, err := cloudflare.NewZone(
		ctx,
		resourceName,
		zoneArgs,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare zone")
	}

	// Zone-wide DNS settings (folded onto the zone).
	if spec.DnsSettings != nil {
		if err := dnsSettings(ctx, resourceName, createdZone, spec.DnsSettings, cloudflareProvider); err != nil {
			return nil, err
		}
	}

	// DNSSEC (folded onto the zone), created only when enabled.
	var createdDnssec *cloudflare.ZoneDnssec
	if spec.Dnssec != nil && spec.Dnssec.Enabled {
		createdDnssec, err = dnssec(ctx, resourceName, createdZone, spec.Dnssec, cloudflareProvider)
		if err != nil {
			return nil, err
		}
	}

	// Inline DNS records.
	if len(spec.Records) > 0 {
		if err := records(ctx, createdZone, spec.Records, cloudflareProvider); err != nil {
			return nil, errors.Wrap(err, "failed to create dns records")
		}
	}

	// Export outputs.
	ctx.Export(OpZoneId, createdZone.ID())
	ctx.Export(OpNameservers, createdZone.NameServers)
	ctx.Export(OpStatus, createdZone.Status)
	exportDnssecOutputs(ctx, createdDnssec)

	return createdZone, nil
}

// exportDnssecOutputs publishes the DS material when DNSSEC is enabled, or empty
// strings otherwise, so the stack output contract is always satisfied.
func exportDnssecOutputs(ctx *pulumi.Context, d *cloudflare.ZoneDnssec) {
	if d == nil {
		empty := pulumi.String("")
		ctx.Export(OpDnssecStatus, empty)
		ctx.Export(OpDnssecDs, empty)
		ctx.Export(OpDnssecDigest, empty)
		ctx.Export(OpDnssecDigestType, empty)
		ctx.Export(OpDnssecDigestAlgorithm, empty)
		ctx.Export(OpDnssecAlgorithm, empty)
		ctx.Export(OpDnssecKeyTag, empty)
		ctx.Export(OpDnssecPublicKey, empty)
		ctx.Export(OpDnssecFlags, empty)
		return
	}
	floatToString := func(o pulumi.Float64Output) pulumi.StringOutput {
		return o.ApplyT(func(f float64) string { return strconv.FormatInt(int64(f), 10) }).(pulumi.StringOutput)
	}
	ctx.Export(OpDnssecStatus, d.Status)
	ctx.Export(OpDnssecDs, d.Ds)
	ctx.Export(OpDnssecDigest, d.Digest)
	ctx.Export(OpDnssecDigestType, d.DigestType)
	ctx.Export(OpDnssecDigestAlgorithm, d.DigestAlgorithm)
	ctx.Export(OpDnssecAlgorithm, d.Algorithm)
	ctx.Export(OpDnssecKeyTag, floatToString(d.KeyTag))
	ctx.Export(OpDnssecPublicKey, d.PublicKey)
	ctx.Export(OpDnssecFlags, floatToString(d.Flags))
}
