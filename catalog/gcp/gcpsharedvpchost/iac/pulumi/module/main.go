package module

import (
	"github.com/pkg/errors"
	gcpsharedvpchostv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsharedvpchost/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpsharedvpchostv1alpha1.GcpSharedVpcHostIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := sharedVpcHost(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to enable shared vpc host")
	}

	return nil
}
