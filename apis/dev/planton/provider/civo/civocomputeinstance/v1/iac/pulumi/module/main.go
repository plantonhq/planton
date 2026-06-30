package module

import (
	"github.com/pkg/errors"
	civocomputeinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civocomputeinstance/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/civo/pulumicivoprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—keeps symmetry with other Planton modules.
func Resources(
	ctx *pulumi.Context,
	stackInput *civocomputeinstancev1.CivoComputeInstanceStackInput,
) error {
	// 1. Prepare locals (metadata, credentials, etc.).
	locals := initializeLocals(ctx, stackInput)

	// 2. Civo provider from supplied credential.
	civoProvider, err := pulumicivoprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup civo provider")
	}

	// 3. Create the compute instance.
	if _, err := instance(ctx, locals, civoProvider); err != nil {
		return errors.Wrap(err, "failed to create compute instance")
	}

	return nil
}
