package module

import (
	"github.com/pkg/errors"
	cloudflarednszonev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarednszone/v1alpha1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dnssec enables DNSSEC on the zone. The DS material Cloudflare computes is
// surfaced through the zone's outputs for entry at the registrar.
func dnssec(
	ctx *pulumi.Context,
	resourceName string,
	zone *cloudflare.Zone,
	spec *cloudflarednszonev1alpha1.CloudflareDnsZoneDnssec,
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
