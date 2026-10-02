package module

import (
	"github.com/pkg/errors"
	gcptpuvmv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcptpuvm/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcptpuvmv1alpha1.GcpTpuVmStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := tpuVm(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create tpu vm")
	}

	return nil
}
