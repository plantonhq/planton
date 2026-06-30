package module

import (
	"github.com/pkg/errors"
	awsiamoidcproviderv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsiamoidcprovider/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awsiamoidcproviderv1.AwsIamOidcProviderStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsIamOidcProvider.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// create the IAM OIDC provider
	if err := oidcProvider(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create iam oidc provider resource")
	}

	return nil
}
