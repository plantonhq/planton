package module

import (
	"github.com/pkg/errors"
	awssecuritygroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awssecuritygroup/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the primary entry point for the aws_security_group Pulumi module.
// It reads the AwsSecurityGroupStackInput, sets up AWS credentials if provided,
// and delegates to the securityGroup() function to create the resource.
func Resources(ctx *pulumi.Context, stackInput *awssecuritygroupv1.AwsSecurityGroupStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsSecurityGroup.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Create the AWS Security Group resource
	if err := securityGroup(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create aws_security_group resource")
	}

	return nil
}
