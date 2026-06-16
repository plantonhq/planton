package module

import (
	"github.com/pkg/errors"
	awsecsservicev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsecsservice/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It reads the AwsEcsServiceStackInput, initializes locals, configures
// an AWS provider, then calls the service(...) function.
func Resources(ctx *pulumi.Context, stackInput *awsecsservicev1.AwsEcsServiceStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsEcsService.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Call the service(...) function to create the AWS ECS Service resource.
	if err := ecsService(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create aws ecs service resource")
	}

	return nil
}
