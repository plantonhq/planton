package module

import (
	"github.com/pkg/errors"
	civodnsrecordv1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civodnsrecord/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/civo/pulumicivoprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point called by the Planton CLI.
func Resources(
	ctx *pulumi.Context,
	stackInput *civodnsrecordv1.CivoDnsRecordStackInput,
) error {
	// 1. Gather handy references.
	locals := initializeLocals(ctx, stackInput)

	// 2. Build a Pulumi Civo provider from the supplied credential.
	civoProvider, err := pulumicivoprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup civo provider")
	}

	// 3. Create the DNS record.
	if _, err := dnsRecord(ctx, locals, civoProvider); err != nil {
		return errors.Wrap(err, "failed to create civo dns record")
	}

	return nil
}
