package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrustgatewaypolicyv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustgatewaypolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarezerotrustgatewaypolicyv1alpha1.CloudflareZeroTrustGatewayPolicyIacInput,
) error {
	// 1. Prepare locals (metadata, credentials).
	locals := initializeLocals(ctx, iacInput)

	// 2. Create a Cloudflare provider from the supplied credential.
	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	// 3. Create the Gateway policy.
	if err := gatewayPolicy(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create gateway policy")
	}

	return nil
}
