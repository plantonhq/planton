package module

import (
	"github.com/pkg/errors"
	gcpalloydbclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpalloydbcluster/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpalloydbclusterv1.GcpAlloydbClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
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
