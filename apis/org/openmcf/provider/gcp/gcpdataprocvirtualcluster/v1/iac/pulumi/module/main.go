package module

import (
	gcpdataprocvirtualclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpdataprocvirtualcluster/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type GcpDataprocVirtualClusterStackInput = gcpdataprocvirtualclusterv1.GcpDataprocVirtualClusterStackInput

// Resources provisions a Dataproc on GKE virtual cluster and exports its outputs.
func Resources(ctx *pulumi.Context, stackInput *GcpDataprocVirtualClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to get gcp provider")
	}

	if err := dataprocVirtualCluster(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create dataproc virtual cluster")
	}

	return nil
}
