package module

import (
	"github.com/pkg/errors"
	civodnszonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civodnszone/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/civo/pulumicivoprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the single public entry‑point that Planton’s CLI invokes.
func Resources(
	ctx *pulumi.Context,
	stackInput *civodnszonev1.CivoDnsZoneStackInput,
) error {
	// 1. Consolidate frequently‑used values into a Locals struct.
	locals := initializeLocals(ctx, stackInput)

	// 2. Instantiate a Civo provider from the incoming credential.
	civoProvider, err := pulumicivoprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to set up Civo provider")
	}

	// 3. Provision the DNS zone (domain + records).
	if _, err := dnsZone(ctx, locals, civoProvider); err != nil {
		return errors.Wrap(err, "failed to create Civo DNS zone")
	}

	return nil
}
