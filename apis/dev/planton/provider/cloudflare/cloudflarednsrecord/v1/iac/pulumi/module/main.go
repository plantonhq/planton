package module

import (
	"github.com/pkg/errors"
	cloudflarednsrecordv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarednsrecord/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point called by the Planton CLI.
func Resources(
	ctx *pulumi.Context,
	stackInput *cloudflarednsrecordv1.CloudflareDnsRecordStackInput,
) error {
	// 1. Gather handy references.
	locals := initializeLocals(ctx, stackInput)

	// 2. Build a Pulumi Cloudflare provider from the supplied credential.
	cloudflareProvider, err := pulumicloudflareprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	// 3. Create the DNS record.
	if _, err := dnsRecord(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare dns record")
	}

	return nil
}
