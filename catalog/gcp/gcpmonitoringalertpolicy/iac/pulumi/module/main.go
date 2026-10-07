package module

import (
	"github.com/pkg/errors"
	gcpmonitoringalertpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmonitoringalertpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpmonitoringalertpolicyv1alpha1.GcpMonitoringAlertPolicyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := alertPolicy(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create alert policy")
	}

	return nil
}
