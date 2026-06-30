package module

import (
	"github.com/pkg/errors"
	civokubernetesnodepoolv1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civokubernetesnodepool/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/civo/pulumicivoprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the single public entry‑point—mirrors Terraform’s single main.tf.
//
// The noun‑based name follows Planton’s convention (no “create*” verbs).
func Resources(
	ctx *pulumi.Context,
	stackInput *civokubernetesnodepoolv1.CivoKubernetesNodePoolStackInput,
) error {
	// 1  Resolve locals.
	locals := initializeLocals(ctx, stackInput)

	// 2  Spin up a Civo provider from the credential.
	civoProvider, err := pulumicivoprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to set up civo provider")
	}

	// 3  Provision the node‑pool.
	if _, err := nodePool(ctx, locals, civoProvider); err != nil {
		return errors.Wrap(err, "failed to create civo kubernetes node pool")
	}

	return nil
}
