package module

import (
	"github.com/pkg/errors"
	gcpplantonrunnerv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpplantonrunner/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the runner appliance. The pieces, in dependency
// order: the runtime service account (the seam keyless cloud access
// rides), the token secret plus the accessor grant scoped to exactly that
// account, and finally the Cloud Run service that runs the container
// pinned to exactly one always-on instance.
//
// ENROLLMENT IS TOKEN-FIRST: the service ships the runner TOKEN (via
// Secret Manager), never an identity. The runner joins the control plane
// on first boot, registers itself under RunnerName, and receives its own
// individually revocable identity; instance replacement re-joins with the
// same token (its lineage re-admits the runner it originally admitted).
func Resources(ctx *pulumi.Context, iacInput *gcpplantonrunnerv1alpha1.GcpPlantonRunnerIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the GCP provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (service-account key,
	// keyless web identity, or ambient chain).
	provider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to create GCP provider")
	}

	serviceAccountEmail, createdServiceAccount, err := runtimeServiceAccount(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to resolve runtime service account")
	}

	createdSecretVersion, createdAccessorGrant, err := tokenSecret(ctx, locals, provider, serviceAccountEmail)
	if err != nil {
		return errors.Wrap(err, "failed to create token secret")
	}

	if err := runnerService(ctx, locals, provider, serviceAccountEmail, createdServiceAccount, createdSecretVersion, createdAccessorGrant); err != nil {
		return errors.Wrap(err, "failed to create runner service")
	}

	return nil
}
