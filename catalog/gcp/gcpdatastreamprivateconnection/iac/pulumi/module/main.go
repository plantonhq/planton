package module

import (
	"github.com/pkg/errors"
	gcpdatastreamprivateconnectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdatastreamprivateconnection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpdatastreamprivateconnectionv1alpha1.GcpDatastreamPrivateConnectionStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := privateConnection(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create datastream private connection")
	}

	return nil
}
