package module

import (
	"github.com/pkg/errors"
	gcpdataprocclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpdataproccluster/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpdataprocclusterv1.GcpDataprocClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := dataprocCluster(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create dataproc cluster")
	}

	return nil
}
