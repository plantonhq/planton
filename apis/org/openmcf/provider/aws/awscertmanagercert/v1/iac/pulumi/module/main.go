package module

import (
	"github.com/pkg/errors"
	awscertv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscertmanagercert/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the aws_cert_manager_cert Pulumi module.
// It prepares context, configures the AWS provider, and calls certManagerCert().
func Resources(ctx *pulumi.Context, stackInput *awscertv1.AwsCertManagerCertStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsCertManagerCert.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Call the core logic for ACM certificate + DNS validation setup.
	if err := certManagerCert(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create aws cert manager cert resource")
	}

	return nil
}
