package module

import (
	"github.com/pkg/errors"
	cloudflarednszonev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflarednszone/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dnssec enables DNSSEC on the zone. The DS material Cloudflare computes is
// surfaced through the zone's stack outputs for entry at the registrar.
func dnssec(
	ctx *pulumi.Context,
	resourceName string,
	zone *cloudflare.Zone,
	spec *cloudflarednszonev1.CloudflareDnsZoneDnssec,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.ZoneDnssec, error) {
	created, err := cloudflare.NewZoneDnssec(
		ctx,
		resourceName+"-dnssec",
		&cloudflare.ZoneDnssecArgs{
			ZoneId:            zone.ID(),
			Status:            pulumi.String("active"),
			DnssecMultiSigner: pulumi.Bool(spec.MultiSigner),
			DnssecPresigned:   pulumi.Bool(spec.Presigned),
			DnssecUseNsec3:    pulumi.Bool(spec.UseNsec3),
		},
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to enable dnssec")
	}
	return created, nil
}
