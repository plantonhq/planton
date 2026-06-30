package module

import (
	"github.com/pkg/errors"
	cloudflarer2bucketv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarer2bucket/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	stackInput *cloudflarer2bucketv1.CloudflareR2BucketStackInput,
) error {
	// 1. Prepare locals.
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Cloudflare provider from the supplied credential.
	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup Cloudflare provider")
	}

	// 3. Create the bucket (and optional domain).
	if _, err := bucket(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create R2 bucket")
	}

	return nil
}
