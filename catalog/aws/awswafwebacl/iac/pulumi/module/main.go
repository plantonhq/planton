package module

import (
	"github.com/pkg/errors"
	awswafwebaclv1alpha1 "github.com/plantonhq/planton/catalog/aws/awswafwebacl/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the primary entry point for the AwsWafWebAcl Pulumi module.
// It creates the Web ACL with rules and optional logging configuration.
func Resources(ctx *pulumi.Context, iacInput *awswafwebaclv1alpha1.AwsWafWebAclIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.WebAcl.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdWebAcl, err := webAcl(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create WAF Web ACL")
	}

	if locals.WebAcl.Spec.Logging != nil {
		if err := logging(ctx, locals, provider, createdWebAcl); err != nil {
			return errors.Wrap(err, "failed to configure WAF logging")
		}
	}

	return nil
}
