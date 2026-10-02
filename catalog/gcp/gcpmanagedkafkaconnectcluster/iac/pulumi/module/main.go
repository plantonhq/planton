package module

import (
	"github.com/pkg/errors"
	gcpmanagedkafkaconnectclusterv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmanagedkafkaconnectcluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpmanagedkafkaconnectclusterv1alpha1.GcpManagedKafkaConnectClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := connectCluster(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create managed kafka connect cluster")
	}

	return nil
}
