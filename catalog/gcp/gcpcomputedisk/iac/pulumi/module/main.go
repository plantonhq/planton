package module

import (
	"github.com/pkg/errors"
	gcpcomputediskv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcomputedisk/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry-point for the GcpComputeDisk
// component.
func Resources(ctx *pulumi.Context, iacInput *gcpcomputediskv1alpha1.GcpComputeDiskIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if _, err := computeDisk(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create compute disk")
	}

	return nil
}
