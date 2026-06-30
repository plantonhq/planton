package module

import (
	"github.com/pkg/errors"
	civovolumev1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civovolume/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/civo/pulumicivoprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—mirrors the pattern used in digital_ocean_volume.
func Resources(
	ctx *pulumi.Context,
	stackInput *civovolumev1.CivoVolumeStackInput,
) error {
	// 1 . Prepare locals (metadata, labels, credentials, etc.).
	locals := initializeLocals(ctx, stackInput)

	// 2 . Create a Civo provider from the supplied credential.
	civoProvider, err := pulumicivoprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup civo provider")
	}

	// 3 . Create the Volume.
	if _, err := volume(ctx, locals, civoProvider); err != nil {
		return errors.Wrap(err, "failed to create civo volume")
	}

	return nil
}
