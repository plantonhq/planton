package module

import (
	"github.com/pkg/errors"
	gcpbinaryauthorizationpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbinaryauthorizationpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpbinaryauthorizationpolicyv1alpha1.GcpBinaryAuthorizationPolicyStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := policy(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to apply the Binary Authorization policy")
	}

	return nil
}
