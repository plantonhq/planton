package module

import (
	"github.com/pkg/errors"
	gcpcloudcomposeruserworkloadsconfigmapv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudcomposeruserworkloadsconfigmap/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpcloudcomposeruserworkloadsconfigmapv1alpha1.GcpCloudComposerUserWorkloadsConfigMapIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := userWorkloadsConfigMap(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create composer user workloads config map")
	}

	return nil
}
