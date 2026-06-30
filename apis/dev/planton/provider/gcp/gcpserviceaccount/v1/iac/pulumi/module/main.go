package module

import (
	"github.com/pkg/errors"
	gcpserviceaccountv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpserviceaccount/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the Service Account, optional key, IAM bindings, and exports outputs.
func Resources(
	ctx *pulumi.Context,
	stackInput *gcpserviceaccountv1.GcpServiceAccountStackInput,
) error {

	// Gather "locals" (mirrors Terraform locals {} convention).
	locals := initializeLocals(ctx, stackInput)

	// Create gcp provider using credentials from the input
	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	// Create the account (and key if requested).
	createdServiceAccount, createdKey, err := serviceAccount(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create service account")
	}

	// Attach IAM roles at project/org scopes.
	if err := iam(ctx, locals, createdServiceAccount, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create IAM bindings")
	}

	// === Export stack outputs ===
	ctx.Export(OpEmail, createdServiceAccount.Email)

	if createdKey != nil {
		ctx.Export(OpKeyBase64, createdKey.PrivateKey)
	}

	return nil
}
