package module

import (
	"github.com/pkg/errors"
	gcpkmskeyringv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpkmskeyring/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpkmskeyringv1.GcpKmsKeyRingStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := keyRing(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create kms key ring")
	}

	return nil
}
