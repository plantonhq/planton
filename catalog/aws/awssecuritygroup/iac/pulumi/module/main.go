package module

import (
	"github.com/pkg/errors"
	awssecuritygroupv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssecuritygroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the primary entry point for the aws_security_group Pulumi module.
// It reads the AwsSecurityGroupIacInput, sets up AWS credentials if provided,
// and delegates to the securityGroup() function to create the resource.
func Resources(ctx *pulumi.Context, iacInput *awssecuritygroupv1alpha1.AwsSecurityGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsSecurityGroup.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Create the AWS Security Group resource
	if err := securityGroup(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create aws_security_group resource")
	}

	return nil
}
