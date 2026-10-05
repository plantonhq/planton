package module

import (
	"github.com/pkg/errors"
	awsbedrockagentcoreevaluationv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockagentcoreevaluation/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/bedrock"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the AgentCore Evaluations bundle
// (evaluators, harnesses, online evaluation configs) and exports
// outputs.
func Resources(ctx *pulumi.Context, iacInput *awsbedrockagentcoreevaluationv1alpha1.AwsBedrockAgentCoreEvaluationIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Evaluators first: online configs may reference in-bundle
	// evaluators by name, and the resolver below hands their created
	// IDs to the config renders.
	createdEvaluators, err := evaluators(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "agentcore evaluators")
	}

	if err := harnesses(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "agentcore harnesses")
	}

	if err := onlineConfigs(ctx, locals, provider, createdEvaluators); err != nil {
		return errors.Wrap(err, "agentcore online evaluation configs")
	}

	return nil
}

// createdEvaluator carries what online configs need from an in-bundle
// evaluator: its AWS-generated ID output.
type createdEvaluator struct {
	resource *bedrock.AgentcoreEvaluator
}
