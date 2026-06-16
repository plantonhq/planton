package module

import (
	"github.com/pkg/errors"
	awswafwebaclv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awswafwebacl/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the primary entry point for the AwsWafWebAcl Pulumi module.
// It creates the Web ACL with rules and optional logging configuration.
func Resources(ctx *pulumi.Context, stackInput *awswafwebaclv1.AwsWafWebAclStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.WebAcl.Spec.Region)
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
