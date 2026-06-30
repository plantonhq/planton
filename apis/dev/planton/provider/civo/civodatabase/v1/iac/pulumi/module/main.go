package module

import (
	"github.com/pkg/errors"
	civodatabasev1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civodatabase/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/civo/pulumicivoprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi entry‑point that Planton invokes.
func Resources(
	ctx *pulumi.Context,
	stackInput *civodatabasev1.CivoDatabaseStackInput,
) error {
	// 1. Collect useful references from stack‑input.
	locals := initializeLocals(ctx, stackInput)

	// 2. Build a Civo provider from the supplied credential.
	civoProvider, err := pulumicivoprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup Civo provider")
	}

	// 3. Provision the managed database.
	if _, err := database(ctx, locals, civoProvider); err != nil {
		return errors.Wrap(err, "failed to create Civo database")
	}

	return nil
}
