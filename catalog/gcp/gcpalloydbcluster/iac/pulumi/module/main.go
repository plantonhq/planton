package module

import (
	"github.com/pkg/errors"
	gcpalloydbclusterv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpalloydbcluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpalloydbclusterv1alpha1.GcpAlloydbClusterIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	createdCluster, err := cluster(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create alloydb cluster")
	}

	if err := primaryInstance(ctx, locals, gcpProvider, createdCluster); err != nil {
		return errors.Wrap(err, "failed to create alloydb primary instance")
	}

	return nil
}
