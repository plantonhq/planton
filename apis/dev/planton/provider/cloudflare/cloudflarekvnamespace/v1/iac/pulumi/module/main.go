package module

import (
	"github.com/pkg/errors"
	cloudflarekvnamespacev1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarekvnamespace/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	stackInput *cloudflarekvnamespacev1.CloudflareKvNamespaceStackInput,
) error {
	// 1. Prepare locals (metadata, credentials).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Cloudflare provider from the supplied credential.
	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	// 3. Create the KV namespace.
	if _, err := kvNamespace(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create workers kv namespace")
	}

	return nil
}
