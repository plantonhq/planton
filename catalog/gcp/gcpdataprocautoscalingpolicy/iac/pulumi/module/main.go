package module

import (
	"github.com/pkg/errors"
	gcpdataprocautoscalingpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdataprocautoscalingpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpdataprocautoscalingpolicyv1alpha1.GcpDataprocAutoscalingPolicyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := autoscalingPolicy(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create dataproc autoscaling policy")
	}

	return nil
}
