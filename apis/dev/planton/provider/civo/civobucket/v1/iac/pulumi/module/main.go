package module

import (
	"github.com/pkg/errors"
	civobucketv1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civobucket/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/civo/pulumicivoprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—mirrors the pattern used in other Planton modules.
func Resources(
	ctx *pulumi.Context,
	stackInput *civobucketv1.CivoBucketStackInput,
) error {
	// 1. Prepare locals (metadata, labels, credentials, etc.).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Civo provider from the supplied credential.
	civoProvider, err := pulumicivoprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup civo provider")
	}

	// 3. Create the bucket.
	if _, err := bucket(ctx, locals, civoProvider); err != nil {
		return errors.Wrap(err, "failed to create bucket")
	}

	return nil
}
