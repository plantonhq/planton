package module

import (
	"github.com/pkg/errors"
	gcpserviceaccountv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpserviceaccount/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the Service Account, optional key, IAM bindings, and exports outputs.
func Resources(
	ctx *pulumi.Context,
	stackInput *gcpserviceaccountv1alpha1.GcpServiceAccountStackInput,
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
	// The ready-made IAM member string ("serviceAccount:<email>") — downstream
	// grants reference this directly, so no consumer ever assembles the prefix.
	ctx.Export(OpMember, createdServiceAccount.Member)
	ctx.Export(OpUniqueId, createdServiceAccount.UniqueId)
	ctx.Export(OpName, createdServiceAccount.Name)

	if createdKey != nil {
		ctx.Export(OpKeyBase64, pulumi.ToSecret(createdKey.PrivateKey))
	}

	return nil
}
