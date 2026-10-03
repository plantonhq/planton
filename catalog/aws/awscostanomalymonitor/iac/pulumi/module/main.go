package module

import (
	"github.com/pkg/errors"
	awscostanomalymonitorv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscostanomalymonitor/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the anomaly monitor and its
// folded alert subscriptions, and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awscostanomalymonitorv1alpha1.AwsCostAnomalyMonitorIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := anomalyMonitor(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "anomaly monitor")
	}

	return nil
}
