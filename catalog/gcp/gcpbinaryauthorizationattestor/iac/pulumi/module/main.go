package module

import (
	"github.com/pkg/errors"
	gcpbinaryauthorizationattestorv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbinaryauthorizationattestor/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpbinaryauthorizationattestorv1alpha1.GcpBinaryAuthorizationAttestorIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := attestor(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create the Binary Authorization attestor")
	}

	return nil
}
