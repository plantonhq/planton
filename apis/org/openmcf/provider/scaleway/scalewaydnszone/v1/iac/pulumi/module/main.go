package module

import (
	"github.com/pkg/errors"
	scalewaydnszonev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewaydnszone/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point that provisions a Scaleway DNS
// zone with optional inline DNS records.
//
// This is a composite resource:
//   - 1x scaleway_domain_zone (the zone itself)
//   - 0..Nx scaleway_domain_record (one per inline record entry)
//
// The zone is created first, then records are created referencing the
// zone's computed name. Stack outputs include the zone name (for
// downstream ScalewayDnsRecord references) and nameservers (for
// domain registrar delegation).
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewaydnszonev1.ScalewayDnsZoneStackInput,
) error {
	// 1. Prepare locals (computed zone name, resolved references).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Scaleway provider from the supplied credential.
	scalewayProvider, err := pulumiscalewayprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup scaleway provider")
	}

	// 3. Create the DNS zone and inline records, export outputs.
	if err := dnsZone(ctx, locals, scalewayProvider); err != nil {
		return errors.Wrap(err, "failed to create dns zone")
	}

	return nil
}
