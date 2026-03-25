package module

import (
	"github.com/pkg/errors"
	cloudflarerulesetv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareruleset/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(
	ctx *pulumi.Context,
	stackInput *cloudflarerulesetv1.CloudflareRulesetStackInput,
) error {
	locals := initializeLocals(ctx, stackInput)

	cloudflareProvider, err := pulumicloudflareprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	if _, err := ruleset(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare ruleset")
	}

	return nil
}
