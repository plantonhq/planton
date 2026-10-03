package module

import (
	"github.com/pkg/errors"
	awscertv1 "github.com/plantonhq/planton/catalog/aws/awscertmanagercert/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the aws_cert_manager_cert Pulumi module.
// It prepares context, configures the AWS provider, and calls certManagerCert().
func Resources(ctx *pulumi.Context, iacInput *awscertv1.AwsCertManagerCertIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsCertManagerCert.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Call the core logic for ACM certificate + DNS validation setup.
	if err := certManagerCert(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create aws cert manager cert resource")
	}

	return nil
}
