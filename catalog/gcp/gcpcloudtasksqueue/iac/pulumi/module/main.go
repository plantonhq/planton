package module

import (
	"github.com/pkg/errors"
	gcpcloudtasksqueuev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudtasksqueue/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpcloudtasksqueuev1alpha1.GcpCloudTasksQueueIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := cloudTasksQueue(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create cloud tasks queue")
	}

	return nil
}
