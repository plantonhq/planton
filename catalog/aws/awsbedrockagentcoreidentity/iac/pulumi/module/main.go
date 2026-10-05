package module

import (
	"github.com/pkg/errors"
	awsbedrockagentcoreidentityv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockagentcoreidentity/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the AgentCore identity bundle
// (workload identities, credential providers, the Cedar policy engine
// with its policies) and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awsbedrockagentcoreidentityv1alpha1.AwsBedrockAgentCoreIdentityIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := identity(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "agentcore identity")
	}

	return nil
}
