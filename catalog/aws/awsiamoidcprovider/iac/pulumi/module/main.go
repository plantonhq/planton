package module

import (
	"github.com/pkg/errors"
	awsiamoidcproviderv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsiamoidcprovider/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awsiamoidcproviderv1alpha1.AwsIamOidcProviderIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsIamOidcProvider.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// create the IAM OIDC provider
	if err := oidcProvider(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create iam oidc provider resource")
	}

	return nil
}
