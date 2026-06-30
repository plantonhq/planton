package module

import (
	"github.com/pkg/errors"
	gcpgkeworkloadidentitybindingv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpgkeworkloadidentitybinding/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi entry‑point invoked by Planton’s CLI.
// It mirrors a Terraform module’s main.tf and keeps the control‑flow flat.
func Resources(ctx *pulumi.Context,
	stackInput *gcpgkeworkloadidentitybindingv1.GcpGkeWorkloadIdentityBindingStackInput) error {

	locals := initializeLocals(ctx, stackInput)

	// Set up the GCP provider from the supplied credential spec.
	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to set up google provider")
	}

	// Create the IAM binding that enables Workload Identity impersonation.
	if err = workloadIdentityBinding(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create workload identity binding")
	}

	return nil
}
