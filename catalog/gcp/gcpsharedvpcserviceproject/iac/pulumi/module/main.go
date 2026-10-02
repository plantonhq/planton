package module

import (
	"github.com/pkg/errors"
	gcpsharedvpcserviceprojectv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsharedvpcserviceproject/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpsharedvpcserviceprojectv1alpha1.GcpSharedVpcServiceProjectStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := sharedVpcServiceProject(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to attach shared vpc service project")
	}

	return nil
}
