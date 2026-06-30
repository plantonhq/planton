package module

import (
	"github.com/pkg/errors"
	civovpcv1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civovpc/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/civo/pulumicivoprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—mirrors the pattern used in other providers.
func Resources(
	ctx *pulumi.Context,
	stackInput *civovpcv1.CivoVpcStackInput,
) error {
	// 1. Prepare locals (metadata, labels, credentials, etc.).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Civo provider from the supplied credential.
	civoProvider, err := pulumicivoprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup civo provider")
	}

	// 3. Create the VPC network.
	if _, err := vpc(ctx, locals, civoProvider); err != nil {
		return errors.Wrap(err, "failed to create vpc")
	}

	return nil
}
