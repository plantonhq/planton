package module

import (
	"github.com/pkg/errors"
	civofirewallv1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civofirewall/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/civo/pulumicivoprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry‑point.
func Resources(
	ctx *pulumi.Context,
	stackInput *civofirewallv1.CivoFirewallStackInput,
) error {

	// 1. Prepare locals.
	locals := initializeLocals(ctx, stackInput)

	// 2. Instantiate provider from credential.
	civoProvider, err := pulumicivoprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to set up Civo provider")
	}

	// 3. Provision firewall.
	if _, err := firewall(ctx, locals, civoProvider); err != nil {
		return errors.Wrap(err, "failed to create firewall")
	}

	return nil
}
