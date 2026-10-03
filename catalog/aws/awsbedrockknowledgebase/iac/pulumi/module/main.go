package module

import (
	"github.com/pkg/errors"
	awsbedrockknowledgebasev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockknowledgebase/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the Bedrock knowledge base and its
// folded data sources and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awsbedrockknowledgebasev1alpha1.AwsBedrockKnowledgeBaseIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := knowledgeBase(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "bedrock knowledge base")
	}

	return nil
}
