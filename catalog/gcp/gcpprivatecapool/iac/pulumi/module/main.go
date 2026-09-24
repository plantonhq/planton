package module

import (
	"github.com/pkg/errors"
	gcpprivatecapoolv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpprivatecapool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpprivatecapoolv1alpha1.GcpPrivateCaPoolStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := caPool(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create certificate authority service ca pool")
	}

	return nil
}
