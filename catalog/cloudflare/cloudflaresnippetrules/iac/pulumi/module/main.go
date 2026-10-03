package module

import (
	"github.com/pkg/errors"
	cloudflaresnippetrulesv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflaresnippetrules/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflaresnippetrulesv1alpha1.CloudflareSnippetRulesIacInput,
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

	// 3. Apply the zone's snippet routing table.
	if err := snippetRules(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to apply snippet rules")
	}

	return nil
}
