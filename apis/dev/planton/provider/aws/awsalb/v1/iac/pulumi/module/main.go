package module

import (
	"github.com/pkg/errors"
	awsalbv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsalb/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the primary entry point for the aws_alb Pulumi module.
func Resources(ctx *pulumi.Context, stackInput *awsalbv1.AwsAlbStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsAlb.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	albResource, err := alb(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create aws_alb resource")
	}

	// If the user wants DNS, set up Route53 records
	if locals.AwsAlb.Spec.Dns.GetEnabled() {
		if err := dns(ctx, locals, provider, albResource); err != nil {
			return errors.Wrap(err, "failed to configure DNS")
		}
	}

	return nil
}
