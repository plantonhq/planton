package module

import (
	"github.com/pkg/errors"
	awsecrrepov1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsecrrepo/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point for the aws_ecr_repo Pulumi module.
// It initializes locals, configures a provider (default or custom), then calls ecrRepo.
func Resources(ctx *pulumi.Context, stackInput *awsecrrepov1.AwsEcrRepoStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsEcrRepo.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := ecrRepo(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create aws_ecr_repo resource")
	}

	return nil
}
