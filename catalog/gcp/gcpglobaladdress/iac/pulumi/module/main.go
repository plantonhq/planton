package module

import (
	"github.com/pkg/errors"
	gcpglobaladdressv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpglobaladdress/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpglobaladdressv1alpha1.GcpGlobalAddressIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := globalAddress(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create global address")
	}

	return nil
}
