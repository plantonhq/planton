package module

import (
	"github.com/pkg/errors"
	gcpgkeworkloadidentitybindingv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgkeworkloadidentitybinding/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi entry‑point invoked by Planton’s CLI.
// It mirrors a Terraform module’s main.tf and keeps the control‑flow flat.
func Resources(ctx *pulumi.Context,
	iacInput *gcpgkeworkloadidentitybindingv1alpha1.GcpGkeWorkloadIdentityBindingIacInput) error {

	locals := initializeLocals(ctx, iacInput)

	// Set up the GCP provider from the supplied credential spec.
	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to set up google provider")
	}

	// Create the IAM binding that enables Workload Identity impersonation.
	if err = workloadIdentityBinding(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create workload identity binding")
	}

	return nil
}
