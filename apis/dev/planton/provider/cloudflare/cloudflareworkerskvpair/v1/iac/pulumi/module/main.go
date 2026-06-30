package module

import (
	"github.com/pkg/errors"
	cloudflareworkerskvpairv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflareworkerskvpair/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—kept small to mirror a Terraform module's main.tf.
func Resources(
	ctx *pulumi.Context,
	stackInput *cloudflareworkerskvpairv1.CloudflareWorkersKvPairStackInput,
) error {
	locals := initializeLocals(ctx, stackInput)

	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	if _, err := kvPair(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare workers kv pair")
	}

	return nil
}
