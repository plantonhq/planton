package module

import (
	"github.com/pkg/errors"
	cloudflarednszonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarednszone/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry‑point called by the Planton CLI.
func Resources(
	ctx *pulumi.Context,
	stackInput *cloudflarednszonev1.CloudflareDnsZoneStackInput,
) error {
	// 1. Gather handy references.
	locals := initializeLocals(ctx, stackInput)

	// 2. Build a Pulumi Cloudflare provider from the supplied credential.
	cloudflareProvider, err := pulumicloudflareprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	// 3. Create the DNS zone.
	if _, err := zone(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare dns zone")
	}

	return nil
}
