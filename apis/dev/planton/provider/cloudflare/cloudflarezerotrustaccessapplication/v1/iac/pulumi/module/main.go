package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrustaccessapplicationv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarezerotrustaccessapplication/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point invoked by the plantonCLI.
func Resources(
	ctx *pulumi.Context,
	stackInput *cloudflarezerotrustaccessapplicationv1.CloudflareZeroTrustAccessApplicationStackInput,
) error {
	// 1. Gather handy references and credentials.
	locals := initializeLocals(ctx, stackInput)

	// 2. Stand‑up a Cloudflare provider from the supplied API token.
	cloudflareProvider, err := pulumicloudflareprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	// 3. Provision the Access Application (and its policy).
	if _, err := application(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare zero trust access application")
	}

	return nil
}
