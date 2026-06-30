package module

import (
	"github.com/pkg/errors"
	cloudflared1databasev1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflared1database/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—kept small to mirror a Terraform module’s main.tf.
func Resources(
	ctx *pulumi.Context,
	stackInput *cloudflared1databasev1.CloudflareD1DatabaseStackInput,
) error {
	// 1.  Prepare locals (metadata, credentials, etc.).
	locals := initializeLocals(ctx, stackInput)

	// 2.  Instantiate a Cloudflare provider from the supplied credential.
	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	// 3.  Create the D1 database.
	if _, err := database(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare d1 database")
	}

	return nil
}
